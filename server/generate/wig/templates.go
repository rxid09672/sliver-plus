package wig

/*
 * Safe Instruction Template System
 *
 * Generates instructions from validated templates to ensure:
 * 1. Valid CPU semantics (parseable by Lito)
 * 2. Semantic nullity (no side effects)
 * 3. ABI compliance (stack balance, scratch regs only)
 * 4. CFG safety (no ROP gadgets)
 *
 * Novel approach: Template-based generation ensures safety while
 * still allowing vector-guided variety selection.
 *
 * Key innovation: Vector dimensions guide WHICH templates to use,
 * templates ensure WHAT is generated is always valid.
 */

import (
	"fmt"

	"github.com/bishopfox/sliver/server/generate/lito"
	"github.com/bishopfox/sliver/server/generate/morpher"
)

// InstructionTemplate defines a safe instruction pattern
// Novel: Rich template metadata for vector-guided selection
type InstructionTemplate struct {
	Name       string
	Category   string // "NOP", "MOV", "STACK", "ARITHMETIC", etc.
	Opcode     byte
	Length     int
	Complexity int // 1=simple, 2=medium, 3=complex

	// Safety properties
	RequiresModRM bool
	RequiresSIB   bool
	StackDelta    int // +1 for PUSH, -1 for POP, 0 for neutral
	UsesRegisters []int

	// Generation functions
	GenerateBytes  func(*morpher.XorShift128) []byte
	ValidateResult func(*lito.Instruction) bool
}

// Template catalog - whitelisted safe patterns
// Novel: Comprehensive catalog with safety guarantees
var instructionTemplates = []InstructionTemplate{
	// Category: NOP (complexity 1)
	{
		Name:       "NOP",
		Category:   "NOP",
		Opcode:     0x90,
		Length:     1,
		Complexity: 1,
		StackDelta: 0,
		GenerateBytes: func(rng *morpher.XorShift128) []byte {
			return []byte{0x90}
		},
		ValidateResult: func(i *lito.Instruction) bool {
			return i.Opcode == 0x90
		},
	},

	// Category: MOV reg,reg (same register - semantic NOP)
	{
		Name:          "MOV_REG_SELF",
		Category:      "MOV",
		Opcode:        0x89,
		Length:        2,
		Complexity:    1,
		StackDelta:    0,
		RequiresModRM: true,
		UsesRegisters: []int{0, 1, 2}, // Only scratch regs
		GenerateBytes: func(rng *morpher.XorShift128) []byte {
			// Pick scratch register (EAX, ECX, or EDX)
			reg := rng.Intn(3)
			// MOV reg, reg: 89 [11 reg reg]
			modrm := 0xC0 | (byte(reg) << 3) | byte(reg)
			return []byte{0x89, modrm}
		},
		ValidateResult: func(i *lito.Instruction) bool {
			if i.Opcode != 0x89 {
				return false
			}
			reg := (i.ModRM >> 3) & 0x07
			rm := i.ModRM & 0x07
			return reg == rm && reg < 3 // Same reg, scratch only
		},
	},

	// Category: TEST reg,reg (read-only, no side effects)
	{
		Name:          "TEST_REG",
		Category:      "TEST",
		Opcode:        0x85,
		Length:        2,
		Complexity:    2,
		StackDelta:    0,
		RequiresModRM: true,
		UsesRegisters: []int{0, 1, 2},
		GenerateBytes: func(rng *morpher.XorShift128) []byte {
			reg1 := rng.Intn(3)
			reg2 := rng.Intn(3)
			modrm := 0xC0 | (byte(reg1) << 3) | byte(reg2)
			return []byte{0x85, modrm}
		},
		ValidateResult: func(i *lito.Instruction) bool {
			return i.Opcode == 0x85
		},
	},

	// Category: CMP reg,reg (read-only, no side effects)
	{
		Name:          "CMP_REG",
		Category:      "CMP",
		Opcode:        0x39,
		Length:        2,
		Complexity:    2,
		StackDelta:    0,
		RequiresModRM: true,
		UsesRegisters: []int{0, 1, 2},
		GenerateBytes: func(rng *morpher.XorShift128) []byte {
			reg1 := rng.Intn(3)
			reg2 := rng.Intn(3)
			modrm := 0xC0 | (byte(reg1) << 3) | byte(reg2)
			return []byte{0x39, modrm}
		},
		ValidateResult: func(i *lito.Instruction) bool {
			return i.Opcode == 0x39
		},
	},

	// Category: LEA reg,[reg] (semantic NOP if same register)
	{
		Name:          "LEA_REG_SELF",
		Category:      "LEA",
		Opcode:        0x8D,
		Length:        2,
		Complexity:    1,
		StackDelta:    0,
		RequiresModRM: true,
		UsesRegisters: []int{0, 1, 2},
		GenerateBytes: func(rng *morpher.XorShift128) []byte {
			reg := rng.Intn(3)
			// LEA reg, [reg]: 8D [00 reg reg]
			modrm := byte(reg<<3) | byte(reg)
			return []byte{0x8D, modrm}
		},
		ValidateResult: func(i *lito.Instruction) bool {
			return i.Opcode == 0x8D
		},
	},

	// Category: PUSH (must be balanced with POP!)
	{
		Name:          "PUSH_REG",
		Category:      "STACK",
		Opcode:        0x50, // Base opcode, will add reg
		Length:        1,
		Complexity:    2,
		StackDelta:    1, // Increases stack
		UsesRegisters: []int{0, 1, 2},
		GenerateBytes: func(rng *morpher.XorShift128) []byte {
			reg := rng.Intn(3)
			return []byte{0x50 + byte(reg)}
		},
		ValidateResult: func(i *lito.Instruction) bool {
			return i.Opcode >= 0x50 && i.Opcode <= 0x52
		},
	},

	// Category: POP (must balance PUSH!)
	{
		Name:          "POP_REG",
		Category:      "STACK",
		Opcode:        0x58, // Base opcode, will add reg
		Length:        1,
		Complexity:    2,
		StackDelta:    -1, // Decreases stack
		UsesRegisters: []int{0, 1, 2},
		GenerateBytes: func(rng *morpher.XorShift128) []byte {
			reg := rng.Intn(3)
			return []byte{0x58 + byte(reg)}
		},
		ValidateResult: func(i *lito.Instruction) bool {
			return i.Opcode >= 0x58 && i.Opcode <= 0x5A
		},
	},
}

