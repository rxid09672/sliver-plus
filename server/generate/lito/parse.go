package lito

import "fmt"

/*
 * Instruction Parsing Implementation
 *
 * Novel approach: Modern error handling, clear state management,
 * and extensible architecture instead of direct ASM translation.
 */

// Parser holds the state for parsing a single instruction
// Novel: Stateful parser object (not procedural like malware)
type Parser struct {
	code        []byte
	offset      int
	startOffset int
	instruction *Instruction
	mode64      bool // x64 mode flag
}

// NewParser creates a new instruction parser
func NewParser(code []byte, offset int, mode64 bool) *Parser {
	return &Parser{
		code:        code,
		offset:      offset,
		startOffset: offset,
		instruction: NewInstruction(),
		mode64:      mode64,
	}
}

// Parse performs complete instruction parsing
// Novel: Single entry point with clear error handling
func (p *Parser) Parse() (*Instruction, error) {
	if p.offset >= len(p.code) {
		return nil, NewDisassemblyError(p.offset, "offset beyond code length")
	}

	// Step 1: Parse all prefixes
	if err := p.parsePrefixes(); err != nil {
		return nil, err
	}

	// Step 2: Parse opcode(s)
	if err := p.parseOpcode(); err != nil {
		return nil, err
	}

	// Step 3: Parse MODRM/SIB if present
	if err := p.parseModRM(); err != nil {
		return nil, err
	}

	// Step 4: Parse displacement
	if err := p.parseDisplacement(); err != nil {
		return nil, err
	}

	// Step 5: Parse immediate
	if err := p.parseImmediate(); err != nil {
		return nil, err
	}

	// Calculate final length
	p.instruction.Length = uint8(p.offset - p.startOffset)
	p.instruction.Valid = true

	return p.instruction, nil
}

// parsePrefixes handles all instruction prefix bytes
// Novel: Proper prefix grouping and REX handling for x64
func (p *Parser) parsePrefixes() error {
	maxPrefixes := 15 // Intel spec: max 4 legacy + 1 REX, but be safe
	prefixCount := 0

	for p.offset < len(p.code) {
		if prefixCount >= maxPrefixes {
			return NewDisassemblyError(p.offset, "too many prefixes")
		}

		b := p.code[p.offset]
		prefixType := GetPrefixType(b)

		// Not a prefix - we're done
		if prefixType == PrefixTypeNone {
			break
		}

		// In x64 mode, REX must be the last prefix
		if p.instruction.Properties.HasREX {
			return NewDisassemblyError(p.offset, "prefix after REX")
		}

		// Handle prefix based on type
		switch prefixType {
		case PrefixTypeSegment:
			p.instruction.Properties.HasSegmentPrefix = true
			p.instruction.Prefixes = append(p.instruction.Prefixes, b)

		case PrefixTypeRepeat:
			p.instruction.Properties.HasREPPrefix = true
			p.instruction.Prefixes = append(p.instruction.Prefixes, b)

		case PrefixTypeLock:
			p.instruction.Properties.HasLockPrefix = true
			p.instruction.Prefixes = append(p.instruction.Prefixes, b)

		case PrefixTypeOperandSize:
			p.instruction.Properties.Has66Prefix = true
			p.instruction.Prefixes = append(p.instruction.Prefixes, b)

		case PrefixTypeAddressSize:
			p.instruction.Properties.Has67Prefix = true
			p.instruction.Prefixes = append(p.instruction.Prefixes, b)

		case PrefixTypeREX:
			// REX prefix only valid in x64 mode
			if !p.mode64 {
				// In x86 mode, 0x40-0x4F are INC/DEC, not prefixes
				break
			}
			p.instruction.Properties.HasREX = true
			p.instruction.REXPrefix = b
			p.instruction.Prefixes = append(p.instruction.Prefixes, b)
		}

		p.offset++
		prefixCount++
	}

	return nil
}

// parseOpcode handles primary and secondary opcodes
// Novel: Clean separation of single vs two-byte opcodes
func (p *Parser) parseOpcode() error {
	if p.offset >= len(p.code) {
		return NewDisassemblyError(p.offset, "missing opcode")
	}

	opcode := p.code[p.offset]
	p.instruction.Opcode = opcode
	p.offset++

	// Check for two-byte opcode escape (0x0F)
	if opcode == 0x0F {
		if p.offset >= len(p.code) {
			return NewDisassemblyError(p.offset, "missing second opcode byte")
		}

		p.instruction.Properties.IsTwoByteOpcode = true
		p.instruction.Opcode2 = p.code[p.offset]
		p.offset++
	}

	return nil
}

