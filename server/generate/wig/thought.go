package wig

/*
 * Chain-of-Thought Vector Transformation
 *
 * Generates a reasoning chain that explains HOW to transform from current
 * code vector to target vector while satisfying all safety constraints.
 *
 * Novel approach: Inspired by LLM chain-of-thought reasoning, but for
 * code space exploration. Each "thought" is a dimension transformation
 * with explicit reasoning.
 *
 * Key innovation: Alien pattern generation through multi-step reasoning
 * that explores non-intuitive regions of code space.
 */

import (
	"fmt"
	"math"

	"github.com/bishopfox/sliver/server/generate/morpher"
)

// ThoughtVector represents a single reasoning step in the transformation
// Novel: Explicit reasoning (like LLM chain-of-thought)
type ThoughtVector struct {
	// Reasoning
	Reasoning   string
	ThoughtType ThoughtType

	// Transformation
	SourceDim    int
	CurrentValue float64
	TargetValue  float64

	// Metadata
	Priority       float64 // How important this transformation is
	ConstraintSafe bool    // Whether it satisfies constraints
	Impact         float64 // Expected impact on distance
}

// ThoughtType categorizes the reasoning
type ThoughtType int

const (
	ThoughtDivergence       ThoughtType = iota // Move away from history centroid
	ThoughtExploration                         // Explore random dimension
	ThoughtBalancing                           // Balance related dimensions
	ThoughtOptimization                        // Optimize for entropy/novelty
	ThoughtConstraintRepair                    // Fix constraint violations
)

// ChainOfThought represents a complete transformation reasoning chain
// Novel: Multi-step reasoning process
type ChainOfThought struct {
	Thoughts        []ThoughtVector
	SourceVector    *CodeVector
	TargetVector    *CodeVector
	TotalDistance   float64
	ConstraintsSafe bool
	Reasoning       string // High-level explanation
}

// GenerateChainOfThought creates a reasoning chain for vector transformation
// Novel: Multi-objective reasoning with constraint awareness
func GenerateChainOfThought(
	current *CodeVector,
	history *VectorHistory,
	manifold *ExecutableManifold,
	rng *morpher.XorShift128,
) *ChainOfThought {

	cot := &ChainOfThought{
		Thoughts:     make([]ThoughtVector, 0),
		SourceVector: current.Clone(),
		TargetVector: NewCodeVector(),
	}

	// High-level reasoning
	if len(history.Vectors) == 0 {
		cot.Reasoning = "No history - exploring initial random region"
	} else {
		avgDist := history.AverageDistance()
		cot.Reasoning = fmt.Sprintf("History cluster at distance %.2f - targeting alien region", avgDist)
	}

	// Step 1: Divergence thoughts (move away from centroid)
	if len(history.Vectors) > 0 {
		divergenceThoughts := generateDivergenceThoughts(current, history.Centroid, rng)
		cot.Thoughts = append(cot.Thoughts, divergenceThoughts...)
	}

	// Step 2: Exploration thoughts (random dimension jumps for alien-ness)
	explorationThoughts := generateExplorationThoughts(current, rng, 5)
	cot.Thoughts = append(cot.Thoughts, explorationThoughts...)

	// Step 3: Optimization thoughts (maximize entropy, variety)
	optimizationThoughts := generateOptimizationThoughts(current, rng)
	cot.Thoughts = append(cot.Thoughts, optimizationThoughts...)

	// Step 4: Build target vector from thoughts
	cot.TargetVector = cot.executeThoughts(current)

	// Step 5: Project onto manifold (CRITICAL - enforce constraints!)
	cot.TargetVector = manifold.Project(cot.TargetVector)
	cot.ConstraintsSafe = manifold.IsValid(cot.TargetVector)

	// Step 6: Add repair thoughts if constraints violated
	if !cot.ConstraintsSafe {
		repairThoughts := generateRepairThoughts(cot.TargetVector, manifold, rng)
		cot.Thoughts = append(cot.Thoughts, repairThoughts...)
		cot.TargetVector = manifold.Project(cot.TargetVector)
	}

	// Calculate total transformation distance
	cot.TotalDistance = WeightedEuclideanDistance(cot.SourceVector, cot.TargetVector)

	return cot
}

