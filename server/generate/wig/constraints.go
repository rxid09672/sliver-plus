package wig

/*
 * Safety Constraints & Executable Manifold
 *
 * Defines the "Executable Manifold" - the subspace of the 100D vector space
 * where all points represent valid, executable code that satisfies:
 * 1. CPU instruction semantics (valid opcodes, encodings)
 * 2. ABI/calling conventions (stack balance, register preservation)
 * 3. Control-flow integrity (valid CFG, no ROP gadgets)
 * 4. Binary format integrity (relocations, symbols)
 *
 * Novel approach: Mathematical constraints on vector space that guarantee
 * executable code while still allowing exploration of alien patterns.
 */

import (
	"fmt"

	"github.com/bishopfox/sliver/server/generate/lito"
)

// DimensionConstraint defines valid range for a single dimension
// Novel: Mathematical bounds ensuring validity
type DimensionConstraint struct {
	DimensionID int
	MinValue    float64
	MaxValue    float64
	MustEqual   *float64 // If set, dimension must equal this value
	Description string
}

// DependencyRule defines relationships between dimensions
// Novel: Enforces semantic constraints (e.g., PUSH count == POP count)
type DependencyRule struct {
	SourceDim        int
	TargetDim        int
	RelationshipType DependencyType
	Description      string

	// Function that calculates valid range for target given source value
	CalculateRange func(sourceValue float64) (min, max float64)
}

// DependencyType categorizes constraint relationships
type DependencyType int

const (
	DepEqual        DependencyType = iota // Target must equal source (stack balance)
	DepInverse                            // Target = 1 - source (complementary)
	DepProportional                       // Target ∝ source (correlated)
	DepBounded                            // Target within range based on source
)

// ExecutableManifold defines the safe subspace of vector space
// Novel: Constrained exploration ensuring all points = valid code
type ExecutableManifold struct {
	// Dimension constraints
	Constraints  map[int]*DimensionConstraint
	Dependencies []*DependencyRule

	// Semantic constraints (Constraint #1: CPU semantics)
	AllowedOpcodes    map[byte]bool
	ForbiddenPatterns [][]byte // ROP gadgets, illegal sequences

	// ABI constraints (Constraint #2: Calling conventions)
	RequireStackBalance bool
	ScratchRegisters    []int // Safe to use
	PreserveRegisters   []int // Must not modify
	StackAlignment      int   // 16 for x64

	// CFG constraints (Constraint #3: Control flow)
	MaxJumpDistance    int
	MinBasicBlockSize  int
	MaxNestingDepth    int
	AllowBackwardJumps bool

	// Binary format constraints (Constraint #4: Relocations)
	OnlyMorphTextSection bool
	AutoFixRelocations   bool
	UpdateSymbolTable    bool

	// Platform info
	Mode64   bool
	Platform string // "windows", "linux", "darwin"
}

// NewExecutableManifold creates a manifold with safe defaults
// Novel: Conservative defaults ensuring validity
func NewExecutableManifold(mode64 bool, platform string) *ExecutableManifold {
	em := &ExecutableManifold{
		Constraints:          make(map[int]*DimensionConstraint),
		Dependencies:         make([]*DependencyRule, 0),
		AllowedOpcodes:       buildAllowedOpcodeWhitelist(),
		ForbiddenPatterns:    buildROPGadgetBlacklist(),
		RequireStackBalance:  true,
		ScratchRegisters:     []int{0, 1, 2},    // EAX, ECX, EDX
		PreserveRegisters:    []int{3, 5, 6, 7}, // EBX, EBP, ESI, EDI
		StackAlignment:       16,
		MaxJumpDistance:      1024, // Reasonable jump range
		MinBasicBlockSize:    3,    // Minimum 3 instructions
		MaxNestingDepth:      3,
		AllowBackwardJumps:   false, // Avoid loop-like structures in junk
		OnlyMorphTextSection: true,
		AutoFixRelocations:   true,
		UpdateSymbolTable:    true,
		Mode64:               mode64,
		Platform:             platform,
	}

	// Define constraints for each dimension
	em.defineConstraints()
	em.defineDependencies()

	return em
}

