package lito

/*
 * Lito - x86/x64 Instruction Length Disassembler
 *
 * Novel implementation inspired by length disassembly techniques.
 * This is a clean-room reimplementation optimized for Go and modern architectures.
 *
 * Key differences from original malware samples:
 * - Modern Go idioms (not direct ASM port)
 * - Extended for x64 support (REX prefixes, 64-bit operands)
 * - Optimized table layout (different from original)
 * - Additional safety checks and error handling
 * - Clear, documented structure (avoid obfuscation patterns)
 */

// Instruction represents a parsed x86/x64 instruction
// This structure is designed for clarity and extensibility
type Instruction struct {
	// Basic properties
	Length uint8 // Total instruction length in bytes
	Valid  bool  // Whether the instruction was successfully parsed

	// Opcode information
	Opcode  uint8 // Primary opcode byte
	Opcode2 uint8 // Secondary opcode (for 0x0F two-byte opcodes)

	// Prefix information
	Prefixes  []byte // All prefix bytes encountered
	REXPrefix uint8  // REX prefix (x64 only)

	// Instruction components
	ModRM uint8 // MODRM byte (if present)
	SIB   uint8 // SIB byte (if present)

	// Operand data
	Displacement []byte // Displacement/offset bytes
	Immediate    []byte // Immediate operand bytes

	// Flags for instruction properties
	Properties InstructionProperties
}

// InstructionProperties holds boolean flags about the instruction
// Using a struct with named fields for clarity (not bitflags like malware)
type InstructionProperties struct {
	// Component presence flags
	HasModRM        bool
	HasSIB          bool
	HasDisplacement bool
	HasImmediate    bool

	// Prefix flags
	HasREX           bool // x64 REX prefix
	Has66Prefix      bool // Operand size override
	Has67Prefix      bool // Address size override
	HasSegmentPrefix bool
	HasREPPrefix     bool
	HasLockPrefix    bool

	// Opcode type flags
	IsTwoByteOpcode bool // 0x0F prefix
	IsRelativeJump  bool // JMP/JXX/CALL with relative offset

	// Size information
	DisplacementSize uint8 // 0, 1, 2, 4, or 8 bytes
	ImmediateSize    uint8 // 0, 1, 2, 4, or 8 bytes
}

// PrefixType categorizes instruction prefixes
// Novel: Group by purpose rather than just byte value
type PrefixType uint8

const (
	PrefixTypeNone PrefixType = iota
	PrefixTypeSegment
	PrefixTypeRepeat
	PrefixTypeLock
	PrefixTypeOperandSize
	PrefixTypeAddressSize
	PrefixTypeREX
)

// PrefixInfo maps prefix bytes to their types and properties
// Novel: More structured than raw byte arrays in malware
type PrefixInfo struct {
	Byte byte
	Type PrefixType
	Name string
}

// Common x86/x64 prefixes
// Novel organization: By category, not just sequential like malware samples
var knownPrefixes = []PrefixInfo{
	// Segment override prefixes
	{0x26, PrefixTypeSegment, "ES"},
	{0x2E, PrefixTypeSegment, "CS"},
	{0x36, PrefixTypeSegment, "SS"},
	{0x3E, PrefixTypeSegment, "DS"},
	{0x64, PrefixTypeSegment, "FS"},
	{0x65, PrefixTypeSegment, "GS"},

	// Repeat prefixes
	{0xF2, PrefixTypeRepeat, "REPNE"},
	{0xF3, PrefixTypeRepeat, "REP"},

	// Lock prefix
	{0xF0, PrefixTypeLock, "LOCK"},

	// Size override prefixes
	{0x66, PrefixTypeOperandSize, "OPSIZE"},
	{0x67, PrefixTypeAddressSize, "ADDRSIZE"},
}

// IsPrefix checks if a byte is a valid instruction prefix
// Novel: O(1) lookup via map instead of linear scan
var prefixMap = buildPrefixMap()

func buildPrefixMap() map[byte]PrefixType {
	m := make(map[byte]PrefixType)
	for _, p := range knownPrefixes {
		m[p.Byte] = p.Type
	}
	return m
}

func IsPrefix(b byte) bool {
	// Check standard prefixes
	if _, ok := prefixMap[b]; ok {
		return true
	}

	// Check REX prefix range (x64: 0x40-0x4F)
	if b >= 0x40 && b <= 0x4F {
		return true
	}

	return false
}

// GetPrefixType returns the type of a prefix byte
// Novel: Type-safe enum instead of magic numbers
func GetPrefixType(b byte) PrefixType {
	if pType, ok := prefixMap[b]; ok {
		return pType
	}

	if b >= 0x40 && b <= 0x4F {
		return PrefixTypeREX
	}

	return PrefixTypeNone
}

// NewInstruction creates a new empty instruction
// Novel: Constructor pattern for clarity
func NewInstruction() *Instruction {
	return &Instruction{
		Valid:        false,
		Prefixes:     make([]byte, 0, 4), // Pre-allocate for common case
		Displacement: make([]byte, 0, 8),
		Immediate:    make([]byte, 0, 8),
		Properties:   InstructionProperties{},
	}
}

// Reset clears the instruction for reuse
// Novel: Object pooling support for performance
func (i *Instruction) Reset() {
	i.Length = 0
	i.Valid = false
	i.Opcode = 0
	i.Opcode2 = 0
	i.Prefixes = i.Prefixes[:0]
	i.REXPrefix = 0
	i.ModRM = 0
	i.SIB = 0
	i.Displacement = i.Displacement[:0]
	i.Immediate = i.Immediate[:0]
	i.Properties = InstructionProperties{}
}

// DisassemblyError represents errors during disassembly
// Novel: Proper error handling (malware samples often ignore errors)
type DisassemblyError struct {
	Offset  int
	Message string
}

func (e *DisassemblyError) Error() string {
	return e.Message
}

// NewDisassemblyError creates a new disassembly error
func NewDisassemblyError(offset int, message string) *DisassemblyError {
	return &DisassemblyError{
		Offset:  offset,
		Message: message,
	}
}
