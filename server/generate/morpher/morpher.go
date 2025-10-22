package morpher

/*
 * Morpher - Metamorphic Code Transformation Engine
 *
 * Core two-pass algorithm for code mutation:
 * Pass 1: Parse, expand, inject dead code, track addresses
 * Pass 2: Relocate all jumps/calls to new addresses
 *
 * Novel approach:
 * - Clean two-pass architecture
 * - Configurable policies
 * - Comprehensive error handling
 * - Reproducible with seed
 * - Statistics and metrics
 *
 * Key differences from malware:
 * - Modern Go architecture
 * - Policy-driven (not hardcoded behavior)
 * - Safe fallbacks on errors
 * - Detailed logging/metrics
 * - Extensible design
 */

import (
	"fmt"

	"github.com/bishopfox/sliver/server/generate/lito"
)

// MorphConfig controls all aspects of code morphing
// Novel: Single config object (not scattered parameters)
type MorphConfig struct {
	Seed            uint32           // RNG seed (0 = auto)
	Mode64          bool             // x64 mode
	ExpansionPolicy *ExpansionPolicy // Instruction expansion rules
	DeadCodeConfig  *DeadCodeConfig  // Dead code insertion rules
	EnableExpansion bool             // Enable instruction expansion
	EnableDeadCode  bool             // Enable dead code injection
	PreserveLength  bool             // Try to keep similar length (less aggressive)
	MaxOutputSize   int              // Maximum output size (safety limit)
}

// DefaultMorphConfig returns sensible defaults
// Novel: Progressive enhancement approach
func DefaultMorphConfig() *MorphConfig {
	return &MorphConfig{
		Seed:            0, // Auto-seed with RDTSC
		Mode64:          false,
		ExpansionPolicy: DefaultExpansionPolicy(),
		DeadCodeConfig:  DefaultDeadCodeConfig(),
		EnableExpansion: true,
		EnableDeadCode:  true,
		PreserveLength:  false,
		MaxOutputSize:   10 * 1024 * 1024, // 10MB safety limit
	}
}

// MorphResult contains the morphed code and metadata
// Novel: Rich result object with statistics
type MorphResult struct {
	Code             []byte
	OriginalSize     int
	MorphedSize      int
	ExpansionRatio   float64
	InstructionCount int
	ExpandedCount    int
	DeadCodeBytes    int
	Seed             uint32
	Tracker          *AddressTracker
	Success          bool
	Error            error
}

// Morpher is the main morphing engine
// Novel: Stateful object for complex operations
type Morpher struct {
	config  *MorphConfig
	rng     *XorShift128
	tracker *AddressTracker
}

// NewMorpher creates a new morphing engine
func NewMorpher(config *MorphConfig) *Morpher {
	// Auto-seed if needed
	if config.Seed == 0 {
		config.Seed = GetRDTSCSeedWithEntropy()
	}

	return &Morpher{
		config:  config,
		rng:     NewXorShift128(config.Seed),
		tracker: NewAddressTracker(),
	}
}

// Morph performs the complete morphing operation
// Novel: Clean entry point with comprehensive result
func (m *Morpher) Morph(input []byte) (*MorphResult, error) {
	result := &MorphResult{
		OriginalSize: len(input),
		Seed:         m.config.Seed,
		Tracker:      m.tracker,
	}

	// Validate input
	if len(input) == 0 {
		result.Error = fmt.Errorf("empty input code")
		return result, result.Error
	}

	// Estimate output size and allocate buffer
	// Novel: Smart buffer allocation (avoid reallocations)
	estimatedSize := len(input) * 2 // Conservative estimate
	if estimatedSize > m.config.MaxOutputSize {
		result.Error = fmt.Errorf("estimated output size %d exceeds limit %d", estimatedSize, m.config.MaxOutputSize)
		return result, result.Error
	}

	output := make([]byte, 0, estimatedSize)

	// Pass 1: Morph code and build tracking table
	var err error
	output, err = m.pass1Morph(input, output)
	if err != nil {
		result.Error = fmt.Errorf("pass 1 failed: %w", err)
		return result, result.Error
	}

	// Pass 2: Fix relocations
	err = m.pass2Relocate(output)
	if err != nil {
		result.Error = fmt.Errorf("pass 2 failed: %w", err)
		return result, result.Error
	}

	// Populate result
	result.Code = output
	result.MorphedSize = len(output)
	result.Success = true

	// Calculate statistics
	stats := m.tracker.GetStats()
	result.InstructionCount = stats.TotalInstructions
	result.ExpandedCount = stats.ExpandedCount
	result.DeadCodeBytes = stats.TotalDeadCodeBytes
	result.ExpansionRatio = stats.ExpansionRatio

	return result, nil
}