// generateDivergenceThoughts creates thoughts that move away from centroid
// Novel: Systematic divergence from historical patterns
func generateDivergenceThoughts(current, centroid *CodeVector, rng *morpher.XorShift128) []ThoughtVector {
	thoughts := make([]ThoughtVector, 0)

	// For high-priority dimensions (opcodes, entropy), diverge significantly
	highPriorityDims := []int{
		DimNOPFrequency, DimMOVFrequency, DimByteEntropy, DimOpcodeEntropy,
		DimControlFlowDensity, DimStackOpDensity,
	}

	for _, dim := range highPriorityDims {
		centroidValue := centroid.Values[dim]
		currentValue := current.Values[dim]

		// Move away from centroid
		var targetValue float64
		if centroidValue < 0.5 {
			// Centroid is low → push high
			targetValue = 0.6 + rng.Float64()*0.4 // [0.6, 1.0]
		} else {
			// Centroid is high → push low
			targetValue = rng.Float64() * 0.4 // [0.0, 0.4]
		}

		thought := ThoughtVector{
			Reasoning:    fmt.Sprintf("Centroid at %.2f in dim %d, moving to %.2f", centroidValue, dim, targetValue),
			ThoughtType:  ThoughtDivergence,
			SourceDim:    dim,
			CurrentValue: currentValue,
			TargetValue:  targetValue,
			Priority:     0.9, // High priority
			Impact:       math.Abs(targetValue - centroidValue),
		}

		thoughts = append(thoughts, thought)
	}

	return thoughts
}

// generateExplorationThoughts creates random jumps for alien-ness
// Novel: Your "alien pattern" concept - non-intuitive exploration
func generateExplorationThoughts(current *CodeVector, rng *morpher.XorShift128, count int) []ThoughtVector {
	thoughts := make([]ThoughtVector, 0)

	for i := 0; i < count; i++ {
		// Pick random dimension
		dim := rng.Intn(TotalDimensions)

		// Jump to random value (alien!)
		targetValue := rng.Float64()

		thought := ThoughtVector{
			Reasoning:    fmt.Sprintf("Random exploration in dim %d", dim),
			ThoughtType:  ThoughtExploration,
			SourceDim:    dim,
			CurrentValue: current.Values[dim],
			TargetValue:  targetValue,
			Priority:     0.5, // Medium priority
			Impact:       math.Abs(targetValue - current.Values[dim]),
		}

		thoughts = append(thoughts, thought)
	}

	return thoughts
}

// generateOptimizationThoughts creates thoughts to maximize quality metrics
// Novel: Target high entropy, high variety (good evasion properties)
func generateOptimizationThoughts(current *CodeVector, rng *morpher.XorShift128) []ThoughtVector {
	thoughts := make([]ThoughtVector, 0)

	// Thought: Maximize byte entropy
	if current.Values[DimByteEntropy] < 0.9 {
		thought := ThoughtVector{
			Reasoning:    "Low byte entropy detected, increasing for better evasion",
			ThoughtType:  ThoughtOptimization,
			SourceDim:    DimByteEntropy,
			CurrentValue: current.Values[DimByteEntropy],
			TargetValue:  0.9 + rng.Float64()*0.1, // Target [0.9, 1.0]
			Priority:     1.0,                     // Highest priority
			Impact:       0.9 - current.Values[DimByteEntropy],
		}
		thoughts = append(thoughts, thought)
	}

	// Thought: Increase opcode variety
	if current.Values[DimOpcodeEntropy] < 0.8 {
		thought := ThoughtVector{
			Reasoning:    "Low opcode variety, diversifying instruction mix",
			ThoughtType:  ThoughtOptimization,
			SourceDim:    DimOpcodeEntropy,
			CurrentValue: current.Values[DimOpcodeEntropy],
			TargetValue:  0.8 + rng.Float64()*0.2,
			Priority:     0.9,
			Impact:       0.8 - current.Values[DimOpcodeEntropy],
		}
		thoughts = append(thoughts, thought)
	}

	return thoughts
}

