package lito

import (
	"testing"
)

// Test simple single-byte instructions
func TestSingleByteInstructions(t *testing.T) {
	tests := []struct {
		name     string
		code     []byte
		expected int
	}{
		{"NOP", []byte{0x90}, 1},
		{"PUSH EAX", []byte{0x50}, 1},
		{"PUSH ECX", []byte{0x51}, 1},
		{"POP EAX", []byte{0x58}, 1},
		{"POP EDI", []byte{0x5F}, 1},
		{"RET", []byte{0xC3}, 1},
		{"INT3", []byte{0xCC}, 1},
		{"CLC", []byte{0xF8}, 1},
		{"STC", []byte{0xF9}, 1},
		{"PUSHA", []byte{0x60}, 1},
		{"POPA", []byte{0x61}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			length, err := DisassembleLength(tt.code, 0, false)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if length != tt.expected {
				t.Errorf("Expected length %d, got %d", tt.expected, length)
			}
		})
	}
}

// Test MODRM instructions
func TestModRMInstructions(t *testing.T) {
	tests := []struct {
		name     string
		code     []byte
		expected int
	}{
		{"MOV EAX, EBX", []byte{0x89, 0xD8}, 2},
		{"MOV EBX, EAX", []byte{0x89, 0xC3}, 2},
		{"ADD EAX, EBX", []byte{0x01, 0xD8}, 2},
		{"XOR ECX, ECX", []byte{0x31, 0xC9}, 2},
		{"TEST EAX, EAX", []byte{0x85, 0xC0}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			length, err := DisassembleLength(tt.code, 0, false)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if length != tt.expected {
				t.Errorf("Expected length %d, got %d", tt.expected, length)
			}
		})
	}
}

// Test immediate instructions
func TestImmediateInstructions(t *testing.T) {
	tests := []struct {
		name     string
		code     []byte
		expected int
	}{
		{"ADD AL, 0x12", []byte{0x04, 0x12}, 2},
		{"ADD EAX, 0x12345678", []byte{0x05, 0x78, 0x56, 0x34, 0x12}, 5},
		{"PUSH 0x42", []byte{0x6A, 0x42}, 2},
		{"PUSH 0x12345678", []byte{0x68, 0x78, 0x56, 0x34, 0x12}, 5},
		{"MOV AL, 0xFF", []byte{0xB0, 0xFF}, 2},
		{"MOV EAX, 0x12345678", []byte{0xB8, 0x78, 0x56, 0x34, 0x12}, 5},
		{"RET 0x10", []byte{0xC2, 0x10, 0x00}, 3},
		{"INT 0x80", []byte{0xCD, 0x80}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			length, err := DisassembleLength(tt.code, 0, false)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if length != tt.expected {
				t.Errorf("Expected length %d, got %d", tt.expected, length)
			}
		})
	}
}

// Test relative jump/call instructions
func TestRelativeJumps(t *testing.T) {
	tests := []struct {
		name     string
		code     []byte
		expected int
	}{
		{"JE SHORT +0x10", []byte{0x74, 0x10}, 2},
		{"JNE SHORT +0x20", []byte{0x75, 0x20}, 2},
		{"JMP SHORT +0x7F", []byte{0xEB, 0x7F}, 2},
		{"JMP SHORT -0x10", []byte{0xEB, 0xF0}, 2},
		{"CALL +0x12345678", []byte{0xE8, 0x78, 0x56, 0x34, 0x12}, 5},
		{"JMP +0x12345678", []byte{0xE9, 0x78, 0x56, 0x34, 0x12}, 5},
		{"LOOP +0x10", []byte{0xE2, 0x10}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			length, err := DisassembleLength(tt.code, 0, false)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if length != tt.expected {
				t.Errorf("Expected length %d, got %d", tt.expected, length)
			}

			// Test control flow detection
			instr, _ := Disassemble(tt.code, 0, false)
			if !instr.IsControlFlow() {
				t.Errorf("Expected control flow instruction")
			}
		})
	}
}

