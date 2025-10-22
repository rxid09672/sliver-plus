package wig

import (
	"testing"

	"github.com/bishopfox/sliver/server/generate/lito"
	"github.com/bishopfox/sliver/server/generate/morpher"
)

// Test vector extraction from code
func TestVectorExtraction(t *testing.T) {
	// Simple code sequence
	code := []byte{
		0x90,       // NOP
		0x50,       // PUSH EAX
		0x89, 0xC0, // MOV EAX, EAX
		0x58, // POP EAX
		0xC3, // RET
	}

	vector, err := ExtractCodeVector(code, false)
	if err != nil {
		t.Fatalf("Failed to extract vector: %v", err)
	}

	if !vector.Valid {
		t.Error("Vector should be valid")
	}

	// Should detect NOP
	if vector.Values[DimNOPFrequency] == 0 {
		t.Error("Should detect NOP frequency")
	}

	// Should detect PUSH and POP (balanced)
	if vector.Values[DimPUSHFrequency] == 0 {
		t.Error("Should detect PUSH")
	}
	if vector.Values[DimPOPFrequency] == 0 {
		t.Error("Should detect POP")
	}
}

// Test executable manifold constraints
func TestManifoldConstraints(t *testing.T) {
	manifold := NewExecutableManifold(false, "windows")

	// Create vector with stack imbalance
	badVector := NewCodeVector()
	badVector.Values[DimPUSHFrequency] = 0.8 // 80% PUSH
	badVector.Values[DimPOPFrequency] = 0.2  // 20% POP (IMBALANCED!)
	badVector.Valid = true

	// Project onto manifold - should fix stack balance
	projected := manifold.Project(badVector)

	// Check if stack is now balanced
	pushFreq := projected.Values[DimPUSHFrequency]
	popFreq := projected.Values[DimPOPFrequency]

	diff := pushFreq - popFreq
	if diff < 0 {
		diff = -diff
	}

	if diff > 0.1 {
		t.Errorf("Stack not balanced after projection: PUSH=%.2f, POP=%.2f", pushFreq, popFreq)
	}
}

// Test distance metrics
func TestDistanceMetrics(t *testing.T) {
	v1 := NewCodeVector()
	v2 := NewCodeVector()

	// Identical vectors
	v1.Values[DimNOPFrequency] = 0.5
	v2.Values[DimNOPFrequency] = 0.5

	dist := EuclideanDistance(v1, v2)
	if dist != 0 {
		t.Errorf("Distance between identical vectors should be 0, got %.2f", dist)
	}

	// Different vectors
	v2.Values[DimNOPFrequency] = 0.8
	dist = EuclideanDistance(v1, v2)
	if dist == 0 {
		t.Error("Distance between different vectors should be > 0")
	}
}

// Test chain-of-thought generation
func TestChainOfThought(t *testing.T) {
	currentVector := NewCodeVector()
	currentVector.Values[DimNOPFrequency] = 0.5
	currentVector.Valid = true

	history := NewVectorHistory(10)
	manifold := NewExecutableManifold(false, "windows")
	rng := morpher.NewXorShift128(12345)

	cot := GenerateChainOfThought(currentVector, history, manifold, rng)

	if len(cot.Thoughts) == 0 {
		t.Error("Chain of thought should have thoughts")
	}

	if cot.TargetVector == nil {
		t.Error("Chain of thought should produce target vector")
	}

	// Validate that target satisfies constraints
	if !manifold.IsValid(cot.TargetVector) {
		t.Error("Target vector should satisfy manifold constraints")
	}
}

// Test template-based generation
func TestTemplateGeneration(t *testing.T) {
	rng := morpher.NewXorShift128(54321)
	vector := NewCodeVector()
	vector.Values[DimNOPFrequency] = 0.8 // High NOP preference
	vector.Valid = true

	// Generate instruction
	bytes, err := GenerateInstructionFromVector(vector, rng, false)
	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	if len(bytes) == 0 {
		t.Error("Should generate at least one byte")
	}

	// Should be parseable
	_, err = lito.Disassemble(bytes, 0, false)
	if err != nil {
		t.Errorf("Generated instruction not parseable: %v", err)
	}
}

// Test stack balance in generated sequences
func TestStackBalance(t *testing.T) {
	rng := morpher.NewXorShift128(99999)
	vector := NewCodeVector()
	vector.Values[DimPUSHFrequency] = 0.5
	vector.Values[DimPOPFrequency] = 0.5
	vector.Valid = true

	// Generate sequence
	sequence, err := GenerateSequenceFromVector(vector, 50, rng, false)
	if err != nil {
		t.Fatalf("Sequence generation failed: %v", err)
	}

	// Parse and check stack balance
	stream := lito.NewInstructionStream(sequence, false)
	stream.ParseAll()

	stackDelta := 0
	for _, instr := range stream.Instructions {
		if instr.Opcode >= 0x50 && instr.Opcode <= 0x57 {
			stackDelta++
		} else if instr.Opcode >= 0x58 && instr.Opcode <= 0x5F {
			stackDelta--
		}
	}

	if stackDelta != 0 {
		t.Errorf("Stack not balanced: delta = %d", stackDelta)
	}
}