// defineConstraints sets bounds for each dimension
func (em *ExecutableManifold) defineConstraints() {
	// Opcode frequencies (0-19): All [0.0, 1.0]
	for i := 0; i < 20; i++ {
		em.Constraints[i] = &DimensionConstraint{
			DimensionID: i,
			MinValue:    0.0,
			MaxValue:    1.0,
			Description: "Opcode frequency",
		}
	}

	// Length distribution (20-29): Must sum to ~1.0
	for i := 20; i < 30; i++ {
		em.Constraints[i] = &DimensionConstraint{
			DimensionID: i,
			MinValue:    0.0,
			MaxValue:    1.0,
			Description: "Length distribution",
		}
	}

	// Complexity (30-34): Must sum to 1.0
	em.Constraints[DimSimpleInstr] = &DimensionConstraint{
		DimensionID: DimSimpleInstr,
		MinValue:    0.0,
		MaxValue:    1.0,
		Description: "Simple instruction ratio",
	}

	// Register usage (35-42): Constrained to scratch regs
	// EAX, ECX, EDX can be used freely
	em.Constraints[DimEAXUsage] = &DimensionConstraint{
		DimensionID: DimEAXUsage,
		MinValue:    0.0,
		MaxValue:    1.0,
		Description: "EAX usage (scratch reg)",
	}

	// ESP must be minimal (stack pointer - dangerous)
	em.Constraints[DimESPUsage] = &DimensionConstraint{
		DimensionID: DimESPUsage,
		MinValue:    0.0,
		MaxValue:    0.1, // Very limited ESP usage
		Description: "ESP usage (minimize)",
	}

	// EBP should be zero (callee-saved)
	em.Constraints[DimEBPUsage] = &DimensionConstraint{
		DimensionID: DimEBPUsage,
		MinValue:    0.0,
		MaxValue:    0.0, // Don't use EBP
		Description: "EBP usage (forbidden)",
	}

	// Control flow density (43): Reasonable limits
	em.Constraints[DimControlFlowDensity] = &DimensionConstraint{
		DimensionID: DimControlFlowDensity,
		MinValue:    0.0,
		MaxValue:    0.3, // Max 30% control flow (avoid gadget chains)
		Description: "Control flow density",
	}

	// Stack balance enforced via dependency (see below)
}

// defineDependencies sets up dimension relationships
// Novel: Enforces semantic constraints via vector relationships
func (em *ExecutableManifold) defineDependencies() {
	// CRITICAL CONSTRAINT #2: Stack Balance (ABI)
	// PUSH frequency must equal POP frequency
	stackBalanceRule := &DependencyRule{
		SourceDim:        DimPUSHFrequency,
		TargetDim:        DimPOPFrequency,
		RelationshipType: DepEqual,
		Description:      "Stack must be balanced (PUSH == POP)",
		CalculateRange: func(sourceValue float64) (min, max float64) {
			// Target must equal source (with small tolerance)
			tolerance := 0.05
			return sourceValue - tolerance, sourceValue + tolerance
		},
	}
	em.Dependencies = append(em.Dependencies, stackBalanceRule)

	// CRITICAL CONSTRAINT #3: Jump Distance vs Block Size (CFG)
	// Longer jumps require larger basic blocks (avoid ROP-like behavior)
	jumpBlockRule := &DependencyRule{
		SourceDim:        DimJumpDistance,
		TargetDim:        DimBasicBlockSize,
		RelationshipType: DepProportional,
		Description:      "Long jumps need large blocks",
		CalculateRange: func(sourceValue float64) (min, max float64) {
			// If jump distance is high, block size must be high
			minBlock := sourceValue * 0.5 // Proportional
			return minBlock, 1.0
		},
	}
	em.Dependencies = append(em.Dependencies, jumpBlockRule)

	// Register coherence: If using many regs, avoid callee-saved
	regCoherenceRule := &DependencyRule{
		SourceDim:        DimEAXUsage,
		TargetDim:        DimEBXUsage,
		RelationshipType: DepInverse,
		Description:      "High scratch usage → low callee-saved usage",
		CalculateRange: func(sourceValue float64) (min, max float64) {
			// Inverse relationship
			return 0.0, 1.0 - sourceValue
		},
	}
	em.Dependencies = append(em.Dependencies, regCoherenceRule)
}

// buildAllowedOpcodeWhitelist creates whitelist of safe opcodes
// Novel: CONSTRAINT #1 - Only semantically-safe instructions
func buildAllowedOpcodeWhitelist() map[byte]bool {
	allowed := make(map[byte]bool)

	// Explicitly safe opcodes (semantic NOPs or safe operations)
	safeOpcodes := []byte{
		0x90,       // NOP
		0x89, 0x8B, // MOV (with constraints)
		0x8D,       // LEA (with constraints)
		0x85,       // TEST (read-only)
		0x39, 0x3B, // CMP (read-only)
		0x87,                                           // XCHG (with constraints)
		0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57, // PUSH (with balance)
		0x58, 0x59, 0x5A, 0x5B, 0x5C, 0x5D, 0x5E, 0x5F, // POP (with balance)
		// Deliberately limited set - safety over variety
	}

	for _, opcode := range safeOpcodes {
		allowed[opcode] = true
	}

	return allowed
}