// generateRepairThoughts fixes constraint violations
// Novel: Automatic constraint satisfaction
func generateRepairThoughts(vector *CodeVector, manifold *ExecutableManifold, rng *morpher.XorShift128) []ThoughtVector {
	thoughts := make([]ThoughtVector, 0)

	// Check dependencies and repair if violated
	for _, dep := range manifold.Dependencies {
		sourceValue := vector.Values[dep.SourceDim]
		targetValue := vector.Values[dep.TargetDim]

		minTarget, maxTarget := dep.CalculateRange(sourceValue)

		if targetValue < minTarget || targetValue > maxTarget {
			// Violation! Create repair thought
			repairedValue := (minTarget + maxTarget) / 2.0 // Use midpoint

			thought := ThoughtVector{
				Reasoning:    fmt.Sprintf("Repairing dependency: dim %d must be in [%.2f, %.2f]", dep.TargetDim, minTarget, maxTarget),
				ThoughtType:  ThoughtConstraintRepair,
				SourceDim:    dep.TargetDim,
				CurrentValue: targetValue,
				TargetValue:  repairedValue,
				Priority:     1.0, // Critical priority
				Impact:       math.Abs(repairedValue - targetValue),
			}

			thoughts = append(thoughts, thought)
		}
	}

	return thoughts
}

// executeThoughts applies all thoughts to create target vector
// Novel: Sequential application with priority ordering
func (cot *ChainOfThought) executeThoughts(source *CodeVector) *CodeVector {
	target := source.Clone()

	// Sort thoughts by priority (highest first)
	sortThoughtsByPriority(cot.Thoughts)

	// Apply each thought
	for _, thought := range cot.Thoughts {
		target.Values[thought.SourceDim] = thought.TargetValue
	}

	return target
}

// sortThoughtsByPriority sorts thoughts by priority (descending)
func sortThoughtsByPriority(thoughts []ThoughtVector) {
	// Simple bubble sort (small arrays)
	n := len(thoughts)
	for i := 0; i < n; i++ {
		for j := 0; j < n-i-1; j++ {
			if thoughts[j].Priority < thoughts[j+1].Priority {
				thoughts[j], thoughts[j+1] = thoughts[j+1], thoughts[j]
			}
		}
	}
}

// Validate checks if chain of thought produces valid result
func (cot *ChainOfThought) Validate(manifold *ExecutableManifold) error {
	if cot.TargetVector == nil {
		return fmt.Errorf("no target vector")
	}

	if !manifold.IsValid(cot.TargetVector) {
		return fmt.Errorf("target vector violates constraints")
	}

	return nil
}

// GetSummary returns a human-readable summary of the reasoning chain
// Novel: Explainability (understand WHY vector changed)
func (cot *ChainOfThought) GetSummary() string {
	summary := fmt.Sprintf("Chain-of-Thought: %s\n", cot.Reasoning)
	summary += fmt.Sprintf("Total distance: %.2f\n", cot.TotalDistance)
	summary += fmt.Sprintf("Constraints safe: %v\n", cot.ConstraintsSafe)
	summary += fmt.Sprintf("Thoughts (%d):\n", len(cot.Thoughts))

	// Show top 5 thoughts by priority
	count := len(cot.Thoughts)
	if count > 5 {
		count = 5
	}

	for i := 0; i < count; i++ {
		thought := cot.Thoughts[i]
		summary += fmt.Sprintf("  %d. [Priority %.2f] %s\n", i+1, thought.Priority, thought.Reasoning)
		summary += fmt.Sprintf("     Dim %d: %.2f → %.2f (impact: %.2f)\n",
			thought.SourceDim, thought.CurrentValue, thought.TargetValue, thought.Impact)
	}

	return summary
}
