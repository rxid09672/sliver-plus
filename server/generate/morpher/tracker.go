package morpher

/*
 * Address Tracking System for Code Morphing
 *
 * Tracks the mapping between original and morphed code addresses
 * to enable proper relocation of jumps, calls, and other relative references.
 *
 * Novel approach: Map-based O(1) lookups instead of linear scans
 */

import (
	"fmt"
	"sync"
)

// AddressEntry represents a mapping between old and new addresses
// Novel: Rich metadata for comprehensive tracking
type AddressEntry struct {
	// Original code
	OldOffset int // Offset in original code
	OldLength int // Original instruction length

	// Morphed code
	NewOffset int // Offset in morphed code
	NewLength int // Morphed instruction length

	// Jump/Call tracking
	JumpTarget    int  // Target offset (if this is a jump/call, -1 otherwise)
	IsControlFlow bool // Whether this is a control flow instruction

	// Metadata
	Opcode        byte // Primary opcode for debugging
	Expanded      bool // Whether instruction was expanded
	DeadCodeAfter int  // Bytes of dead code inserted after this instruction
}

// AddressTracker manages the address mapping during morphing
// Novel: Thread-safe with fast lookups
type AddressTracker struct {
	mu       sync.RWMutex
	entries  []*AddressEntry
	oldToNew map[int]*AddressEntry // Fast lookup: old offset → entry
	newToOld map[int]*AddressEntry // Reverse lookup: new offset → entry
}

// NewAddressTracker creates a new address tracker
func NewAddressTracker() *AddressTracker {
	return &AddressTracker{
		entries:  make([]*AddressEntry, 0, 256),
		oldToNew: make(map[int]*AddressEntry),
		newToOld: make(map[int]*AddressEntry),
	}
}

// Add registers a new address mapping
// Novel: Automatic bidirectional indexing
func (t *AddressTracker) Add(entry *AddressEntry) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.entries = append(t.entries, entry)
	t.oldToNew[entry.OldOffset] = entry
	t.newToOld[entry.NewOffset] = entry
}

// GetNewAddress returns the new address for an old address
// Novel: O(1) map lookup instead of O(n) scan
func (t *AddressTracker) GetNewAddress(oldOffset int) (int, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if entry, ok := t.oldToNew[oldOffset]; ok {
		return entry.NewOffset, true
	}
	return 0, false
}

// GetOldAddress returns the old address for a new address
// Novel: Reverse lookup capability
func (t *AddressTracker) GetOldAddress(newOffset int) (int, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if entry, ok := t.newToOld[newOffset]; ok {
		return entry.OldOffset, true
	}
	return 0, false
}

// GetEntry returns the complete entry for an old address
func (t *AddressTracker) GetEntry(oldOffset int) (*AddressEntry, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if entry, ok := t.oldToNew[oldOffset]; ok {
		return entry, true
	}
	return nil, false
}

// GetEntryByNewAddress returns entry by new address
func (t *AddressTracker) GetEntryByNewAddress(newOffset int) (*AddressEntry, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if entry, ok := t.newToOld[newOffset]; ok {
		return entry, true
	}
	return nil, false
}

// GetAllEntries returns all entries (for iteration)
func (t *AddressTracker) GetAllEntries() []*AddressEntry {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Return a copy to prevent external modification
	entries := make([]*AddressEntry, len(t.entries))
	copy(entries, t.entries)
	return entries
}

// GetControlFlowEntries returns only control flow instructions
// Novel: Filter for jump/call relocation
func (t *AddressTracker) GetControlFlowEntries() []*AddressEntry {
	t.mu.RLock()
	defer t.mu.RUnlock()

	controlFlow := make([]*AddressEntry, 0)
	for _, entry := range t.entries {
		if entry.IsControlFlow {
			controlFlow = append(controlFlow, entry)
		}
	}
	return controlFlow
}

// Count returns the number of tracked addresses
func (t *AddressTracker) Count() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.entries)
}

// GetStats returns statistics about the tracking
// Novel: Metrics for debugging and optimization
type TrackerStats struct {
	TotalInstructions  int
	ControlFlowCount   int
	ExpandedCount      int
	TotalDeadCodeBytes int
	OldCodeSize        int
	NewCodeSize        int
	ExpansionRatio     float64
}

func (t *AddressTracker) GetStats() *TrackerStats {
	t.mu.RLock()
	defer t.mu.RUnlock()

	stats := &TrackerStats{}

	for _, entry := range t.entries {
		stats.TotalInstructions++

		if entry.IsControlFlow {
			stats.ControlFlowCount++
		}

		if entry.Expanded {
			stats.ExpandedCount++
		}

		stats.TotalDeadCodeBytes += entry.DeadCodeAfter
		stats.OldCodeSize += entry.OldLength
		stats.NewCodeSize += entry.NewLength + entry.DeadCodeAfter
	}

	if stats.OldCodeSize > 0 {
		stats.ExpansionRatio = float64(stats.NewCodeSize) / float64(stats.OldCodeSize)
	}

	return stats
}

// Validate checks the tracker for consistency
// Novel: Safety validation before relocation
func (t *AddressTracker) Validate() error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Check for gaps or overlaps
	for i := 1; i < len(t.entries); i++ {
		prev := t.entries[i-1]
		curr := t.entries[i]

		// Old offsets should be sequential
		expectedOld := prev.OldOffset + prev.OldLength
		if curr.OldOffset != expectedOld {
			return fmt.Errorf("gap in old offsets: %d → %d (expected %d)",
				prev.OldOffset, curr.OldOffset, expectedOld)
		}

		// New offsets should be sequential (accounting for dead code)
		expectedNew := prev.NewOffset + prev.NewLength + prev.DeadCodeAfter
		if curr.NewOffset != expectedNew {
			return fmt.Errorf("gap in new offsets: %d → %d (expected %d)",
				prev.NewOffset, curr.NewOffset, expectedNew)
		}
	}

	return nil
}

// Reset clears all tracking data for reuse
// Novel: Memory efficiency for batch processing
func (t *AddressTracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.entries = t.entries[:0]
	t.oldToNew = make(map[int]*AddressEntry)
	t.newToOld = make(map[int]*AddressEntry)
}
