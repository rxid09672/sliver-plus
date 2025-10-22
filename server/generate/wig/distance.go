package wig

/*
 * Vector Space Distance Metrics & History Analysis
 *
 * Implements distance calculations in 100D code vector space to:
 * - Measure similarity between builds
 * - Find "alien" regions (far from history)
 * - Cluster analysis (where have we been?)
 * - Novelty scoring (how different is this?)
 *
 * Novel approach:
 * - Multiple distance metrics (Euclidean, Cosine, Manhattan)
 * - Weighted dimensions (some more important than others)
 * - Centroid tracking (average of history)
 * - Outlier detection (find unexplored regions)
 */

import (
	"math"

	"github.com/bishopfox/sliver/server/generate/morpher"
)

// DistanceMetric defines how to measure distance in vector space
type DistanceMetric int

const (
	MetricEuclidean DistanceMetric = iota // Standard L2 distance
	MetricCosine                          // Angle-based similarity
	MetricManhattan                       // L1 distance (sum of absolute differences)
	MetricWeighted                        // Weighted Euclidean (important dims weighted more)
)

// DimensionWeight assigns importance to dimensions
// Novel: Not all dimensions equally important for diversity
var dimensionWeights = [TotalDimensions]float64{
	// Opcode frequencies (0-19): HIGH importance (visible signatures)
	1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0,
	1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0,

	// Length distribution (20-29): MEDIUM importance
	0.7, 0.7, 0.7, 0.7, 0.7, 0.7, 0.7, 0.7, 0.7, 0.7,

	// Complexity (30-34): MEDIUM importance
	0.6, 0.6, 0.6, 0.6, 0.6,

	// Register usage (35-42): LOW importance (less visible)
	0.4, 0.4, 0.4, 0.4, 0.4, 0.4, 0.4, 0.4,

	// Structural (43-62): HIGH importance (behavioral signatures)
	0.9, 0.9, 0.9, 0.9, 0.9, 0.8, 0.8, 0.8, 0.8, 0.8,
	0.8, 0.8, 0.8, 0.8, 0.8, 0.8, 0.8, 0.8, 0.8, 0.8,

	// Statistical (63-99): VERY HIGH importance (entropy = key evasion metric)
	1.2, 1.2, 1.2, 1.2, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0,
	// ... remaining dimensions default weight
}

// EuclideanDistance calculates L2 distance between vectors
// Novel: Standard distance metric in vector space
func EuclideanDistance(v1, v2 *CodeVector) float64 {
	sum := 0.0
	for i := 0; i < TotalDimensions; i++ {
		diff := v1.Values[i] - v2.Values[i]
		sum += diff * diff
	}
	return math.Sqrt(sum)
}

// WeightedEuclideanDistance calculates weighted L2 distance
// Novel: Important dimensions (entropy, opcodes) weighted more heavily
func WeightedEuclideanDistance(v1, v2 *CodeVector) float64 {
	sum := 0.0
	for i := 0; i < TotalDimensions; i++ {
		diff := v1.Values[i] - v2.Values[i]
		weight := dimensionWeights[i]
		sum += weight * diff * diff
	}
	return math.Sqrt(sum)
}