// Test two-byte opcodes
func TestTwoByteOpcodes(t *testing.T) {
	tests := []struct {
		name     string
		code     []byte
		expected int
	}{
		{"JE NEAR +0x100", []byte{0x0F, 0x84, 0x00, 0x01, 0x00, 0x00}, 6},
		{"JNE NEAR +0x200", []byte{0x0F, 0x85, 0x00, 0x02, 0x00, 0x00}, 6},
		{"SETE AL", []byte{0x0F, 0x94, 0xC0}, 3},
		{"MOVZX EAX, BL", []byte{0x0F, 0xB6, 0xC3}, 3},
		{"RDTSC", []byte{0x0F, 0x31}, 2},
		{"CPUID", []byte{0x0F, 0xA2}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			length, err := DisassembleLength(tt.code, 0, false)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if length != tt.expected {
				t.Errorf("Expected length %d, got %d", tt.expected, length)
			}
		})
	}
}

// Test prefixed instructions
func TestPrefixedInstructions(t *testing.T) {
	tests := []struct {
		name     string
		code     []byte
		expected int
	}{
		{"REP MOVSB", []byte{0xF3, 0xA4}, 2},
		{"REP MOVSD", []byte{0xF3, 0xA5}, 2},
		{"REPNE SCASB", []byte{0xF2, 0xAE}, 2},
		{"LOCK ADD [EAX], EBX", []byte{0xF0, 0x01, 0x18}, 3},
		{"FS: MOV EAX, [EBX]", []byte{0x64, 0x8B, 0x03}, 3},
		{"GS: MOV ECX, [EDX]", []byte{0x65, 0x8B, 0x0A}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			length, err := DisassembleLength(tt.code, 0, false)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if length != tt.expected {
				t.Errorf("Expected length %d, got %d", tt.expected, length)
			}

			// Verify prefix detection
			instr, _ := Disassemble(tt.code, 0, false)
			if len(instr.Prefixes) == 0 {
				t.Errorf("Expected prefix to be detected")
			}
		})
	}
}

// Test MODRM with displacement
func TestModRMWithDisplacement(t *testing.T) {
	tests := []struct {
		name     string
		code     []byte
		expected int
	}{
		{"MOV EAX, [EBX+0x10]", []byte{0x8B, 0x43, 0x10}, 3},
		{"MOV EAX, [EBX+0x12345678]", []byte{0x8B, 0x83, 0x78, 0x56, 0x34, 0x12}, 6},
		{"MOV [ECX+0x20], EDX", []byte{0x89, 0x51, 0x20}, 3},
		{"ADD [EDI], 0x42", []byte{0x83, 0x07, 0x42}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			length, err := DisassembleLength(tt.code, 0, false)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if length != tt.expected {
				t.Errorf("Expected length %d, got %d", tt.expected, length)
			}
		})
	}
}

// Test SIB byte instructions
func TestSIBInstructions(t *testing.T) {
	tests := []struct {
		name     string
		code     []byte
		expected int
	}{
		{"MOV EAX, [ESP]", []byte{0x8B, 0x04, 0x24}, 3},
		{"MOV EAX, [ESP+0x10]", []byte{0x8B, 0x44, 0x24, 0x10}, 4},
		{"MOV EAX, [EBP+ESI*4]", []byte{0x8B, 0x04, 0xB5, 0x00, 0x00, 0x00, 0x00}, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			length, err := DisassembleLength(tt.code, 0, false)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if length != tt.expected {
				t.Errorf("Expected length %d, got %d", tt.expected, length)
			}

			// Verify SIB detection
			instr, _ := Disassemble(tt.code, 0, false)
			if !instr.Properties.HasSIB {
				t.Errorf("Expected SIB byte to be detected")
			}
		})
	}
}

// Test x64 REX prefixes
func TestREXPrefixes(t *testing.T) {
	tests := []struct {
		name     string
		code     []byte
		expected int
		mode64   bool
	}{
		{"REX.W + ADD", []byte{0x48, 0x01, 0xC3}, 3, true},
		{"REX.W + MOV", []byte{0x48, 0x89, 0xC0}, 3, true},
		{"REX + PUSH", []byte{0x41, 0x50}, 2, true},

		// In x86 mode, 0x40-0x4F are INC/DEC, not REX
		{"INC EAX (x86)", []byte{0x40}, 1, false},
		{"DEC EAX (x86)", []byte{0x48}, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			length, err := DisassembleLength(tt.code, 0, tt.mode64)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if length != tt.expected {
				t.Errorf("Expected length %d, got %d", tt.expected, length)
			}

			// In x64 mode, check REX detection
			if tt.mode64 && tt.code[0] >= 0x40 && tt.code[0] <= 0x4F {
				instr, _ := Disassemble(tt.code, 0, tt.mode64)
				if !instr.Properties.HasREX {
					t.Errorf("Expected REX prefix to be detected")
				}
			}
		})
	}
}