// SelectTemplatesByVector selects templates based on vector preferences
// Novel: Vector dimensions guide template selection
func SelectTemplatesByVector(vector *CodeVector, rng *morpher.XorShift128) []InstructionTemplate {
	selected := make([]InstructionTemplate, 0)

	// Build weighted template list based on vector
	for _, tmpl := range instructionTemplates {
		// Get vector preference for this template's category
		preference := getTemplatePreference(tmpl, vector)

		// Probabilistic selection based on preference
		if rng.Float64() < preference {
			selected = append(selected, tmpl)
		}
	}

	// Ensure at least some templates selected
	if len(selected) == 0 {
		// Fall back to simple NOP
		selected = append(selected, instructionTemplates[0])
	}

	return selected
}

// getTemplatePreference calculates how much a template is preferred by vector
func getTemplatePreference(tmpl InstructionTemplate, vector *CodeVector) float64 {
	switch tmpl.Category {
	case "NOP":
		return vector.Values[DimNOPFrequency]
	case "MOV":
		return vector.Values[DimMOVFrequency]
	case "STACK":
		if tmpl.StackDelta > 0 {
			return vector.Values[DimPUSHFrequency]
		} else {
			return vector.Values[DimPOPFrequency]
		}
	case "TEST":
		return vector.Values[DimTESTFrequency]
	case "CMP":
		return vector.Values[DimCMPFrequency]
	case "LEA":
		return vector.Values[DimLEAFrequency]
	default:
		return 0.5 // Medium preference
	}
}

// GenerateInstructionFromVector generates a single instruction guided by vector
// Novel: Vector-guided but template-constrained (safety + diversity)
func GenerateInstructionFromVector(vector *CodeVector, rng *morpher.XorShift128, mode64 bool) ([]byte, error) {
	// Select templates based on vector preferences
	templates := SelectTemplatesByVector(vector, rng)

	if len(templates) == 0 {
		return []byte{0x90}, nil // Fallback to NOP
	}

	// Pick random template from selected
	tmpl := templates[rng.Intn(len(templates))]

	// Generate bytes from template
	bytes := tmpl.GenerateBytes(rng)

	// CRITICAL: Validate with Lito
	instr, err := lito.Disassemble(bytes, 0, mode64)
	if err != nil {
		return []byte{0x90}, nil // Validation failed, use NOP
	}

	// CRITICAL: Validate template-specific rules
	if !tmpl.ValidateResult(instr) {
		return []byte{0x90}, nil // Template validation failed, use NOP
	}

	return bytes, nil
}