// Test novelty scoring
func TestNoveltyScoring(t *testing.T) {
	history := NewVectorHistory(10)

	// Add some vectors to history
	for i := 0; i < 5; i++ {
		v := NewCodeVector()
		v.Values[DimNOPFrequency] = 0.5
		v.Valid = true
		history.Add(v)
	}

	// Create novel vector (different from history)
	novel := NewCodeVector()
	novel.Values[DimNOPFrequency] = 0.9 // Very different
	novel.Valid = true

	novelty := history.NoveltyScore(novel)
	if novelty == 0 {
		t.Error("Novel vector should have non-zero novelty score")
	}

	// Create similar vector
	similar := NewCodeVector()
	similar.Values[DimNOPFrequency] = 0.5 // Same as history
	similar.Valid = true

	similarNovelty := history.NoveltyScore(similar)

	// Novel should have higher score than similar
	if novelty <= similarNovelty {
		t.Errorf("Novel vector should have higher novelty (%.2f) than similar (%.2f)", novelty, similarNovelty)
	}
}

// Test alien region finding
func TestAlienRegionFinding(t *testing.T) {
	history := make([]*CodeVector, 0)

	// Create clustered history
	for i := 0; i < 10; i++ {
		v := NewCodeVector()
		v.Values[DimNOPFrequency] = 0.5 // All clustered around 0.5
		v.Valid = true
		history = append(history, v)
	}

	rng := morpher.NewXorShift128(11111)
	alien := FindAlienRegion(history, rng)

	// Alien should be far from 0.5
	nopFreq := alien.Values[DimNOPFrequency]
	distFrom05 := nopFreq - 0.5
	if distFrom05 < 0 {
		distFrom05 = -distFrom05
	}

	if distFrom05 < 0.2 {
		t.Errorf("Alien region should be far from cluster (0.5), got %.2f", nopFreq)
	}
}

// Test end-to-end vector-guided morphing
func TestVectorGuidedMorphing(t *testing.T) {
	// Simple code
	code := []byte{
		0x90, // NOP
		0x50, // PUSH EAX
		0x58, // POP EAX
		0xC3, // RET
	}

	vsw := NewVectorSpaceWig(12345, false, "windows")

	// Morph with vector guidance
	result, cot, err := vsw.MorphWithVectorGuidance(code)
	if err != nil {
		t.Fatalf("Vector-guided morphing failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Morphing unsuccessful: %v", result.Error)
	}

	// Output should be larger (expansion + dead code)
	if len(result.Code) <= len(code) {
		t.Errorf("Morphed code should be larger: %d vs %d", len(result.Code), len(code))
	}

	// Should have chain of thought
	if cot == nil {
		t.Error("Should generate chain of thought")
	}

	// History should have one entry
	if len(vsw.History.Vectors) != 1 {
		t.Errorf("History should have 1 entry, got %d", len(vsw.History.Vectors))
	}

	t.Logf("Vector-guided morph: %d → %d bytes", len(code), len(result.Code))
	if cot != nil {
		t.Logf("Chain of thought: %d thoughts", len(cot.Thoughts))
	}
}

// Test multiple generations show increasing diversity
func TestIncreasingDiversity(t *testing.T) {
	code := []byte{0x90, 0x90, 0x90, 0xC3} // Simple code

	vsw := NewVectorSpaceWig(33333, false, "windows")

	distances := make([]float64, 0)

	// Generate 5 builds
	for i := 0; i < 5; i++ {
		result, _, err := vsw.MorphWithVectorGuidance(code)
		if err != nil {
			t.Fatalf("Generation %d failed: %v", i, err)
		}

		if len(vsw.History.Vectors) > 1 {
			// Calculate distance from previous
			prev := vsw.History.Vectors[len(vsw.History.Vectors)-2]
			curr := vsw.History.Vectors[len(vsw.History.Vectors)-1]
			dist := WeightedEuclideanDistance(prev, curr)
			distances = append(distances, dist)
		}

		t.Logf("Build %d: %d bytes, novelty %.2f", i+1, len(result.Code), vsw.AvgNovelty)
	}

	// Check that builds are different from each other
	for i, dist := range distances {
		if dist == 0 {
			t.Errorf("Build %d is identical to previous (distance = 0)", i+2)
		}
	}
}

// Test constraint satisfaction
func TestConstraintSatisfaction(t *testing.T) {
	manifold := NewExecutableManifold(false, "windows")
	rng := morpher.NewXorShift128(44444)

	// Generate many random vectors and project
	for i := 0; i < 100; i++ {
		random := RandomValidVector(rng)
		projected := manifold.Project(random)

		// All projected vectors must be valid
		if !manifold.IsValid(projected) {
			t.Errorf("Projected vector %d failed validation", i)
		}

		// Stack balance check
		pushFreq := projected.Values[DimPUSHFrequency]
		popFreq := projected.Values[DimPOPFrequency]
		diff := pushFreq - popFreq
		if diff < 0 {
			diff = -diff
		}
		if diff > 0.1 {
			t.Errorf("Vector %d: Stack imbalance %.2f", i, diff)
		}
	}
}

// Benchmark vector extraction
func BenchmarkVectorExtraction(b *testing.B) {
	code := make([]byte, 1000)
	for i := 0; i < 1000; i++ {
		code[i] = 0x90 // NOP
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExtractCodeVector(code, false)
	}
}

// Benchmark chain-of-thought generation
func BenchmarkChainOfThought(b *testing.B) {
	vector := NewCodeVector()
	vector.Valid = true
	history := NewVectorHistory(10)
	manifold := NewExecutableManifold(false, "windows")
	rng := morpher.NewXorShift128(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateChainOfThought(vector, history, manifold, rng)
	}
}

// Benchmark vector-guided morphing
func BenchmarkVectorGuidedMorph(b *testing.B) {
	code := []byte{0x90, 0x50, 0x58, 0xC3}
	vsw := NewVectorSpaceWig(55555, false, "windows")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vsw.MorphWithVectorGuidance(code)
	}
}
