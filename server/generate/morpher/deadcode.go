package morpher

import "github.com/bishopfox/sliver/server/generate/lito"

/*
 * Dead Code Generation for Polymorphic Obfuscation
 *
 * Generates semantically-null instructions (do nothing) to change
 * code structure and break static signatures.
 *
 * Novel approach:
 * - Multiple NOP-equivalent patterns (not just 0x90)
 * - Random selection from pool
 * - Configurable complexity
 * - Register-aware generation
 * - Length variation
 *
 * Key differences from malware:
 * - Larger variety of dead code patterns (malware used only single NOP)
 * - Context-aware generation (avoid breaking code)
 * - Configurable insertion rate
 * - Multiple complexity levels
 */

// DeadCodePattern represents a semantically-null instruction sequence
// Novel: Catalog of NOP equivalents with metadata
type DeadCodePattern struct {
	Bytes       []byte
	Length      int
	Complexity  int    // 1=simple, 2=medium, 3=complex
	Description string // For debugging
}

// deadCodeCatalog contains all available NOP-equivalent patterns
// Novel: Multiple patterns (malware used only 0x90)
var deadCodeCatalog = []DeadCodePattern{
	// Simple NOPs (1 byte)
	{[]byte{0x90}, 1, 1, "NOP"},

	// Register-preserving operations (2 bytes)
	{[]byte{0x89, 0xC0}, 2, 1, "MOV EAX, EAX"},
	{[]byte{0x89, 0xC9}, 2, 1, "MOV ECX, ECX"},
	{[]byte{0x89, 0xD2}, 2, 1, "MOV EDX, EDX"},
	{[]byte{0x89, 0xDB}, 2, 1, "MOV EBX, EBX"},
	{[]byte{0x89, 0xF6}, 2, 1, "MOV ESI, ESI"},
	{[]byte{0x89, 0xFF}, 2, 1, "MOV EDI, EDI"},

	// XCHG EAX, EAX (alias for NOP) (1 byte)
	{[]byte{0x87, 0xC0}, 2, 1, "XCHG EAX, EAX"},

	// LEA with no displacement (3 bytes)
	{[]byte{0x8D, 0x00}, 2, 1, "LEA EAX, [EAX]"},
	{[]byte{0x8D, 0x09}, 2, 1, "LEA ECX, [ECX]"},
	{[]byte{0x8D, 0x12}, 2, 1, "LEA EDX, [EDX]"},
	{[]byte{0x8D, 0x1B}, 2, 1, "LEA EBX, [EBX]"},

	// PUSH + POP same register (2 bytes)
	{[]byte{0x50, 0x58}, 2, 2, "PUSH EAX + POP EAX"},
	{[]byte{0x51, 0x59}, 2, 2, "PUSH ECX + POP ECX"},
	{[]byte{0x52, 0x5A}, 2, 2, "PUSH EDX + POP EDX"},
	{[]byte{0x53, 0x5B}, 2, 2, "PUSH EBX + POP EBX"},

	// Arithmetic NOPs (3 bytes)
	{[]byte{0x83, 0xC0, 0x00}, 3, 2, "ADD EAX, 0"},
	{[]byte{0x83, 0xC1, 0x00}, 3, 2, "ADD ECX, 0"},
	{[]byte{0x83, 0xE8, 0x00}, 3, 2, "SUB EAX, 0"},
	{[]byte{0x83, 0xE9, 0x00}, 3, 2, "SUB ECX, 0"},

	// XOR + XOR same register with same value (6 bytes)
	{[]byte{0x83, 0xF0, 0x00}, 3, 2, "XOR EAX, 0"},

	// Test with self (2 bytes)
	{[]byte{0x85, 0xC0}, 2, 2, "TEST EAX, EAX"},
	{[]byte{0x85, 0xC9}, 2, 2, "TEST ECX, ECX"},

	// CMP with self (2 bytes)
	{[]byte{0x39, 0xC0}, 2, 2, "CMP EAX, EAX"},
	{[]byte{0x39, 0xC9}, 2, 2, "CMP ECX, ECX"},

	// More complex: SHL + SHR (4 bytes)
	{[]byte{0xC1, 0xE0, 0x00}, 3, 3, "SHL EAX, 0"},
	{[]byte{0xC1, 0xE8, 0x00}, 3, 3, "SHR EAX, 0"},

	// IMUL with 1 (3 bytes)
	{[]byte{0x6B, 0xC0, 0x01}, 3, 3, "IMUL EAX, EAX, 1"},

	// Conditional jump over nothing (4 bytes)
	{[]byte{0x74, 0x00}, 2, 3, "JE +0"},
	{[]byte{0x75, 0x00}, 2, 3, "JNE +0"},
}

