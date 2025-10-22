package wig

/*
 * Vector Space Wig - Main Engine
 *
 * Orchestrates multi-dimensional code space exploration with chain-of-thought
 * reasoning to generate alien patterns that maximize distance from historical
 * builds while maintaining executable code guarantees.
 *
 * Novel approach: LLM-inspired vector space + chain-of-thought + genetic lottery
 *
 * Flow:
 * 1. Extract vector from current code
 * 2. Generate chain-of-thought for transformation
 * 3. Find alien target in vector space
 * 4. Project onto executable manifold (safety)
 * 5. Convert to Morpher config
 * 6. Generate morphed code
 * 7. Validate and add to history
 *
 * Key innovation: Systematic exploration of non-intuitive code patterns
 * that are foreign to human defenders but mathematically guaranteed valid.
 */

import (
	"fmt"

	"github.com/bishopfox/sliver/server/generate/morpher"
	"github.com/bishopfox/sliver/server/log"
)

var (
	wigLog = log.NamedLogger("generate", "wig")
)

// VectorSpaceResult wraps the result of vector space morphing
// Novel: Comprehensive result with vector space metadata
type VectorSpaceResult struct {
	*morpher.MorphResult
	NoveltyScore         float64
	ThoughtSteps         int
	AlienRegion          string
	ConstraintViolations int
	Seed                 uint32
}

// VectorSpaceWig is the main engine for vector-guided metamorphism
// Novel: Complete vector space exploration system
type VectorSpaceWig struct {
	History  *VectorHistory
	Manifold *ExecutableManifold
	RNG      *morpher.XorShift128

	// Configuration
	MinNovelty  float64 // Minimum novelty score (0.0-1.0)
	MaxAttempts int     // Max attempts to find valid alien vector

	// Statistics
	GenerationCount int
	TotalDistance   float64
	AvgNovelty      float64
}

// NewVectorSpaceWig creates a new vector space exploration engine
func NewVectorSpaceWig(seed uint32, mode64 bool, platform string) *VectorSpaceWig {
	if seed == 0 {
		seed = morpher.GetRDTSCSeedWithEntropy()
	}

	return &VectorSpaceWig{
		History:         NewVectorHistory(50), // Keep last 50 builds
		Manifold:        NewExecutableManifold(mode64, platform),
		RNG:             morpher.NewXorShift128(seed),
		MinNovelty:      0.3, // At least 30% novel
		MaxAttempts:     20,  // Try 20 times to find valid alien pattern
		GenerationCount: 0,
		TotalDistance:   0,
		AvgNovelty:      0,
	}
}

// GenerateAlienVector generates a novel vector far from history
// Novel: Chain-of-thought guided exploration with safety guarantees
func (vsw *VectorSpaceWig) GenerateAlienVector(currentCode []byte) (*CodeVector, *ChainOfThought, error) {
	// Extract current vector
	var currentVector *CodeVector
	var err error

	if len(currentCode) > 0 {
		currentVector, err = ExtractCodeVector(currentCode, vsw.Manifold.Mode64)
		if err != nil {
			// Can't extract - use default
			currentVector = NewCodeVector()
		}
	} else {
		currentVector = NewCodeVector()
	}

	// Try multiple times to find a valid alien vector
	for attempt := 0; attempt < vsw.MaxAttempts; attempt++ {
		// Generate chain of thought
		cot := GenerateChainOfThought(currentVector, vsw.History, vsw.Manifold, vsw.RNG)

		// Validate chain of thought
		if err := cot.Validate(vsw.Manifold); err != nil {
			wigLog.Debugf("Chain-of-thought validation failed (attempt %d): %v", attempt, err)
			continue
		}

		// Check novelty score
		novelty := vsw.History.NoveltyScore(cot.TargetVector)
		if novelty < vsw.MinNovelty {
			wigLog.Debugf("Novelty %.2f below threshold %.2f (attempt %d)", novelty, vsw.MinNovelty, attempt)
			continue
		}

		// Success! Found valid alien vector
		wigLog.Infof("Found alien vector with novelty %.2f (attempt %d)", novelty, attempt)
		return cot.TargetVector, cot, nil
	}

	// Failed to find - use fallback
	wigLog.Warnf("Failed to find alien vector after %d attempts, using fallback", vsw.MaxAttempts)
	fallback := vsw.History.FindOutlierRegion(vsw.RNG, vsw.MinNovelty)
	fallback = vsw.Manifold.Project(fallback)

	return fallback, nil, nil
}