// buildROPGadgetBlacklist creates blacklist of ROP patterns
// Novel: CONSTRAINT #3 - No ROP-like gadgets
func buildROPGadgetBlacklist() [][]byte {
	return [][]byte{
		{0x5C, 0xC3},       // POP ESP; RET
		{0x58, 0xC3},       // POP EAX; RET
		{0x59, 0xC3},       // POP ECX; RET
		{0xFF, 0xE0},       // JMP EAX
		{0xFF, 0xE4},       // JMP ESP
		{0xFF, 0xD0},       // CALL EAX (suspicious in junk)
		{0xC3, 0x90, 0x90}, // RET followed by NOPs (gadget-like)
		// Add more known ROP patterns
	}
}

// Project projects a vector onto the executable manifold
// Novel: Ensures vector represents valid code
func (em *ExecutableManifold) Project(vector *CodeVector) *CodeVector {
	projected := vector.Clone()

	// Step 1: Apply dimension constraints
	for dim, constraint := range em.Constraints {
		value := projected.Values[dim]

		// Enforce must-equal constraints
		if constraint.MustEqual != nil {
			projected.Values[dim] = *constraint.MustEqual
			continue
		}

		// Clamp to valid range
		if value < constraint.MinValue {
			projected.Values[dim] = constraint.MinValue
		} else if value > constraint.MaxValue {
			projected.Values[dim] = constraint.MaxValue
		}
	}

	// Step 2: Enforce dependencies (CRITICAL!)
	for _, dep := range em.Dependencies {
		sourceValue := projected.Values[dep.SourceDim]
		targetValue := projected.Values[dep.TargetDim]

		// Calculate valid range for target
		minTarget, maxTarget := dep.CalculateRange(sourceValue)

		// Enforce range
		if targetValue < minTarget {
			projected.Values[dep.TargetDim] = minTarget
		} else if targetValue > maxTarget {
			projected.Values[dep.TargetDim] = maxTarget
		}
	}

	// Step 3: Normalize distributions (must sum to 1.0)
	projected.normalizeDistributions()

	return projected
}

// normalizeDistributions ensures probability distributions sum to 1.0
func (cv *CodeVector) normalizeDistributions() {
	// Normalize complexity distribution (dims 30-32)
	complexitySum := cv.Values[DimSimpleInstr] + cv.Values[DimMediumInstr] + cv.Values[DimComplexInstr]
	if complexitySum > 0 {
		cv.Values[DimSimpleInstr] /= complexitySum
		cv.Values[DimMediumInstr] /= complexitySum
		cv.Values[DimComplexInstr] /= complexitySum
	}

	// Normalize length distribution (dims 20-24)
	lengthSum := cv.Values[DimLength1Byte] + cv.Values[DimLength2Byte] +
		cv.Values[DimLength3Byte] + cv.Values[DimLength4to6Byte] + cv.Values[DimLength7PlusByte]
	if lengthSum > 0 {
		cv.Values[DimLength1Byte] /= lengthSum
		cv.Values[DimLength2Byte] /= lengthSum
		cv.Values[DimLength3Byte] /= lengthSum
		cv.Values[DimLength4to6Byte] /= lengthSum
		cv.Values[DimLength7PlusByte] /= lengthSum
	}
}

// IsValid checks if a vector satisfies all constraints
// Novel: Pre-generation validation
func (em *ExecutableManifold) IsValid(vector *CodeVector) bool {
	// Check dimension constraints
	for dim, constraint := range em.Constraints {
		value := vector.Values[dim]

		if constraint.MustEqual != nil {
			if value != *constraint.MustEqual {
				return false
			}
		}

		if value < constraint.MinValue || value > constraint.MaxValue {
			return false
		}
	}

	// Check dependencies (CRITICAL: Stack balance, etc.)
	for _, dep := range em.Dependencies {
		sourceValue := vector.Values[dep.SourceDim]
		targetValue := vector.Values[dep.TargetDim]

		minTarget, maxTarget := dep.CalculateRange(sourceValue)

		if targetValue < minTarget || targetValue > maxTarget {
			return false // Dependency violated!
		}
	}

	return true
}