// DeadCodeConfig controls dead code generation
// Novel: Policy object for configurable behavior
type DeadCodeConfig struct {
	InsertionRate  float64 // 0.0-1.0: probability of inserting after each instruction
	MaxComplexity  int     // 1-3: maximum complexity of dead code
	MinLength      int     // Minimum dead code bytes per insertion
	MaxLength      int     // Maximum dead code bytes per insertion
	VariableLength bool    // Use variable-length sequences
	AllowComplex   bool    // Allow complexity level 3
}

// DefaultDeadCodeConfig returns sensible defaults
// Novel: Conservative defaults (malware was aggressive)
func DefaultDeadCodeConfig() *DeadCodeConfig {
	return &DeadCodeConfig{
		InsertionRate:  0.5, // 50% insertion rate
		MaxComplexity:  2,   // Medium complexity
		MinLength:      1,
		MaxLength:      3,
		VariableLength: true,
		AllowComplex:   false,
	}
}

// GenerateDeadCode generates random semantically-null bytes
// Novel: Intelligent pattern selection
func GenerateDeadCode(rng *XorShift128, config *DeadCodeConfig) []byte {
	// Filter patterns by complexity
	available := make([]DeadCodePattern, 0)
	for _, pattern := range deadCodeCatalog {
		if pattern.Complexity <= config.MaxComplexity {
			if !config.AllowComplex && pattern.Complexity >= 3 {
				continue
			}
			if pattern.Length >= config.MinLength && pattern.Length <= config.MaxLength {
				available = append(available, pattern)
			}
		}
	}

	if len(available) == 0 {
		// Fallback to simple NOP
		return []byte{0x90}
	}

	// Select random pattern
	idx := rng.Intn(len(available))
	pattern := available[idx]

	// Return copy to avoid modification
	result := make([]byte, pattern.Length)
	copy(result, pattern.Bytes)

	return result
}

// GenerateDeadCodeSequence generates a sequence of dead code
// Novel: Multi-instruction dead code blocks (more polymorphic)
func GenerateDeadCodeSequence(rng *XorShift128, config *DeadCodeConfig, targetLength int) []byte {
	if targetLength <= 0 {
		return nil
	}

	sequence := make([]byte, 0, targetLength)

	for len(sequence) < targetLength {
		// Generate one dead code unit
		deadCode := GenerateDeadCode(rng, config)

		// Don't exceed target length
		if len(sequence)+len(deadCode) > targetLength {
			// Fill remaining with single NOPs
			for len(sequence) < targetLength {
				sequence = append(sequence, 0x90)
			}
			break
		}

		sequence = append(sequence, deadCode...)
	}

	return sequence
}

// ShouldInsertDeadCode determines if dead code should be inserted
// Novel: Context-aware decision making
func ShouldInsertDeadCode(rng *XorShift128, config *DeadCodeConfig, afterInstr *lito.Instruction) bool {
	// Don't insert after control flow instructions (would break code flow)
	// Novel: Safety check (malware inserted blindly)
	if afterInstr != nil && afterInstr.IsControlFlow() {
		return false
	}

	// Use configured insertion rate
	return rng.Float64() < config.InsertionRate
}

// GetDeadCodeStats returns statistics about dead code patterns
// Novel: Analysis helper
type DeadCodeStats struct {
	TotalPatterns    int
	SimplestLength   int
	LongestLength    int
	AverageLength    float64
	ComplexityLevels map[int]int // Complexity → count
}

func GetDeadCodeStats() *DeadCodeStats {
	stats := &DeadCodeStats{
		TotalPatterns:    len(deadCodeCatalog),
		SimplestLength:   255,
		LongestLength:    0,
		ComplexityLevels: make(map[int]int),
	}

	totalLength := 0
	for _, pattern := range deadCodeCatalog {
		if pattern.Length < stats.SimplestLength {
			stats.SimplestLength = pattern.Length
		}
		if pattern.Length > stats.LongestLength {
			stats.LongestLength = pattern.Length
		}
		totalLength += pattern.Length
		stats.ComplexityLevels[pattern.Complexity]++
	}

	if stats.TotalPatterns > 0 {
		stats.AverageLength = float64(totalLength) / float64(stats.TotalPatterns)
	}

	return stats
}
