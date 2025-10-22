package lito

/*
 * Lito - x86/x64 Instruction Length Disassembler
 *
 * Main API and convenience functions
 *
 * Novel implementation inspired by length disassembly techniques from
 * malware analysis, but with modern Go idioms and extended x64 support.
 *
 * Key differences from original samples:
 * - Clean API design
 * - Proper error handling
 * - x64 support (REX prefixes, RIP-relative, etc.)
 * - Thread-safe (no global state)
 * - Extensible for future instruction sets
 */

import (
	"fmt"
)

// InstructionStream represents a sequence of instructions
// Novel: Stream processing abstraction
type InstructionStream struct {
	Code         []byte
	Instructions []*Instruction
	Mode64       bool
}

// NewInstructionStream creates a new instruction stream
func NewInstructionStream(code []byte, mode64 bool) *InstructionStream {
	return &InstructionStream{
		Code:         code,
		Instructions: make([]*Instruction, 0),
		Mode64:       mode64,
	}
}

// ParseAll parses all instructions in the stream
func (s *InstructionStream) ParseAll() error {
	offset := 0

	for offset < len(s.Code) {
		instr, err := Disassemble(s.Code, offset, s.Mode64)
		if err != nil {
			return fmt.Errorf("failed to parse at offset %d: %w", offset, err)
		}

		s.Instructions = append(s.Instructions, instr)
		offset += int(instr.Length)
	}

	return nil
}

// GetTotalLength returns the total length of all instructions
func (s *InstructionStream) GetTotalLength() int {
	total := 0
	for _, instr := range s.Instructions {
		total += int(instr.Length)
	}
	return total
}

// GetControlFlowInstructions returns only control flow instructions
// Novel: Filter helper for Morpher's jump relocation
func (s *InstructionStream) GetControlFlowInstructions() []*Instruction {
	controlFlow := make([]*Instruction, 0)
	for _, instr := range s.Instructions {
		if instr.IsControlFlow() {
			controlFlow = append(controlFlow, instr)
		}
	}
	return controlFlow
}

// QuickLength provides fast instruction length calculation
// Novel: Optimized for common case (no full parsing)
func QuickLength(code []byte, offset int, mode64 bool) int {
	length, err := DisassembleLength(code, offset, mode64)
	if err != nil {
		// On error, assume single-byte instruction
		return 1
	}
	return length
}

// ValidateCodeBlock checks if a code block is valid x86/x64 code
// Novel: Validation helper for Morpher input
func ValidateCodeBlock(code []byte, mode64 bool) error {
	offset := 0
	instrCount := 0

	for offset < len(code) {
		_, err := Disassemble(code, offset, mode64)
		if err != nil {
			return fmt.Errorf("invalid instruction at offset %d: %w", offset, err)
		}

		length, _ := DisassembleLength(code, offset, mode64)
		if length == 0 {
			return fmt.Errorf("zero-length instruction at offset %d", offset)
		}

		offset += length
		instrCount++

		// Sanity check: prevent infinite loops
		if instrCount > 100000 {
			return fmt.Errorf("too many instructions (possible infinite loop)")
		}
	}

	return nil
}

// CodeStats provides statistics about a code block
// Novel: Analysis helper for debugging and optimization
type CodeStats struct {
	TotalBytes          int
	InstructionCount    int
	AverageLength       float64
	ControlFlowCount    int
	PrefixCount         int
	TwoByteOpcodes      int
	LongestInstruction  int
	ShortestInstruction int
}

// AnalyzeCode returns statistics about a code block
// Novel: Metrics for Morpher optimization decisions
func AnalyzeCode(code []byte, mode64 bool) (*CodeStats, error) {
	stream := NewInstructionStream(code, mode64)
	if err := stream.ParseAll(); err != nil {
		return nil, err
	}

	stats := &CodeStats{
		TotalBytes:          len(code),
		InstructionCount:    len(stream.Instructions),
		LongestInstruction:  0,
		ShortestInstruction: 255,
	}

	for _, instr := range stream.Instructions {
		// Update length stats
		instrLen := int(instr.Length)
		if instrLen > stats.LongestInstruction {
			stats.LongestInstruction = instrLen
		}
		if instrLen < stats.ShortestInstruction {
			stats.ShortestInstruction = instrLen
		}

		// Count prefixes
		stats.PrefixCount += len(instr.Prefixes)

		// Count two-byte opcodes
		if instr.Properties.IsTwoByteOpcode {
			stats.TwoByteOpcodes++
		}

		// Count control flow
		if instr.IsControlFlow() {
			stats.ControlFlowCount++
		}
	}

	if stats.InstructionCount > 0 {
		stats.AverageLength = float64(stats.TotalBytes) / float64(stats.InstructionCount)
	}

	return stats, nil
}

// FindInstructionBoundaries returns all instruction start offsets
// Novel: Helper for code block splitting/analysis
func FindInstructionBoundaries(code []byte, mode64 bool) ([]int, error) {
	boundaries := make([]int, 0)
	offset := 0

	for offset < len(code) {
		boundaries = append(boundaries, offset)

		length, err := DisassembleLength(code, offset, mode64)
		if err != nil {
			return boundaries, err
		}

		if length == 0 {
			return boundaries, fmt.Errorf("zero-length instruction at offset %d", offset)
		}

		offset += length
	}

	return boundaries, nil
}

// SplitAtBoundary splits code at instruction boundaries
// Novel: Safe code splitting (useful for Morpher's code block processing)
func SplitAtBoundary(code []byte, splitOffset int, mode64 bool) ([]byte, []byte, error) {
	// Find nearest instruction boundary at or after splitOffset
	boundaries, err := FindInstructionBoundaries(code, mode64)
	if err != nil {
		return nil, nil, err
	}

	// Find the boundary closest to splitOffset
	actualSplit := 0
	for _, boundary := range boundaries {
		if boundary >= splitOffset {
			actualSplit = boundary
			break
		}
	}

	if actualSplit == 0 && splitOffset > 0 {
		return nil, nil, fmt.Errorf("no instruction boundary found at or after offset %d", splitOffset)
	}

	return code[:actualSplit], code[actualSplit:], nil
}

// InstructionAt returns the instruction at a specific code offset
// Novel: Random access helper
func InstructionAt(code []byte, targetOffset int, mode64 bool) (*Instruction, error) {
	offset := 0

	for offset < len(code) {
		if offset == targetOffset {
			return Disassemble(code, offset, mode64)
		}

		length, err := DisassembleLength(code, offset, mode64)
		if err != nil {
			return nil, err
		}

		offset += length
	}

	return nil, fmt.Errorf("no instruction found at offset %d", targetOffset)
}

// CountInstructions quickly counts instructions without full parsing
// Novel: Performance optimization
func CountInstructions(code []byte, mode64 bool) (int, error) {
	count := 0
	offset := 0

	for offset < len(code) {
		length, err := DisassembleLength(code, offset, mode64)
		if err != nil {
			return count, err
		}

		if length == 0 {
			return count, fmt.Errorf("zero-length instruction at offset %d", offset)
		}

		offset += length
		count++

		// Sanity check
		if count > 1000000 {
			return count, fmt.Errorf("instruction count exceeded safety limit")
		}
	}

	return count, nil
}