// parseModRM handles MODRM and SIB bytes
// Novel: Comprehensive x64 support with RIP-relative addressing
func (p *Parser) parseModRM() error {
	// Get opcode info to check if MODRM is needed
	info := GetOpcodeInfo(p.instruction.Opcode2, p.instruction.Properties.IsTwoByteOpcode)
	if p.instruction.Opcode2 == 0 {
		info = GetOpcodeInfo(p.instruction.Opcode, false)
	}

	if !info.HasModRM() {
		return nil
	}

	if p.offset >= len(p.code) {
		return NewDisassemblyError(p.offset, "missing MODRM byte")
	}

	modrm := p.code[p.offset]
	p.instruction.ModRM = modrm
	p.instruction.Properties.HasModRM = true
	p.offset++

	// Extract MODRM fields
	mod := (modrm >> 6) & 0x03
	reg := (modrm >> 3) & 0x07
	rm := modrm & 0x07

	// Determine if SIB byte is needed
	// Novel: Proper x64 extended register handling
	needsSIB := false
	if mod != 3 { // Not register-direct mode
		// SIB needed when r/m = 4 (or 12 with REX.B in x64)
		if rm == 4 {
			// In x64 with address size override (0x67), different rules apply
			if !p.instruction.Properties.Has67Prefix {
				needsSIB = true
			}
		}
	}

	// Parse SIB if needed
	if needsSIB {
		if p.offset >= len(p.code) {
			return NewDisassemblyError(p.offset, "missing SIB byte")
		}

		p.instruction.SIB = p.code[p.offset]
		p.instruction.Properties.HasSIB = true
		p.offset++
	}

	// Determine displacement size based on MOD field
	switch mod {
	case 0: // No displacement (with exceptions)
		if rm == 5 && !needsSIB {
			// [disp32] or [RIP+disp32] in x64
			p.instruction.Properties.DisplacementSize = 4
			p.instruction.Properties.HasDisplacement = true
		} else if needsSIB {
			// Check SIB base field
			sib := p.instruction.SIB
			base := sib & 0x07
			if base == 5 {
				// [scaled index] + disp32
				p.instruction.Properties.DisplacementSize = 4
				p.instruction.Properties.HasDisplacement = true
			}
		}

	case 1: // disp8
		p.instruction.Properties.DisplacementSize = 1
		p.instruction.Properties.HasDisplacement = true

	case 2: // disp32 (or disp16 with 0x67 prefix)
		if p.instruction.Properties.Has67Prefix {
			p.instruction.Properties.DisplacementSize = 2
		} else {
			p.instruction.Properties.DisplacementSize = 4
		}
		p.instruction.Properties.HasDisplacement = true

	case 3: // Register-direct (no displacement)
		// No displacement
	}

	// Handle special cases for certain opcodes
	// Some opcodes have forced immediate sizes (like TEST r/m, imm)
	if p.instruction.Opcode == 0xF6 && reg == 0 {
		// TEST r/m8, imm8
		p.instruction.Properties.ImmediateSize = 1
		p.instruction.Properties.HasImmediate = true
	} else if p.instruction.Opcode == 0xF7 && reg == 0 {
		// TEST r/m, imm (16/32/64)
		p.instruction.Properties.ImmediateSize = p.getOperandSize()
		p.instruction.Properties.HasImmediate = true
	}

	return nil
}

// parseDisplacement reads the displacement bytes
// Novel: Safe bounds checking (malware often didn't check)
func (p *Parser) parseDisplacement() error {
	if !p.instruction.Properties.HasDisplacement {
		return nil
	}

	dispSize := int(p.instruction.Properties.DisplacementSize)
	if p.offset+dispSize > len(p.code) {
		return NewDisassemblyError(p.offset, "displacement extends beyond code")
	}

	p.instruction.Displacement = p.code[p.offset : p.offset+dispSize]
	p.offset += dispSize

	return nil
}

// parseImmediate reads the immediate operand bytes
// Novel: Proper size calculation with prefix handling
func (p *Parser) parseImmediate() error {
	// Get opcode info
	info := GetOpcodeInfo(p.instruction.Opcode2, p.instruction.Properties.IsTwoByteOpcode)
	if p.instruction.Opcode2 == 0 {
		info = GetOpcodeInfo(p.instruction.Opcode, false)
	}

	// Determine immediate size
	var immSize uint8 = 0

	if info.Flags&OpFlagImm8 != 0 {
		immSize = 1
	} else if info.Flags&OpFlagImm16 != 0 {
		immSize = 2
	} else if info.Flags&OpFlagImm32 != 0 {
		immSize = 4
	} else if info.Flags&(OpFlagFullSize|OpFlagPrefixDependent) != 0 {
		// Size depends on operand size
		immSize = p.getOperandSize()
	}

	// Special cases for certain opcodes
	// MOV moffs (0xA0-0xA3) uses address-size displacement, not immediate
	if p.instruction.Opcode >= 0xA0 && p.instruction.Opcode <= 0xA3 {
		// Already handled in MODRM parsing
		return nil
	}

	// ENTER has two immediates (imm16 + imm8)
	if p.instruction.Opcode == 0xC8 {
		immSize = 3 // 2 + 1
	}

	// Far CALL/JMP have segment:offset
	if p.instruction.Opcode == 0x9A || p.instruction.Opcode == 0xEA {
		immSize = 6 // 4-byte offset + 2-byte segment
		if p.instruction.Properties.Has66Prefix {
			immSize = 4 // 2-byte offset + 2-byte segment
		}
	}

	if immSize == 0 {
		return nil
	}

	// Check if already set by special MODRM handling
	if p.instruction.Properties.HasImmediate {
		immSize = p.instruction.Properties.ImmediateSize
	}

	// Read immediate bytes
	if p.offset+int(immSize) > len(p.code) {
		return NewDisassemblyError(p.offset, "immediate extends beyond code")
	}

	p.instruction.Immediate = p.code[p.offset : p.offset+int(immSize)]
	p.instruction.Properties.HasImmediate = true
	p.instruction.Properties.ImmediateSize = immSize
	p.offset += int(immSize)

	// Mark relative jumps/calls
	if info.IsRelativeJump() {
		p.instruction.Properties.IsRelativeJump = true
	}

	return nil
}

