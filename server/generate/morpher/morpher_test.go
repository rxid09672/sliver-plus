package morpher

import (
	"testing"

	"github.com/bishopfox/sliver/server/generate/lito"
)

// Test address tracker functionality
func TestAddressTracker(t *testing.T) {
	tracker := NewAddressTracker()

	// Add entries
	entry1 := &AddressEntry{OldOffset: 0, NewOffset: 0, OldLength: 2, NewLength: 6}
	entry2 := &AddressEntry{OldOffset: 2, NewOffset: 6, OldLength: 2, NewLength: 2}

	tracker.Add(entry1)
	tracker.Add(entry2)

	// Test lookups
	newAddr, ok := tracker.GetNewAddress(0)
	if !ok || newAddr != 0 {
		t.Errorf("Expected new address 0, got %d", newAddr)
	}

	newAddr, ok = tracker.GetNewAddress(2)
	if !ok || newAddr != 6 {
		t.Errorf("Expected new address 6, got %d", newAddr)
	}

	// Test reverse lookup
	oldAddr, ok := tracker.GetOldAddress(6)
	if !ok || oldAddr != 2 {
		t.Errorf("Expected old address 2, got %d", oldAddr)
	}
}

// Test Xorshift-128 RNG
func TestXorShift128(t *testing.T) {
	rng := NewXorShift128(12345)

	// Test basic generation
	val1 := rng.Uint32()
	val2 := rng.Uint32()

	if val1 == val2 {
		t.Error("Sequential values should be different")
	}

	// Test Intn
	for i := 0; i < 100; i++ {
		val := rng.Intn(10)
		if val < 0 || val >= 10 {
			t.Errorf("Intn(10) returned %d (out of range)", val)
		}
	}

	// Test Float64
	for i := 0; i < 100; i++ {
		val := rng.Float64()
		if val < 0.0 || val >= 1.0 {
			t.Errorf("Float64() returned %f (out of range)", val)
		}
	}
}

// Test reproducibility with seeds
func TestReproducibleSeeds(t *testing.T) {
	seed := uint32(42)

	rng1 := NewXorShift128(seed)
	rng2 := NewXorShift128(seed)

	// Should generate identical sequences
	for i := 0; i < 100; i++ {
		val1 := rng1.Uint32()
		val2 := rng2.Uint32()

		if val1 != val2 {
			t.Errorf("Mismatch at iteration %d: %d != %d", i, val1, val2)
		}
	}
}

// Test instruction expansion
func TestInstructionExpansion(t *testing.T) {
	tests := []struct {
		name         string
		code         []byte
		canExpand    bool
		expandedSize int
	}{
		{"JE SHORT", []byte{0x74, 0x10}, true, 6},
		{"JNE SHORT", []byte{0x75, 0x20}, true, 6},
		{"JMP SHORT", []byte{0xEB, 0x30}, true, 5},
		{"MOV (no expand)", []byte{0x89, 0xC0}, false, 2},
		{"NOP (no expand)", []byte{0x90}, false, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instr, err := lito.Disassemble(tt.code, 0, false)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			// Test expansion capability
			if CanExpandInstruction(instr) != tt.canExpand {
				t.Errorf("Expected canExpand=%v", tt.canExpand)
			}

			// Test expansion
			if tt.canExpand {
				expanded, err := ExpandInstruction(instr)
				if err != nil {
					t.Fatalf("Expansion failed: %v", err)
				}

				if len(expanded) != tt.expandedSize {
					t.Errorf("Expected expanded size %d, got %d", tt.expandedSize, len(expanded))
				}
			}
		})
	}
}

// Test dead code generation
func TestDeadCodeGeneration(t *testing.T) {
	rng := NewXorShift128(12345)
	config := DefaultDeadCodeConfig()

	// Generate dead code
	for i := 0; i < 100; i++ {
		deadCode := GenerateDeadCode(rng, config)

		if len(deadCode) == 0 {
			t.Error("Generated empty dead code")
		}

		if len(deadCode) > config.MaxLength {
			t.Errorf("Dead code length %d exceeds max %d", len(deadCode), config.MaxLength)
		}
	}
}

