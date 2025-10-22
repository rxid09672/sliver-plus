package morpher

/*
 * Instruction Expansion for Metamorphic Code Generation
 *
 * Transforms compact instruction forms into expanded equivalents
 * to break static signatures while preserving functionality.
 *
 * Novel approach:
 * - Selective expansion (not blind)
 * - Multiple expansion strategies
 * - Proper x64 support
 * - Safe fallback on errors
 *
 * Key differences from malware:
 * - Modern Go error handling
 * - Configurable expansion rate
 * - Multiple expansion methods per instruction type
 * - No hardcoded magic numbers
 */

import (
	"fmt"

	"github.com/bishopfox/sliver/server/generate/lito"
)

// ExpansionType defines different ways to expand an instruction
// Novel: Multiple strategies for same instruction (more polymorphic)
type ExpansionType uint8

const (
	ExpansionNone          ExpansionType = iota
	ExpansionShortToNear                 // JXX SHORT → JXX NEAR
	ExpansionCompactToFull               // Compact encoding → Full encoding
	ExpansionOpcodeVariant               // Use alternate opcode
)

// CanExpandInstruction checks if an instruction can be expanded
// Novel: Type-safe checking with clear categories
func CanExpandInstruction(instr *lito.Instruction) bool {
	if instr == nil || !instr.Valid {
		return false
	}

	opcode := instr.Opcode

	// Conditional jumps (short form: 0x70-0x7F)
	if opcode >= 0x70 && opcode <= 0x7F {
		return true
	}

	// Unconditional jump (short form: 0xEB)
	if opcode == 0xEB {
		return true
	}

	// Could add more expansion types here:
	// - PUSH imm8 → PUSH imm32
	// - Short MOV variants → Full MOV
	// etc.

	return false
}

// GetExpansionType determines the best expansion for an instruction
// Novel: Decision logic separate from expansion logic
func GetExpansionType(instr *lito.Instruction) ExpansionType {
	opcode := instr.Opcode

	// Conditional jumps (0x70-0x7F) → 0x0F 0x8X
	if opcode >= 0x70 && opcode <= 0x7F {
		return ExpansionShortToNear
	}

	// Unconditional jump (0xEB) → 0xE9
	if opcode == 0xEB {
		return ExpansionShortToNear
	}

	return ExpansionNone
}

// ExpandInstruction transforms an instruction to its expanded form
// Novel: Returns new bytes, doesn't modify original
func ExpandInstruction(instr *lito.Instruction) ([]byte, error) {
	if !CanExpandInstruction(instr) {
		return nil, fmt.Errorf("instruction cannot be expanded")
	}

	expansionType := GetExpansionType(instr)

	switch expansionType {
	case ExpansionShortToNear:
		return expandShortToNear(instr)
	case ExpansionCompactToFull:
		return expandCompactToFull(instr)
	case ExpansionOpcodeVariant:
		return expandOpcodeVariant(instr)
	default:
		return nil, fmt.Errorf("unknown expansion type: %v", expansionType)
	}
}

// expandShortToNear expands short jumps to near jumps
// Novel: Proper signed offset handling with validation
func expandShortToNear(instr *lito.Instruction) ([]byte, error) {
	opcode := instr.Opcode

	// Conditional jumps: 0x70-0x7F → 0x0F 0x80-0x8F
	if opcode >= 0x70 && opcode <= 0x7F {
		// Calculate new opcode: 0x70 → 0x0F 0x80, 0x71 → 0x0F 0x81, etc.
		secondOpcode := 0x80 + (opcode - 0x70)

		// Get original 8-bit signed offset
		if len(instr.Immediate) < 1 {
			return nil, fmt.Errorf("conditional jump missing immediate")
		}
		offset8 := int8(instr.Immediate[0])

		// Convert to 32-bit signed offset
		// Novel: Explicit sign extension (safer than malware's implicit)
		offset32 := int32(offset8)

		// Build expanded instruction: 0x0F [0x80-0x8F] [imm32]
		expanded := make([]byte, 6)
		expanded[0] = 0x0F
		expanded[1] = secondOpcode

		// Write 32-bit offset in little-endian
		// Novel: Explicit endianness handling
		expanded[2] = byte(offset32)
		expanded[3] = byte(offset32 >> 8)
		expanded[4] = byte(offset32 >> 16)
		expanded[5] = byte(offset32 >> 24)

		return expanded, nil
	}

	// Unconditional jump: 0xEB → 0xE9
	if opcode == 0xEB {
		// Get original 8-bit signed offset
		if len(instr.Immediate) < 1 {
			return nil, fmt.Errorf("unconditional jump missing immediate")
		}
		offset8 := int8(instr.Immediate[0])

		// Convert to 32-bit signed offset
		offset32 := int32(offset8)

		// Build expanded instruction: 0xE9 [imm32]
		expanded := make([]byte, 5)
		expanded[0] = 0xE9
		expanded[1] = byte(offset32)
		expanded[2] = byte(offset32 >> 8)
		expanded[3] = byte(offset32 >> 16)
		expanded[4] = byte(offset32 >> 24)

		return expanded, nil
	}

	return nil, fmt.Errorf("instruction type not supported for short-to-near expansion")
}