// Test relative jump target calculation
func TestRelativeTargetCalculation(t *testing.T) {
	tests := []struct {
		name        string
		code        []byte
		instrAddr   uint64
		expectedTgt uint64
	}{
		{"JE SHORT +0x10", []byte{0x74, 0x10}, 0x1000, 0x1012},
		{"JMP SHORT -0x10", []byte{0xEB, 0xF0}, 0x1000, 0x0FF2},
		{"CALL +0x100", []byte{0xE8, 0x00, 0x01, 0x00, 0x00}, 0x2000, 0x2105},
		{"JNE NEAR +0x1000", []byte{0x0F, 0x85, 0x00, 0x10, 0x00, 0x00}, 0x3000, 0x4006},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instr, err := Disassemble(tt.code, 0, false)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			target, err := instr.GetRelativeTarget(tt.instrAddr)
			if err != nil {
				t.Fatalf("Failed to calculate target: %v", err)
			}

			if target != tt.expectedTgt {
				t.Errorf("Expected target 0x%X, got 0x%X", tt.expectedTgt, target)
			}
		})
	}
}

// Test control flow detection
func TestControlFlowDetection(t *testing.T) {
	tests := []struct {
		name          string
		code          []byte
		isControlFlow bool
	}{
		{"JE SHORT", []byte{0x74, 0x10}, true},
		{"JMP SHORT", []byte{0xEB, 0x20}, true},
		{"CALL", []byte{0xE8, 0x00, 0x00, 0x00, 0x00}, true},
		{"RET", []byte{0xC3}, true},
		{"JNE NEAR", []byte{0x0F, 0x85, 0x00, 0x00, 0x00, 0x00}, true},

		{"MOV", []byte{0x89, 0xC0}, false},
		{"ADD", []byte{0x01, 0xC3}, false},
		{"NOP", []byte{0x90}, false},
		{"PUSH", []byte{0x50}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instr, err := Disassemble(tt.code, 0, false)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if instr.IsControlFlow() != tt.isControlFlow {
				t.Errorf("Expected IsControlFlow=%v, got %v", tt.isControlFlow, instr.IsControlFlow())
			}
		})
	}
}

// Test instruction stream parsing
func TestInstructionStream(t *testing.T) {
	// Simple code sequence
	code := []byte{
		0x50,       // PUSH EAX
		0x51,       // PUSH ECX
		0x89, 0xC8, // MOV EAX, ECX
		0x05, 0x10, 0x00, 0x00, 0x00, // ADD EAX, 0x10
		0x59, // POP ECX
		0x58, // POP EAX
		0xC3, // RET
	}

	stream := NewInstructionStream(code, false)
	err := stream.ParseAll()
	if err != nil {
		t.Fatalf("Failed to parse stream: %v", err)
	}

	// Should have 7 instructions
	if len(stream.Instructions) != 7 {
		t.Errorf("Expected 7 instructions, got %d", len(stream.Instructions))
	}

	// Total length should equal code length
	if stream.GetTotalLength() != len(code) {
		t.Errorf("Expected total length %d, got %d", len(code), stream.GetTotalLength())
	}

	// Should have 1 control flow instruction (RET)
	controlFlow := stream.GetControlFlowInstructions()
	if len(controlFlow) != 1 {
		t.Errorf("Expected 1 control flow instruction, got %d", len(controlFlow))
	}
}

// Test code validation
func TestCodeValidation(t *testing.T) {
	validCode := []byte{
		0x90, // NOP
		0x50, // PUSH EAX
		0x58, // POP EAX
		0xC3, // RET
	}

	err := ValidateCodeBlock(validCode, false)
	if err != nil {
		t.Errorf("Valid code rejected: %v", err)
	}

	// Test with truncated instruction (should fail gracefully)
	invalidCode := []byte{
		0x90, // NOP
		0xE8, // CALL (incomplete - missing offset)
	}

	err = ValidateCodeBlock(invalidCode, false)
	if err == nil {
		t.Errorf("Invalid code accepted")
	}
}