// Test simple morph operation
func TestSimpleMorph(t *testing.T) {
	// Simple code: JE +0x05, NOP, NOP, NOP, NOP, NOP, RET
	code := []byte{
		0x74, 0x05, // JE SHORT +0x05
		0x90, // NOP
		0x90, // NOP
		0x90, // NOP
		0x90, // NOP
		0x90, // NOP
		0xC3, // RET
	}

	config := DefaultMorphConfig()
	config.Seed = 12345
	config.EnableExpansion = true
	config.EnableDeadCode = true

	morpher := NewMorpher(config)
	result, err := morpher.Morph(code)

	if err != nil {
		t.Fatalf("Morph failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Morph unsuccessful: %v", result.Error)
	}

	// Output should be larger (expansion + dead code)
	if len(result.Code) <= len(code) {
		t.Errorf("Expected morphed code to be larger: %d vs %d", len(result.Code), len(code))
	}

	// Should have expanded the JE
	if result.ExpandedCount == 0 {
		t.Error("Expected at least one expansion")
	}

	t.Logf("Morph stats: %d → %d bytes (%.2fx)", result.OriginalSize, result.MorphedSize, result.ExpansionRatio)
}

// Test morph with no expansion (dead code only)
func TestMorphDeadCodeOnly(t *testing.T) {
	code := []byte{0x90, 0x90, 0x90, 0xC3} // NOP, NOP, NOP, RET

	config := DefaultMorphConfig()
	config.EnableExpansion = false
	config.EnableDeadCode = true
	config.Seed = 54321

	result, err := MorphWithConfig(code, config)

	if err != nil {
		t.Fatalf("Morph failed: %v", err)
	}

	// Should have added dead code
	if result.DeadCodeBytes == 0 {
		t.Error("Expected dead code to be injected")
	}
}

// Test reproducible morphing
func TestReproducibleMorph(t *testing.T) {
	code := []byte{
		0x74, 0x05, // JE SHORT
		0x90, 0x90, 0x90, 0x90, 0x90,
		0xC3, // RET
	}

	seed := uint32(99999)

	// Morph twice with same seed
	result1, err1 := MorphCode(code, seed)
	result2, err2 := MorphCode(code, seed)

	if err1 != nil || err2 != nil {
		t.Fatalf("Morph failed: %v, %v", err1, err2)
	}

	// Results should be identical
	if len(result1.Code) != len(result2.Code) {
		t.Errorf("Reproducibility failed: different sizes %d vs %d",
			len(result1.Code), len(result2.Code))
	}

	// Byte-for-byte comparison
	for i := 0; i < len(result1.Code) && i < len(result2.Code); i++ {
		if result1.Code[i] != result2.Code[i] {
			t.Errorf("Reproducibility failed at byte %d: %02X vs %02X",
				i, result1.Code[i], result2.Code[i])
			break
		}
	}
}

// Test that morphing preserves code semantics
func TestSemanticPreservation(t *testing.T) {
	// Code that returns (simple)
	code := []byte{0xC3} // RET

	result, err := MorphCode(code, 11111)
	if err != nil {
		t.Fatalf("Morph failed: %v", err)
	}

	// Should still end with RET
	if result.Code[len(result.Code)-1] != 0xC3 {
		t.Error("Morphed code doesn't preserve final RET")
	}
}

// Test relocation stats
func TestRelocationStats(t *testing.T) {
	tracker := NewAddressTracker()

	// Add some control flow entries
	entry1 := &AddressEntry{
		OldOffset: 0, NewOffset: 0,
		OldLength: 2, NewLength: 6,
		JumpTarget:    7,
		IsControlFlow: true,
		Expanded:      true,
	}

	tracker.Add(entry1)

	stats := GetRelocationStats(tracker, []byte{})

	if stats.TotalRelocations == 0 {
		t.Error("Expected relocations to be counted")
	}
}

// Benchmark morphing performance
func BenchmarkMorph(b *testing.B) {
	code := make([]byte, 1000)
	for i := 0; i < 1000; i++ {
		code[i] = 0x90 // NOP
	}

	config := DefaultMorphConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MorphWithConfig(code, config)
	}
}

// Benchmark RNG performance
func BenchmarkXorShift128(b *testing.B) {
	rng := NewXorShift128(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rng.Uint32()
	}
}
