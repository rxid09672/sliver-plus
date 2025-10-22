package wig

/*
 * Code Vector Space Representation
 *
 * Novel approach: Represent code as points in high-dimensional vector space,
 * similar to LLM embeddings. Each dimension captures a different aspect of
 * code structure, enabling exploration of "alien" patterns that are non-intuitive
 * to human defenders.
 *
 * Key innovation: Constrained vector space where every point represents
 * valid, executable code (the "Executable Manifold").
 *
 * Dimensions designed to:
 * - Capture code characteristics
 * - Enable distance-based optimization
 * - Support chain-of-thought transformations
 * - Generate alien patterns
 */

import (
	"fmt"
	"math"

	"github.com/bishopfox/sliver/server/generate/lito"
)

// Dimension indices (for clarity and type safety)
// Novel: Named dimensions instead of magic indices
const (
	// Opcode distribution dimensions (0-19)
	DimNOPFrequency  = 0
	DimMOVFrequency  = 1
	DimPUSHFrequency = 2
	DimPOPFrequency  = 3
	DimADDFrequency  = 4
	DimSUBFrequency  = 5
	DimXORFrequency  = 6
	DimTESTFrequency = 7
	DimCMPFrequency  = 8
	DimLEAFrequency  = 9
	DimJMPFrequency  = 10
	DimCALLFrequency = 11
	DimJCCFrequency  = 12 // Conditional jumps
	DimXCHGFrequency = 13
	DimIMULFrequency = 14
	// ... reserve 0-19 for opcode frequencies

	// Length distribution dimensions (20-29)
	DimLength1Byte     = 20
	DimLength2Byte     = 21
	DimLength3Byte     = 22
	DimLength4to6Byte  = 23
	DimLength7PlusByte = 24
	DimAvgLength       = 25

	// Complexity dimensions (30-34)
	DimSimpleInstr       = 30
	DimMediumInstr       = 31
	DimComplexInstr      = 32
	DimOverallComplexity = 33

	// Register usage dimensions (35-42)
	DimEAXUsage = 35
	DimECXUsage = 36
	DimEDXUsage = 37
	DimEBXUsage = 38
	DimESPUsage = 39
	DimEBPUsage = 40
	DimESIUsage = 41
	DimEDIUsage = 42

	// Structural dimensions (43-62)
	DimControlFlowDensity = 43
	DimStackOpDensity     = 44
	DimArithmeticDensity  = 45
	DimLogicOpDensity     = 46
	DimDataMoveDensity    = 47
	DimBasicBlockSize     = 48
	DimJumpDistance       = 49

	// Statistical dimensions (63-72)
	DimByteEntropy     = 63
	DimOpcodeEntropy   = 64
	DimSequenceEntropy = 65
	DimRegisterEntropy = 66

	// Sequence pattern dimensions (73-99)
	DimBigramVariety  = 73
	DimTrigramVariety = 74

	// Total dimensions
	TotalDimensions = 100
)

// CodeVector represents code as a point in high-dimensional space
// Novel: Comprehensive code characterization
type CodeVector struct {
	Values [TotalDimensions]float64

	// Metadata
	SourceCode []byte // Original code (optional, for debugging)
	Valid      bool   // Whether this vector is valid
}

// NewCodeVector creates a zero-initialized vector
func NewCodeVector() *CodeVector {
	return &CodeVector{
		Values: [TotalDimensions]float64{},
		Valid:  false,
	}
}