// Test code statistics
func TestCodeStats(t *testing.T) {
	code := []byte{
		0x50,       // PUSH EAX (1 byte)
		0x89, 0xC8, // MOV EAX, ECX (2 bytes)
		0x05, 0x10, 0x00, 0x00, 0x00, // ADD EAX, 0x10 (5 bytes)
		0x74, 0x05, // JE SHORT (2 bytes)
		0xC3, // RET (1 byte)
	}

	stats, err := AnalyzeCode(code, false)
	if err != nil {
		t.Fatalf("Failed to analyze code: %v", err)
	}

	if stats.InstructionCount != 5 {
		t.Errorf("Expected 5 instructions, got %d", stats.InstructionCount)
	}

	if stats.TotalBytes != len(code) {
		t.Errorf("Expected %d total bytes, got %d", len(code), stats.TotalBytes)
	}

	if stats.ControlFlowCount != 2 {
		t.Errorf("Expected 2 control flow instructions, got %d", stats.ControlFlowCount)
	}

	if stats.LongestInstruction != 5 {
		t.Errorf("Expected longest instruction 5, got %d", stats.LongestInstruction)
	}

	if stats.ShortestInstruction != 1 {
		t.Errorf("Expected shortest instruction 1, got %d", stats.ShortestInstruction)
	}
}

// Test instruction boundary detection
func TestInstructionBoundaries(t *testing.T) {
	code := []byte{
		0x90, // NOP - boundary at 0
		0x50, // PUSH - boundary at 1
		0x58, // POP - boundary at 2
		0xC3, // RET - boundary at 3
	}

	boundaries, err := FindInstructionBoundaries(code, false)
	if err != nil {
		t.Fatalf("Failed to find boundaries: %v", err)
	}

	expected := []int{0, 1, 2, 3}
	if len(boundaries) != len(expected) {
		t.Errorf("Expected %d boundaries, got %d", len(expected), len(boundaries))
	}

	for i, boundary := range boundaries {
		if boundary != expected[i] {
			t.Errorf("Boundary %d: expected %d, got %d", i, expected[i], boundary)
		}
	}
}

// Test code splitting at boundaries
func TestCodeSplitting(t *testing.T) {
	code := []byte{
		0x90,       // NOP
		0x50,       // PUSH EAX
		0x89, 0xC8, // MOV EAX, ECX
		0x58, // POP EAX
		0xC3, // RET
	}

	// Split after PUSH (offset 2)
	part1, part2, err := SplitAtBoundary(code, 2, false)
	if err != nil {
		t.Fatalf("Failed to split: %v", err)
	}

	// Part 1 should be NOP + PUSH (2 bytes)
	if len(part1) != 2 {
		t.Errorf("Expected part1 length 2, got %d", len(part1))
	}

	// Part 2 should be rest (5 bytes)
	if len(part2) != 5 {
		t.Errorf("Expected part2 length 5, got %d", len(part2))
	}

	// Both parts should be valid code
	if err := ValidateCodeBlock(part1, false); err != nil {
		t.Errorf("Part1 invalid: %v", err)
	}
	if err := ValidateCodeBlock(part2, false); err != nil {
		t.Errorf("Part2 invalid: %v", err)
	}
}

// Test instruction counting
func TestInstructionCounting(t *testing.T) {
	code := []byte{
		0x90, 0x90, 0x90, // 3x NOP
		0x50, 0x51, 0x52, // 3x PUSH
		0x5A, 0x59, 0x58, // 3x POP
		0xC3, // RET
	}

	count, err := CountInstructions(code, false)
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}

	if count != 10 {
		t.Errorf("Expected 10 instructions, got %d", count)
	}
}

// Test error handling with malformed code
func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name string
		code []byte
	}{
		{"Empty code", []byte{}},
		{"Truncated CALL", []byte{0xE8, 0x00}},
		{"Truncated two-byte", []byte{0x0F}},
		{"Truncated MODRM", []byte{0x89}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Disassemble(tt.code, 0, false)
			if err == nil {
				t.Errorf("Expected error for malformed code")
			}
		})
	}
}

// Benchmark instruction length calculation
func BenchmarkDisassembleLength(b *testing.B) {
	code := []byte{0x89, 0xC8} // MOV EAX, ECX

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DisassembleLength(code, 0, false)
	}
}

// Benchmark full instruction parsing
func BenchmarkDisassembleFull(b *testing.B) {
	code := []byte{0x89, 0xC8} // MOV EAX, ECX

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Disassemble(code, 0, false)
	}
}

// Benchmark stream parsing
func BenchmarkDisassembleStream(b *testing.B) {
	// 100 instructions
	code := make([]byte, 0, 200)
	for i := 0; i < 100; i++ {
		code = append(code, 0x90) // NOP
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stream := NewInstructionStream(code, false)
		stream.ParseAll()
	}
}