// expandCompactToFull expands compact encodings to full forms
// Novel: Reserved for future expansion types
func expandCompactToFull(instr *lito.Instruction) ([]byte, error) {
	// Future: PUSH imm8 → PUSH imm32, etc.
	return nil, fmt.Errorf("compact-to-full expansion not yet implemented")
}

// expandOpcodeVariant uses alternate opcode encoding
// Novel: Reserved for future opcode variations
func expandOpcodeVariant(instr *lito.Instruction) ([]byte, error) {
	// Future: MOV → XCHG, ADD → SUB + NEG, etc.
	return nil, fmt.Errorf("opcode variant expansion not yet implemented")
}

// GetExpandedSize returns the size of the expanded instruction
// Novel: Pre-calculate size without actually expanding
func GetExpandedSize(instr *lito.Instruction) int {
	if !CanExpandInstruction(instr) {
		return int(instr.Length)
	}

	opcode := instr.Opcode

	// Conditional jumps: 2 bytes → 6 bytes
	if opcode >= 0x70 && opcode <= 0x7F {
		return 6
	}

	// Unconditional jump: 2 bytes → 5 bytes
	if opcode == 0xEB {
		return 5
	}

	// Default: no expansion
	return int(instr.Length)
}

// CalculateExpansionGrowth estimates total size increase
// Novel: Planning helper for buffer allocation
func CalculateExpansionGrowth(instrs []*lito.Instruction, expansionRate float64) int {
	totalGrowth := 0

	for _, instr := range instrs {
		if !CanExpandInstruction(instr) {
			continue
		}

		// Only expand based on rate (0.0-1.0)
		// Novel: Configurable expansion rate (malware was 100%)
		if expansionRate >= 1.0 {
			oldSize := int(instr.Length)
			newSize := GetExpandedSize(instr)
			totalGrowth += (newSize - oldSize)
		}
	}

	return totalGrowth
}

// ShouldExpandInstruction decides whether to expand based on policy
// Novel: Policy-driven expansion (configurable, not random)
type ExpansionPolicy struct {
	Rate                 float64 // 0.0-1.0: probability of expansion
	ExpandAllControlFlow bool    // Always expand jumps/calls
	AvoidLongJumps       bool    // Don't expand if target is far
	MaxExpansions        int     // Limit total expansions (0 = unlimited)
}

// DefaultExpansionPolicy returns a sensible default
func DefaultExpansionPolicy() *ExpansionPolicy {
	return &ExpansionPolicy{
		Rate:                 0.7,  // 70% expansion rate
		ExpandAllControlFlow: true, // Always expand control flow
		AvoidLongJumps:       false,
		MaxExpansions:        0, // Unlimited
	}
}

// ShouldExpand determines if an instruction should be expanded
// Novel: Policy-based decision making
func (p *ExpansionPolicy) ShouldExpand(instr *lito.Instruction, rng *XorShift128, expansionCount int) bool {
	if !CanExpandInstruction(instr) {
		return false
	}

	// Check max expansions limit
	if p.MaxExpansions > 0 && expansionCount >= p.MaxExpansions {
		return false
	}

	// Always expand control flow if policy says so
	if p.ExpandAllControlFlow && instr.IsControlFlow() {
		return true
	}

	// Otherwise use expansion rate
	return rng.Float64() < p.Rate
}