// Validate performs comprehensive validation
// Novel: 5-layer validation matching your 4 constraints
func (em *ExecutableManifold) Validate(vector *CodeVector, generatedCode []byte) error {
	// Layer 1: Vector constraint satisfaction
	if !em.IsValid(vector) {
		return fmt.Errorf("vector violates constraints")
	}

	// Layer 2: CPU semantics (CONSTRAINT #1)
	if err := em.validateSemantics(generatedCode); err != nil {
		return fmt.Errorf("semantic validation failed: %w", err)
	}

	// Layer 3: ABI compliance (CONSTRAINT #2)
	if err := em.validateABI(generatedCode); err != nil {
		return fmt.Errorf("ABI validation failed: %w", err)
	}

	// Layer 4: CFG integrity (CONSTRAINT #3)
	if err := em.validateCFG(generatedCode); err != nil {
		return fmt.Errorf("CFG validation failed: %w", err)
	}

	// Layer 5: Binary format (CONSTRAINT #4)
	// Note: Binary format validation happens at binary-level, not code chunk

	return nil
}

// validateSemantics ensures CPU instruction validity
// CONSTRAINT #1: CPU instruction semantics
func (em *ExecutableManifold) validateSemantics(code []byte) error {
	// Parse with Lito - ensures valid x86/x64
	stream := lito.NewInstructionStream(code, em.Mode64)
	if err := stream.ParseAll(); err != nil {
		return fmt.Errorf("invalid instruction encoding: %w", err)
	}

	// Check each instruction uses allowed opcodes only
	for _, instr := range stream.Instructions {
		if !em.AllowedOpcodes[instr.Opcode] {
			return fmt.Errorf("forbidden opcode: 0x%02X", instr.Opcode)
		}
	}

	// Check for forbidden patterns (ROP gadgets)
	for _, pattern := range em.ForbiddenPatterns {
		if containsPattern(code, pattern) {
			return fmt.Errorf("contains forbidden pattern (ROP gadget)")
		}
	}

	return nil
}

// validateABI ensures calling convention compliance
// CONSTRAINT #2: ABI/Calling conventions
func (em *ExecutableManifold) validateABI(code []byte) error {
	stream := lito.NewInstructionStream(code, em.Mode64)
	stream.ParseAll()

	// Check stack balance
	if em.RequireStackBalance {
		stackDelta := 0
		for _, instr := range stream.Instructions {
			if instr.Opcode >= 0x50 && instr.Opcode <= 0x57 {
				stackDelta-- // PUSH
			} else if instr.Opcode >= 0x58 && instr.Opcode <= 0x5F {
				stackDelta++ // POP
			}
		}

		if stackDelta != 0 {
			return fmt.Errorf("stack imbalance: delta = %d", stackDelta)
		}
	}

	// Check register usage (only scratch registers)
	for _, instr := range stream.Instructions {
		if instr.Properties.HasModRM {
			reg := int((instr.ModRM >> 3) & 0x07)
			rm := int(instr.ModRM & 0x07)

			// Check if using forbidden registers
			for _, forbidden := range em.PreserveRegisters {
				if reg == forbidden || rm == forbidden {
					return fmt.Errorf("uses callee-saved register: %d", forbidden)
				}
			}
		}
	}

	return nil
}

// validateCFG ensures control flow integrity
// CONSTRAINT #3: Control-flow integrity
func (em *ExecutableManifold) validateCFG(code []byte) error {
	stream := lito.NewInstructionStream(code, em.Mode64)
	stream.ParseAll()

	// Get control flow instructions
	controlFlow := stream.GetControlFlowInstructions()

	// Check jump distances
	for _, instr := range controlFlow {
		if instr.Properties.IsRelativeJump {
			// Calculate jump distance
			target, err := instr.GetRelativeTarget(0) // Relative to start
			if err != nil {
				continue
			}

			distance := int(target)
			if distance < 0 {
				distance = -distance
			}

			// Enforce maximum jump distance
			if distance > em.MaxJumpDistance {
				return fmt.Errorf("jump distance %d exceeds limit %d", distance, em.MaxJumpDistance)
			}
		}
	}

	// Check for ROP-like patterns
	if len(controlFlow) > 0 {
		// Too many tiny jumps = gadget chain
		avgBlockSize := len(stream.Instructions) / (len(controlFlow) + 1)
		if avgBlockSize < em.MinBasicBlockSize {
			return fmt.Errorf("basic blocks too small: avg %d, min %d", avgBlockSize, em.MinBasicBlockSize)
		}
	}

	return nil
}

// containsPattern checks if code contains a byte pattern
func containsPattern(code []byte, pattern []byte) bool {
	if len(pattern) > len(code) {
		return false
	}

	for i := 0; i <= len(code)-len(pattern); i++ {
		match := true
		for j := 0; j < len(pattern); j++ {
			if code[i+j] != pattern[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}

	return false
}