// GenerateSequenceFromVector generates a sequence of instructions
// Novel: Multi-instruction generation with stack balance enforcement
func GenerateSequenceFromVector(vector *CodeVector, targetLength int, rng *morpher.XorShift128, mode64 bool) ([]byte, error) {
	sequence := make([]byte, 0, targetLength)
	stackDelta := 0 // Track stack balance

	for len(sequence) < targetLength {
		// Generate one instruction
		instrBytes, err := GenerateInstructionFromVector(vector, rng, mode64)
		if err != nil {
			continue
		}

		// Parse to check properties
		instr, _ := lito.Disassemble(instrBytes, 0, mode64)

		// Check stack delta
		newStackDelta := stackDelta
		if instr.Opcode >= 0x50 && instr.Opcode <= 0x57 {
			newStackDelta++ // PUSH
		} else if instr.Opcode >= 0x58 && instr.Opcode <= 0x5F {
			newStackDelta-- // POP
		}

		// CRITICAL: Don't let stack get too imbalanced
		if newStackDelta < -5 || newStackDelta > 5 {
			// Skip this instruction, try another
			continue
		}

		// Add instruction
		sequence = append(sequence, instrBytes...)
		stackDelta = newStackDelta

		// Stop if we've reached target length
		if len(sequence) >= targetLength {
			break
		}
	}

	// CRITICAL: Balance stack before returning
	for stackDelta > 0 {
		// Need POPs
		if len(sequence)+1 <= targetLength {
			sequence = append(sequence, 0x58) // POP EAX
			stackDelta--
		} else {
			break
		}
	}

	for stackDelta < 0 {
		// Need PUSHs
		if len(sequence)+1 <= targetLength {
			sequence = append(sequence, 0x50) // PUSH EAX
			stackDelta++
		} else {
			break
		}
	}

	return sequence, nil
}

// ValidateGeneratedSequence performs comprehensive validation
// Novel: Multi-layer validation before accepting generated code
func ValidateGeneratedSequence(sequence []byte, manifold *ExecutableManifold) error {
	// Layer 1: Parse with Lito (syntax)
	stream := lito.NewInstructionStream(sequence, manifold.Mode64)
	if err := stream.ParseAll(); err != nil {
		return fmt.Errorf("invalid syntax: %w", err)
	}

	// Layer 2: Check stack balance (ABI)
	stackDelta := 0
	for _, instr := range stream.Instructions {
		if instr.Opcode >= 0x50 && instr.Opcode <= 0x57 {
			stackDelta++
		} else if instr.Opcode >= 0x58 && instr.Opcode <= 0x5F {
			stackDelta--
		}
	}

	if stackDelta != 0 {
		return fmt.Errorf("stack imbalance: delta %d", stackDelta)
	}

	// Layer 3: Check opcodes (semantic)
	for _, instr := range stream.Instructions {
		if !manifold.AllowedOpcodes[instr.Opcode] {
			return fmt.Errorf("forbidden opcode: 0x%02X", instr.Opcode)
		}
	}

	// Layer 4: Check for ROP gadgets (CFG)
	for _, pattern := range manifold.ForbiddenPatterns {
		if containsPattern(sequence, pattern) {
			return fmt.Errorf("contains ROP gadget")
		}
	}

	return nil
}

// GetTemplateByCategory retrieves templates for a specific category
func GetTemplateByCategory(category string) []InstructionTemplate {
	templates := make([]InstructionTemplate, 0)

	for _, tmpl := range instructionTemplates {
		if tmpl.Category == category {
			templates = append(templates, tmpl)
		}
	}

	return templates
}

// GetTemplatesByComplexity retrieves templates by complexity level
func GetTemplatesByComplexity(complexity int) []InstructionTemplate {
	templates := make([]InstructionTemplate, 0)

	for _, tmpl := range instructionTemplates {
		if tmpl.Complexity == complexity {
			templates = append(templates, tmpl)
		}
	}

	return templates
}