// pass1Morph performs the first pass: parse, expand, inject dead code
// Novel: Clean separation of passes
func (m *Morpher) pass1Morph(input []byte, output []byte) ([]byte, error) {
	offset := 0
	expansionCount := 0

	for offset < len(input) {
		// Parse instruction
		instr, err := lito.Disassemble(input, offset, m.config.Mode64)
		if err != nil {
			return output, fmt.Errorf("failed to parse at offset %d: %w", offset, err)
		}

		// Track original offset
		oldOffset := offset
		newOffset := len(output)

		// Create tracking entry
		entry := &AddressEntry{
			OldOffset:     oldOffset,
			OldLength:     int(instr.Length),
			NewOffset:     newOffset,
			Opcode:        instr.Opcode,
			IsControlFlow: instr.IsControlFlow(),
			JumpTarget:    -1, // Will be calculated if needed
		}

		// Calculate jump target if this is a relative jump/call
		if instr.Properties.IsRelativeJump {
			target, err := instr.GetRelativeTarget(uint64(oldOffset))
			if err == nil {
				entry.JumpTarget = int(target)
			}
		}

		// Decide whether to expand instruction
		shouldExpand := m.config.EnableExpansion &&
			m.config.ExpansionPolicy.ShouldExpand(instr, m.rng, expansionCount)

		var instrBytes []byte

		if shouldExpand {
			// Expand instruction
			expanded, err := ExpandInstruction(instr)
			if err != nil {
				// Fallback: use original instruction
				instrBytes = input[offset : offset+int(instr.Length)]
			} else {
				instrBytes = expanded
				entry.Expanded = true
				entry.NewLength = len(expanded)
				expansionCount++
			}
		} else {
			// Use original instruction
			instrBytes = input[offset : offset+int(instr.Length)]
			entry.NewLength = int(instr.Length)
		}

		// Write instruction to output
		output = append(output, instrBytes...)

		// Decide whether to insert dead code after this instruction
		shouldInsert := m.config.EnableDeadCode &&
			ShouldInsertDeadCode(m.rng, m.config.DeadCodeConfig, instr)

		if shouldInsert {
			// Generate dead code
			deadCode := GenerateDeadCode(m.rng, m.config.DeadCodeConfig)
			output = append(output, deadCode...)
			entry.DeadCodeAfter = len(deadCode)
		}

		// Add entry to tracker
		m.tracker.Add(entry)

		// Move to next instruction
		offset += int(instr.Length)

		// Safety check: prevent runaway output
		if len(output) > m.config.MaxOutputSize {
			return output, fmt.Errorf("output size exceeded limit at offset %d", offset)
		}
	}

	// Validate tracker consistency
	if err := m.tracker.Validate(); err != nil {
		return output, fmt.Errorf("tracker validation failed: %w", err)
	}

	return output, nil
}

// pass2Relocate performs the second pass: fix all jump/call offsets
// This is implemented in relocate.go (Chunk 6)
func (m *Morpher) pass2Relocate(code []byte) error {
	// Get all control flow instructions
	controlFlowEntries := m.tracker.GetControlFlowEntries()

	// Fix each one
	for _, entry := range controlFlowEntries {
		if entry.JumpTarget < 0 {
			continue // Not a relative jump/call
		}

		// Get new target address
		newTarget, ok := m.tracker.GetNewAddress(entry.JumpTarget)
		if !ok {
			// Target might be outside morphed region - skip for now
			// Novel: Graceful handling of edge cases
			continue
		}

		// Calculate new relative offset
		// Target is relative to the NEXT instruction
		nextInstrOffset := entry.NewOffset + entry.NewLength
		relativeOffset := newTarget - nextInstrOffset

		// Write new offset to code (implemented in Chunk 6)
		if err := WriteRelativeOffset(code, entry, relativeOffset); err != nil {
			return fmt.Errorf("failed to relocate at offset %d: %w", entry.NewOffset, err)
		}
	}

	return nil
}

// MorphCode is a convenience function for one-shot morphing
// Novel: Simple API for common case
func MorphCode(input []byte, seed uint32) (*MorphResult, error) {
	config := DefaultMorphConfig()
	config.Seed = seed

	morpher := NewMorpher(config)
	return morpher.Morph(input)
}

// MorphWithConfig allows full configuration control
// Novel: Power user API
func MorphWithConfig(input []byte, config *MorphConfig) (*MorphResult, error) {
	morpher := NewMorpher(config)
	return morpher.Morph(input)
}

// EstimateMorphedSize predicts output size without morphing
// Novel: Planning helper for buffer allocation
func EstimateMorphedSize(input []byte, config *MorphConfig) (int, error) {
	// Parse to count expandable instructions
	stream := lito.NewInstructionStream(input, config.Mode64)
	if err := stream.ParseAll(); err != nil {
		return 0, err
	}

	estimated := len(input)

	// Add expansion growth
	if config.EnableExpansion {
		expansionGrowth := CalculateExpansionGrowth(
			stream.Instructions,
			config.ExpansionPolicy.Rate,
		)
		estimated += expansionGrowth
	}

	// Add dead code growth
	if config.EnableDeadCode {
		avgDeadCode := (config.DeadCodeConfig.MinLength + config.DeadCodeConfig.MaxLength) / 2
		deadCodeGrowth := int(float64(len(stream.Instructions)) *
			config.DeadCodeConfig.InsertionRate *
			float64(avgDeadCode))
		estimated += deadCodeGrowth
	}

	return estimated, nil
}