// MorphWithVectorGuidance performs vector-guided morphing
// Novel: Complete pipeline from vector space to morphed code
func (vsw *VectorSpaceWig) MorphWithVectorGuidance(code []byte) (*morpher.MorphResult, *ChainOfThought, error) {
	// Generate alien vector
	alienVector, cot, err := vsw.GenerateAlienVector(code)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate alien vector: %w", err)
	}

	// Log chain of thought
	if cot != nil {
		wigLog.Debugf("Chain-of-Thought:\n%s", cot.GetSummary())
	}

	// Convert vector to Morpher config
	morphConfig := VectorToMorphConfig(alienVector, vsw.Manifold)

	// Create morpher with vector-guided config
	morpher := morpher.NewMorpher(morphConfig)

	// Morph the code
	result, err := morpher.Morph(code)
	if err != nil {
		return nil, cot, fmt.Errorf("morphing failed: %w", err)
	}

	if !result.Success {
		return nil, cot, fmt.Errorf("morphing unsuccessful: %v", result.Error)
	}

	// Extract vector from morphed code and add to history
	morphedVector, err := ExtractCodeVector(result.Code, vsw.Manifold.Mode64)
	if err == nil {
		vsw.History.Add(morphedVector)

		// Update statistics
		vsw.GenerationCount++
		distance := WeightedEuclideanDistance(alienVector, morphedVector)
		vsw.TotalDistance += distance

		novelty := vsw.History.NoveltyScore(morphedVector)
		vsw.AvgNovelty = (vsw.AvgNovelty*float64(vsw.GenerationCount-1) + novelty) / float64(vsw.GenerationCount)

		wigLog.Infof("Vector space stats:")
		wigLog.Infof("  Target-Actual distance: %.2f", distance)
		wigLog.Infof("  Novelty score: %.2f", novelty)
		wigLog.Infof("  History size: %d", len(vsw.History.Vectors))
		wigLog.Infof("  Avg novelty: %.2f", vsw.AvgNovelty)
	}

	return result, cot, nil
}

// GetVectorSpaceStats returns statistics about vector space exploration
// Novel: Metrics for understanding exploration behavior
type VectorSpaceStats struct {
	GenerationCount   int
	HistorySize       int
	AvgNovelty        float64
	ClusterRadius     float64
	TotalDistance     float64
	CentroidSignature string
}

func (vsw *VectorSpaceWig) GetStats() *VectorSpaceStats {
	return &VectorSpaceStats{
		GenerationCount:   vsw.GenerationCount,
		HistorySize:       len(vsw.History.Vectors),
		AvgNovelty:        vsw.AvgNovelty,
		ClusterRadius:     vsw.History.GetClusterRadius(),
		TotalDistance:     vsw.TotalDistance,
		CentroidSignature: vsw.History.Centroid.GetSignature(),
	}
}

// Reset clears history and statistics
// Novel: Fresh start for new operation
func (vsw *VectorSpaceWig) Reset() {
	vsw.History = NewVectorHistory(50)
	vsw.GenerationCount = 0
	vsw.TotalDistance = 0
	vsw.AvgNovelty = 0
}

// ExportHistory exports vector history for analysis
// Novel: Forensics/debugging capability
func (vsw *VectorSpaceWig) ExportHistory() []*CodeVector {
	return vsw.History.Vectors
}

// ImportHistory imports vector history
// Novel: Resume from previous session
func (vsw *VectorSpaceWig) ImportHistory(vectors []*CodeVector) {
	vsw.History.Vectors = vectors
	vsw.History.Centroid = CalculateCentroid(vectors)
}

// MorphWithVectorGuidanceResult performs vector-guided morphing and returns VectorSpaceResult
// Novel: Wrapper for diversity integration
func (vsw *VectorSpaceWig) MorphWithVectorGuidanceResult(code []byte) (*VectorSpaceResult, error) {
	morphResult, cot, err := vsw.MorphWithVectorGuidance(code)
	if err != nil {
		return nil, err
	}

	// Calculate novelty score
	novelty := float64(0)
	if len(vsw.History.Vectors) > 1 {
		novelty = vsw.History.NoveltyScore(vsw.History.Vectors[len(vsw.History.Vectors)-1])
	}

	// Count constraint violations (simplified)
	violations := 0
	if cot != nil {
		for _, thought := range cot.Thoughts {
			if thought.ThoughtType == ThoughtConstraintRepair {
				violations++
			}
		}
	}

	// Determine alien region
	alienRegion := "unknown"
	if cot != nil && len(cot.Thoughts) > 0 {
		lastThought := cot.Thoughts[len(cot.Thoughts)-1]
		if lastThought.ThoughtType == ThoughtDivergence {
			alienRegion = "divergence"
		} else if lastThought.ThoughtType == ThoughtExploration {
			alienRegion = "exploration"
		}
	}

	return &VectorSpaceResult{
		MorphResult:          morphResult,
		NoveltyScore:         novelty,
		ThoughtSteps:         len(cot.Thoughts),
		AlienRegion:          alienRegion,
		ConstraintViolations: violations,
		Seed:                 vsw.RNG.SaveState()[0],
	}, nil
}