// ExtractCodeVector analyzes code and extracts its vector representation
// Novel: Multi-dimensional feature extraction
func ExtractCodeVector(code []byte, mode64 bool) (*CodeVector, error) {
	if len(code) == 0 {
		return nil, fmt.Errorf("empty code")
	}

	vector := NewCodeVector()

	// Parse all instructions with Lito
	stream := lito.NewInstructionStream(code, mode64)
	if err := stream.ParseAll(); err != nil {
		return nil, fmt.Errorf("failed to parse code: %w", err)
	}

	if len(stream.Instructions) == 0 {
		return nil, fmt.Errorf("no instructions found")
	}

	totalInstrs := float64(len(stream.Instructions))

	// Extract opcode frequencies (dimensions 0-19)
	opcodeCounts := make(map[byte]int)
	for _, instr := range stream.Instructions {
		opcodeCounts[instr.Opcode]++
	}

	// Normalize to frequencies
	for opcode, count := range opcodeCounts {
		freq := float64(count) / totalInstrs

		// Map specific opcodes to dimensions
		switch opcode {
		case 0x90:
			vector.Values[DimNOPFrequency] = freq
		case 0x89, 0x8B:
			vector.Values[DimMOVFrequency] += freq
		case 0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57:
			vector.Values[DimPUSHFrequency] += freq
		case 0x58, 0x59, 0x5A, 0x5B, 0x5C, 0x5D, 0x5E, 0x5F:
			vector.Values[DimPOPFrequency] += freq
		case 0x01, 0x03, 0x05, 0x83: // ADD variants
			vector.Values[DimADDFrequency] += freq
		case 0x29, 0x2B, 0x2D: // SUB variants
			vector.Values[DimSUBFrequency] += freq
		case 0x31, 0x33, 0x35: // XOR variants
			vector.Values[DimXORFrequency] += freq
		case 0x85:
			vector.Values[DimTESTFrequency] = freq
		case 0x39, 0x3B, 0x3D:
			vector.Values[DimCMPFrequency] += freq
		case 0x8D:
			vector.Values[DimLEAFrequency] = freq
		}

		// Conditional jumps (0x70-0x7F for single-byte)
		if opcode >= 0x70 && opcode <= 0x7F {
			vector.Values[DimJCCFrequency] += freq
		}
	}

	// Check two-byte conditional jumps (0x0F 0x80-0x8F)
	for _, instr := range stream.Instructions {
		if instr.Properties.IsTwoByteOpcode && instr.Opcode2 >= 0x80 && instr.Opcode2 <= 0x8F {
			vector.Values[DimJCCFrequency] += 1.0 / totalInstrs
		}
	}

	// Extract length distribution (dimensions 20-29)
	lengthCounts := make(map[int]int)
	totalLength := 0
	for _, instr := range stream.Instructions {
		length := int(instr.Length)
		lengthCounts[length]++
		totalLength += length
	}

	for length, count := range lengthCounts {
		freq := float64(count) / totalInstrs

		switch {
		case length == 1:
			vector.Values[DimLength1Byte] = freq
		case length == 2:
			vector.Values[DimLength2Byte] = freq
		case length == 3:
			vector.Values[DimLength3Byte] = freq
		case length >= 4 && length <= 6:
			vector.Values[DimLength4to6Byte] += freq
		case length >= 7:
			vector.Values[DimLength7PlusByte] += freq
		}
	}

	vector.Values[DimAvgLength] = float64(totalLength) / totalInstrs / 15.0 // Normalize to [0,1]

	// Extract complexity (dimensions 30-34)
	simpleCount := 0
	mediumCount := 0
	complexCount := 0

	for _, instr := range stream.Instructions {
		complexity := estimateComplexity(instr)
		switch complexity {
		case 1:
			simpleCount++
		case 2:
			mediumCount++
		case 3:
			complexCount++
		}
	}

	vector.Values[DimSimpleInstr] = float64(simpleCount) / totalInstrs
	vector.Values[DimMediumInstr] = float64(mediumCount) / totalInstrs
	vector.Values[DimComplexInstr] = float64(complexCount) / totalInstrs
	vector.Values[DimOverallComplexity] = (float64(mediumCount) + 2*float64(complexCount)) / totalInstrs

	// Extract register usage (dimensions 35-42)
	registerCounts := make(map[int]int)
	for _, instr := range stream.Instructions {
		if instr.Properties.HasModRM {
			reg := int((instr.ModRM >> 3) & 0x07)
			rm := int(instr.ModRM & 0x07)
			registerCounts[reg]++
			registerCounts[rm]++
		}
	}

	totalRegRefs := 0
	for _, count := range registerCounts {
		totalRegRefs += count
	}

	if totalRegRefs > 0 {
		vector.Values[DimEAXUsage] = float64(registerCounts[0]) / float64(totalRegRefs)
		vector.Values[DimECXUsage] = float64(registerCounts[1]) / float64(totalRegRefs)
		vector.Values[DimEDXUsage] = float64(registerCounts[2]) / float64(totalRegRefs)
		vector.Values[DimEBXUsage] = float64(registerCounts[3]) / float64(totalRegRefs)
		vector.Values[DimESPUsage] = float64(registerCounts[4]) / float64(totalRegRefs)
		vector.Values[DimEBPUsage] = float64(registerCounts[5]) / float64(totalRegRefs)
		vector.Values[DimESIUsage] = float64(registerCounts[6]) / float64(totalRegRefs)
		vector.Values[DimEDIUsage] = float64(registerCounts[7]) / float64(totalRegRefs)
	}

	// Extract structural metrics (dimensions 43-62)
	controlFlowCount := 0
	stackOpCount := 0
	arithmeticCount := 0

	for _, instr := range stream.Instructions {
		if instr.IsControlFlow() {
			controlFlowCount++
		}

		if instr.Opcode >= 0x50 && instr.Opcode <= 0x5F {
			stackOpCount++
		}

		if instr.Opcode >= 0x00 && instr.Opcode <= 0x3F {
			arithmeticCount++
		}
	}

	codeLength := float64(len(code))
	vector.Values[DimControlFlowDensity] = float64(controlFlowCount) / (codeLength / 100.0)
	vector.Values[DimStackOpDensity] = float64(stackOpCount) / (codeLength / 100.0)
	vector.Values[DimArithmeticDensity] = float64(arithmeticCount) / (codeLength / 100.0)

	// Normalize densities to [0, 1]
	vector.Values[DimControlFlowDensity] = math.Min(vector.Values[DimControlFlowDensity]/10.0, 1.0)
	vector.Values[DimStackOpDensity] = math.Min(vector.Values[DimStackOpDensity]/10.0, 1.0)
	vector.Values[DimArithmeticDensity] = math.Min(vector.Values[DimArithmeticDensity]/10.0, 1.0)

	// Calculate entropy (dimensions 63-72)
	vector.Values[DimByteEntropy] = calculateByteEntropy(code) / 8.0 // Normalize to [0,1]
	vector.Values[DimOpcodeEntropy] = calculateOpcodeEntropy(stream.Instructions) / 8.0

	vector.Valid = true
	vector.SourceCode = code

	return vector, nil
}