// CosineSimilarity calculates cosine similarity (angle between vectors)
// Novel: 0 = orthogonal (very different), 1 = identical, -1 = opposite
func CosineSimilarity(v1, v2 *CodeVector) float64 {
	dotProduct := 0.0
	norm1 := 0.0
	norm2 := 0.0

	for i := 0; i < TotalDimensions; i++ {
		dotProduct += v1.Values[i] * v2.Values[i]
		norm1 += v1.Values[i] * v1.Values[i]
		norm2 += v2.Values[i] * v2.Values[i]
	}

	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// ManhattanDistance calculates L1 distance (sum of absolute differences)
// Novel: Alternative metric, sometimes more intuitive
func ManhattanDistance(v1, v2 *CodeVector) float64 {
	sum := 0.0
	for i := 0; i < TotalDimensions; i++ {
		sum += math.Abs(v1.Values[i] - v2.Values[i])
	}
	return sum
}

// Distance calculates distance using specified metric
func Distance(v1, v2 *CodeVector, metric DistanceMetric) float64 {
	switch metric {
	case MetricEuclidean:
		return EuclideanDistance(v1, v2)
	case MetricCosine:
		return 1.0 - CosineSimilarity(v1, v2) // Convert similarity to distance
	case MetricManhattan:
		return ManhattanDistance(v1, v2)
	case MetricWeighted:
		return WeightedEuclideanDistance(v1, v2)
	default:
		return EuclideanDistance(v1, v2)
	}
}

// CalculateCentroid finds the center of mass of a set of vectors
// Novel: Represents "average" code pattern from history
func CalculateCentroid(vectors []*CodeVector) *CodeVector {
	if len(vectors) == 0 {
		return NewCodeVector()
	}

	centroid := NewCodeVector()

	// Sum all vectors
	for _, v := range vectors {
		for i := 0; i < TotalDimensions; i++ {
			centroid.Values[i] += v.Values[i]
		}
	}

	// Average
	count := float64(len(vectors))
	for i := 0; i < TotalDimensions; i++ {
		centroid.Values[i] /= count
	}

	centroid.Valid = true
	return centroid
}

// FindAlienRegion identifies a region of vector space far from history
// Novel: Generates truly novel patterns by maximizing distance
func FindAlienRegion(history []*CodeVector, rng *morpher.XorShift128) *CodeVector {
	if len(history) == 0 {
		// No history - return random valid vector
		return RandomValidVector(rng)
	}

	// Calculate centroid of history
	centroid := CalculateCentroid(history)

	// Generate candidate alien vectors
	bestAlien := NewCodeVector()
	bestDistance := 0.0

	// Try multiple candidates, keep the most distant
	// Novel: Monte Carlo search for alien regions
	for attempt := 0; attempt < 20; attempt++ {
		candidate := NewCodeVector()

		// For each dimension, move AWAY from centroid
		for dim := 0; dim < TotalDimensions; dim++ {
			centroidValue := centroid.Values[dim]

			// Generate value on opposite side of centroid
			if centroidValue < 0.5 {
				// Centroid is low → go high
				candidate.Values[dim] = 0.5 + rng.Float64()*0.5 // [0.5, 1.0]
			} else {
				// Centroid is high → go low
				candidate.Values[dim] = rng.Float64() * 0.5 // [0.0, 0.5]
			}

			// Add some randomness for true alien-ness
			if rng.Float64() < 0.2 { // 20% chance
				candidate.Values[dim] = rng.Float64() // Completely random
			}
		}

		// Calculate distance from centroid
		distance := WeightedEuclideanDistance(candidate, centroid)

		if distance > bestDistance {
			bestDistance = distance
			bestAlien = candidate
		}
	}

	bestAlien.Valid = true
	return bestAlien
}

// RandomValidVector generates a random vector
func RandomValidVector(rng *morpher.XorShift128) *CodeVector {
	vector := NewCodeVector()

	for i := 0; i < TotalDimensions; i++ {
		vector.Values[i] = rng.Float64()
	}

	vector.Valid = true
	return vector
}

// VectorHistory tracks previous builds in vector space
// Novel: Maintains exploration history for novelty optimization
type VectorHistory struct {
	Vectors    []*CodeVector
	MaxHistory int
	Centroid   *CodeVector
	Metric     DistanceMetric
}

// NewVectorHistory creates a new history tracker
func NewVectorHistory(maxHistory int) *VectorHistory {
	return &VectorHistory{
		Vectors:    make([]*CodeVector, 0, maxHistory),
		MaxHistory: maxHistory,
		Centroid:   NewCodeVector(),
		Metric:     MetricWeighted, // Use weighted by default
	}
}

// Add adds a vector to history
func (vh *VectorHistory) Add(vector *CodeVector) {
	vh.Vectors = append(vh.Vectors, vector.Clone())

	// Maintain max history size (sliding window)
	if len(vh.Vectors) > vh.MaxHistory {
		vh.Vectors = vh.Vectors[1:] // Remove oldest
	}

	// Update centroid
	vh.Centroid = CalculateCentroid(vh.Vectors)
}

// AverageDistance calculates average distance from centroid
// Novel: Measures clustering tightness
func (vh *VectorHistory) AverageDistance() float64 {
	if len(vh.Vectors) == 0 {
		return 0
	}

	total := 0.0
	for _, v := range vh.Vectors {
		total += Distance(v, vh.Centroid, vh.Metric)
	}

	return total / float64(len(vh.Vectors))
}

// MinDistanceFromHistory calculates minimum distance to any historical vector
// Novel: Novelty score - higher = more novel
func (vh *VectorHistory) MinDistanceFromHistory(vector *CodeVector) float64 {
	if len(vh.Vectors) == 0 {
		return math.MaxFloat64 // Infinitely novel (no history)
	}

	minDist := math.MaxFloat64

	for _, historical := range vh.Vectors {
		dist := Distance(vector, historical, vh.Metric)
		if dist < minDist {
			minDist = dist
		}
	}

	return minDist
}

// NoveltyScore calculates how novel a vector is
// Novel: Higher score = more different from history
func (vh *VectorHistory) NoveltyScore(vector *CodeVector) float64 {
	if len(vh.Vectors) == 0 {
		return 1.0 // Maximally novel
	}

	// Distance from centroid (normalized)
	centroidDist := Distance(vector, vh.Centroid, vh.Metric)

	// Minimum distance from any historical point
	minHistDist := vh.MinDistanceFromHistory(vector)

	// Combined novelty score
	// Novel: Weighted combination of both metrics
	novelty := (centroidDist * 0.6) + (minHistDist * 0.4)

	// Normalize to [0, 1]
	// Max possible distance in [0,1]^100 space is sqrt(100) ≈ 10
	novelty = math.Min(novelty/10.0, 1.0)

	return novelty
}

// GetClusterRadius calculates the radius of the historical cluster
// Novel: Measure exploration spread
func (vh *VectorHistory) GetClusterRadius() float64 {
	if len(vh.Vectors) < 2 {
		return 0
	}

	maxDist := 0.0

	for _, v := range vh.Vectors {
		dist := Distance(v, vh.Centroid, vh.Metric)
		if dist > maxDist {
			maxDist = dist
		}
	}

	return maxDist
}

// IsInCluster checks if a vector is within historical cluster
func (vh *VectorHistory) IsInCluster(vector *CodeVector, radiusMultiplier float64) bool {
	radius := vh.GetClusterRadius() * radiusMultiplier
	dist := Distance(vector, vh.Centroid, vh.Metric)
	return dist <= radius
}

// FindOutlierRegion finds a region far from historical cluster
// Novel: Systematic alien region discovery
func (vh *VectorHistory) FindOutlierRegion(rng *morpher.XorShift128, minDistance float64) *CodeVector {
	maxAttempts := 50

	for attempt := 0; attempt < maxAttempts; attempt++ {
		candidate := FindAlienRegion(vh.Vectors, rng)

		// Check if it's far enough
		novelty := vh.NoveltyScore(candidate)

		if novelty >= minDistance {
			return candidate
		}
	}

	// Fallback: return best candidate found
	return FindAlienRegion(vh.Vectors, rng)
}