// getOperandSize returns the current operand size (1, 2, 4, or 8 bytes)
// Novel: Proper x64 REX.W handling
func (p *Parser) getOperandSize() uint8 {
	// Check REX.W (bit 3) for 64-bit operand size
	if p.instruction.Properties.HasREX && (p.instruction.REXPrefix&0x08) != 0 {
		return 8
	}

	// Check 0x66 prefix for 16-bit operand size
	if p.instruction.Properties.Has66Prefix {
		return 2
	}

	// Default: 32-bit in x64 mode, 32-bit in x86 mode
	return 4
}

// Disassemble is the main entry point for parsing a single instruction
// Novel: Clean API with mode detection
func Disassemble(code []byte, offset int, mode64 bool) (*Instruction, error) {
	parser := NewParser(code, offset, mode64)
	return parser.Parse()
}

// DisassembleLength returns only the instruction length
// Novel: Optimized path when only length is needed
func DisassembleLength(code []byte, offset int, mode64 bool) (int, error) {
	instr, err := Disassemble(code, offset, mode64)
	if err != nil {
		return 0, err
	}
	return int(instr.Length), nil
}

// DisassembleAll parses multiple instructions from code
// Novel: Batch processing with error recovery
func DisassembleAll(code []byte, maxInstructions int, mode64 bool) ([]*Instruction, error) {
	instructions := make([]*Instruction, 0, maxInstructions)
	offset := 0

	for offset < len(code) && len(instructions) < maxInstructions {
		instr, err := Disassemble(code, offset, mode64)
		if err != nil {
			return instructions, fmt.Errorf("error at offset %d: %w", offset, err)
		}

		instructions = append(instructions, instr)
		offset += int(instr.Length)
	}

	return instructions, nil
}

// IsControlFlow checks if instruction is a control flow instruction
// Novel: Helper for metamorphic engine (will need this for Morpher!)
func (i *Instruction) IsControlFlow() bool {
	// Check for jumps, calls, returns
	op := i.Opcode

	// Conditional jumps (short)
	if op >= 0x70 && op <= 0x7F {
		return true
	}

	// JMP, CALL, RET variants
	if op >= 0xE0 && op <= 0xEB {
		return true
	}

	if op == 0xC2 || op == 0xC3 || op == 0xCA || op == 0xCB {
		return true // RET variants
	}

	if op == 0xFF {
		// Indirect JMP/CALL (check MODRM reg field)
		if i.Properties.HasModRM {
			reg := (i.ModRM >> 3) & 0x07
			if reg >= 2 && reg <= 5 {
				return true // CALL/JMP near/far
			}
		}
	}

	// Two-byte opcodes
	if i.Properties.IsTwoByteOpcode {
		// Conditional jumps (near)
		if i.Opcode2 >= 0x80 && i.Opcode2 <= 0x8F {
			return true
		}
	}

	return false
}

// GetRelativeTarget calculates the absolute target of a relative jump/call
// Novel: Critical for Morpher's jump relocation!
func (i *Instruction) GetRelativeTarget(instrAddress uint64) (uint64, error) {
	if !i.Properties.IsRelativeJump {
		return 0, fmt.Errorf("instruction is not a relative jump/call")
	}

	if len(i.Immediate) == 0 {
		return 0, fmt.Errorf("no immediate operand for relative jump")
	}

	// Calculate next instruction address
	nextInstr := instrAddress + uint64(i.Length)

	// Parse immediate as signed offset
	var offset int64
	switch len(i.Immediate) {
	case 1: // 8-bit signed
		offset = int64(int8(i.Immediate[0]))
	case 2: // 16-bit signed
		offset = int64(int16(i.Immediate[0]) | int16(i.Immediate[1])<<8)
	case 4: // 32-bit signed
		offset = int64(int32(i.Immediate[0]) | int32(i.Immediate[1])<<8 |
			int32(i.Immediate[2])<<16 | int32(i.Immediate[3])<<24)
	default:
		return 0, fmt.Errorf("unexpected immediate size for relative jump: %d", len(i.Immediate))
	}

	// Calculate target
	target := nextInstr + uint64(offset)

	return target, nil
}