// estimateComplexity estimates instruction complexity
// Novel: Complexity metric for dimension extraction
func estimateComplexity(instr *lito.Instruction) int {
	length := int(instr.Length)

	if length == 1 {
		return 1 // Simple
	} else if length <= 3 {
		if instr.Properties.HasModRM {
			return 2 // Medium
		}
		return 1
	} else {
		return 3 // Complex
	}
}

// calculateByteEntropy calculates Shannon entropy of bytes
// Novel: Information theory metric
func calculateByteEntropy(code []byte) float64 {
	if len(code) == 0 {
		return 0
	}

	// Count byte frequencies
	freq := make(map[byte]int)
	for _, b := range code {
		freq[b]++
	}

	// Calculate Shannon entropy
	entropy := 0.0
	total := float64(len(code))

	for _, count := range freq {
		p := float64(count) / total
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}

	return entropy
}

// calculateOpcodeEntropy calculates entropy of opcode distribution
func calculateOpcodeEntropy(instructions []*lito.Instruction) float64 {
	if len(instructions) == 0 {
		return 0
	}

	freq := make(map[byte]int)
	for _, instr := range instructions {
		freq[instr.Opcode]++
	}

	entropy := 0.0
	total := float64(len(instructions))

	for _, count := range freq {
		p := float64(count) / total
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}

	return entropy
}

// GetDimension safely retrieves a dimension value
// Novel: Bounds checking
func (cv *CodeVector) GetDimension(dim int) float64 {
	if dim < 0 || dim >= TotalDimensions {
		return 0.0
	}
	return cv.Values[dim]
}

// SetDimension safely sets a dimension value
func (cv *CodeVector) SetDimension(dim int, value float64) {
	if dim >= 0 && dim < TotalDimensions {
		// Clamp to [0, 1]
		cv.Values[dim] = math.Max(0.0, math.Min(1.0, value))
	}
}

// ToArray converts vector to flat array
func (cv *CodeVector) ToArray() []float64 {
	result := make([]float64, TotalDimensions)
	copy(result, cv.Values[:])
	return result
}

// FromArray loads vector from flat array
func (cv *CodeVector) FromArray(values []float64) {
	length := len(values)
	if length > TotalDimensions {
		length = TotalDimensions
	}
	copy(cv.Values[:length], values)
	cv.Valid = true
}

// Clone creates an independent copy
func (cv *CodeVector) Clone() *CodeVector {
	clone := NewCodeVector()
	clone.Values = cv.Values
	clone.Valid = cv.Valid
	// Note: Don't copy SourceCode (can be large)
	return clone
}

// String returns a human-readable representation
// Novel: Debugging helper
func (cv *CodeVector) String() string {
	return fmt.Sprintf("CodeVector{NOP:%.2f MOV:%.2f PUSH:%.2f POP:%.2f Entropy:%.2f Valid:%v}",
		cv.Values[DimNOPFrequency],
		cv.Values[DimMOVFrequency],
		cv.Values[DimPUSHFrequency],
		cv.Values[DimPOPFrequency],
		cv.Values[DimByteEntropy],
		cv.Valid,
	)
}

// GetSignature returns a compact representation of key dimensions
// Novel: For quick comparison/logging
func (cv *CodeVector) GetSignature() string {
	return fmt.Sprintf("[%.2f,%.2f,%.2f,%.2f,%.2f]",
		cv.Values[DimNOPFrequency],
		cv.Values[DimControlFlowDensity],
		cv.Values[DimStackOpDensity],
		cv.Values[DimByteEntropy],
		cv.Values[DimOverallComplexity],
	)
}
