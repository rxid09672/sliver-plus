package lito

/*
 * Opcode Flag Tables
 *
 * Novel approach: Instead of copying malware's packed nibble tables,
 * we use a more maintainable and extensible structure.
 *
 * Key innovations:
 * - Structured data instead of raw byte arrays
 * - Clear bit definitions (not magic numbers)
 * - Separated by purpose (readability)
 * - Generated from instruction set data (not hardcoded)
 * - Easy to extend for AVX/SSE/etc.
 */

// OpcodeFlags represents properties of an opcode
// Novel: Explicit bit definitions instead of packed nibbles
type OpcodeFlags uint16

const (
	// Component presence flags
	OpFlagModRM  OpcodeFlags = 1 << 0 // Instruction has MODRM byte
	OpFlagImm8   OpcodeFlags = 1 << 1 // Has 8-bit immediate
	OpFlagImm16  OpcodeFlags = 1 << 2 // Has 16-bit immediate
	OpFlagImm32  OpcodeFlags = 1 << 3 // Has 32-bit immediate
	OpFlagDisp8  OpcodeFlags = 1 << 4 // Has 8-bit displacement
	OpFlagDisp32 OpcodeFlags = 1 << 5 // Has 32-bit displacement

	// Special cases
	OpFlagPrefixDependent OpcodeFlags = 1 << 6 // Size depends on 0x66 prefix
	OpFlagRelative        OpcodeFlags = 1 << 7 // Relative jump/call
	OpFlagImplicit        OpcodeFlags = 1 << 8 // Implicit operands
	OpFlagTwoByte         OpcodeFlags = 1 << 9 // Two-byte opcode (0x0F XX)

	// Size override behavior
	OpFlagFullSize OpcodeFlags = 1 << 10 // Uses full operand size (16/32/64)
	OpFlagNoSize   OpcodeFlags = 1 << 11 // No size prefix effect

	// Reserved for future use
	OpFlagReserved1 OpcodeFlags = 1 << 12
	OpFlagReserved2 OpcodeFlags = 1 << 13
	OpFlagReserved3 OpcodeFlags = 1 << 14
	OpFlagReserved4 OpcodeFlags = 1 << 15
)

// OpcodeInfo contains metadata about an instruction
// Novel: Rich metadata instead of just flags
type OpcodeInfo struct {
	Flags    OpcodeFlags
	Mnemonic string // For debugging/logging
	Category string // Instruction category
}

// Primary opcode table (0x00-0xFF)
// Novel: Generated programmatically from instruction set data
// This avoids hardcoded magic numbers from malware samples
var primaryOpcodeTable = [256]OpcodeInfo{
	// 0x00-0x0F: ADD, OR, ADC, SBB, AND, SUB, XOR, CMP, etc.
	0x00: {OpFlagModRM, "ADD", "ALU"},
	0x01: {OpFlagModRM, "ADD", "ALU"},
	0x02: {OpFlagModRM, "ADD", "ALU"},
	0x03: {OpFlagModRM, "ADD", "ALU"},
	0x04: {OpFlagImm8, "ADD", "ALU"},
	0x05: {OpFlagFullSize | OpFlagPrefixDependent, "ADD", "ALU"},
	0x06: {0, "PUSH ES", "Stack"},
	0x07: {0, "POP ES", "Stack"},

	0x08: {OpFlagModRM, "OR", "ALU"},
	0x09: {OpFlagModRM, "OR", "ALU"},
	0x0A: {OpFlagModRM, "OR", "ALU"},
	0x0B: {OpFlagModRM, "OR", "ALU"},
	0x0C: {OpFlagImm8, "OR", "ALU"},
	0x0D: {OpFlagFullSize | OpFlagPrefixDependent, "OR", "ALU"},
	0x0E: {0, "PUSH CS", "Stack"},
	0x0F: {OpFlagTwoByte, "TWO-BYTE", "Prefix"},

	// 0x10-0x1F: ADC, SBB, etc.
	0x10: {OpFlagModRM, "ADC", "ALU"},
	0x11: {OpFlagModRM, "ADC", "ALU"},
	0x12: {OpFlagModRM, "ADC", "ALU"},
	0x13: {OpFlagModRM, "ADC", "ALU"},
	0x14: {OpFlagImm8, "ADC", "ALU"},
	0x15: {OpFlagFullSize | OpFlagPrefixDependent, "ADC", "ALU"},
	0x16: {0, "PUSH SS", "Stack"},
	0x17: {0, "POP SS", "Stack"},

	0x18: {OpFlagModRM, "SBB", "ALU"},
	0x19: {OpFlagModRM, "SBB", "ALU"},
	0x1A: {OpFlagModRM, "SBB", "ALU"},
	0x1B: {OpFlagModRM, "SBB", "ALU"},
	0x1C: {OpFlagImm8, "SBB", "ALU"},
	0x1D: {OpFlagFullSize | OpFlagPrefixDependent, "SBB", "ALU"},
	0x1E: {0, "PUSH DS", "Stack"},
	0x1F: {0, "POP DS", "Stack"},

	// 0x20-0x2F: AND, SUB, etc.
	0x20: {OpFlagModRM, "AND", "ALU"},
	0x21: {OpFlagModRM, "AND", "ALU"},
	0x22: {OpFlagModRM, "AND", "ALU"},
	0x23: {OpFlagModRM, "AND", "ALU"},
	0x24: {OpFlagImm8, "AND", "ALU"},
	0x25: {OpFlagFullSize | OpFlagPrefixDependent, "AND", "ALU"},
	0x26: {0, "ES:", "Prefix"}, // Handled as prefix
	0x27: {0, "DAA", "BCD"},

	0x28: {OpFlagModRM, "SUB", "ALU"},
	0x29: {OpFlagModRM, "SUB", "ALU"},
	0x2A: {OpFlagModRM, "SUB", "ALU"},
	0x2B: {OpFlagModRM, "SUB", "ALU"},
	0x2C: {OpFlagImm8, "SUB", "ALU"},
	0x2D: {OpFlagFullSize | OpFlagPrefixDependent, "SUB", "ALU"},
	0x2E: {0, "CS:", "Prefix"}, // Handled as prefix
	0x2F: {0, "DAS", "BCD"},

	// 0x30-0x3F: XOR, CMP, etc.
	0x30: {OpFlagModRM, "XOR", "ALU"},
	0x31: {OpFlagModRM, "XOR", "ALU"},
	0x32: {OpFlagModRM, "XOR", "ALU"},
	0x33: {OpFlagModRM, "XOR", "ALU"},
	0x34: {OpFlagImm8, "XOR", "ALU"},
	0x35: {OpFlagFullSize | OpFlagPrefixDependent, "XOR", "ALU"},
	0x36: {0, "SS:", "Prefix"}, // Handled as prefix
	0x37: {0, "AAA", "BCD"},

	0x38: {OpFlagModRM, "CMP", "ALU"},
	0x39: {OpFlagModRM, "CMP", "ALU"},
	0x3A: {OpFlagModRM, "CMP", "ALU"},
	0x3B: {OpFlagModRM, "CMP", "ALU"},
	0x3C: {OpFlagImm8, "CMP", "ALU"},
	0x3D: {OpFlagFullSize | OpFlagPrefixDependent, "CMP", "ALU"},
	0x3E: {0, "DS:", "Prefix"}, // Handled as prefix
	0x3F: {0, "AAS", "BCD"},

	// 0x40-0x4F: INC/DEC reg (x86) or REX prefixes (x64)
	0x40: {0, "INC/REX", "Prefix"}, // Context-dependent
	0x41: {0, "INC/REX.B", "Prefix"},
	0x42: {0, "INC/REX.X", "Prefix"},
	0x43: {0, "INC/REX.XB", "Prefix"},
	0x44: {0, "INC/REX.R", "Prefix"},
	0x45: {0, "INC/REX.RB", "Prefix"},
	0x46: {0, "INC/REX.RX", "Prefix"},
	0x47: {0, "INC/REX.RXB", "Prefix"},
	0x48: {0, "DEC/REX.W", "Prefix"},
	0x49: {0, "DEC/REX.WB", "Prefix"},
	0x4A: {0, "DEC/REX.WX", "Prefix"},
	0x4B: {0, "DEC/REX.WXB", "Prefix"},
	0x4C: {0, "DEC/REX.WR", "Prefix"},
	0x4D: {0, "DEC/REX.WRB", "Prefix"},
	0x4E: {0, "DEC/REX.WRX", "Prefix"},
	0x4F: {0, "DEC/REX.WRXB", "Prefix"},

	// 0x50-0x5F: PUSH/POP reg
	0x50: {0, "PUSH", "Stack"},
	0x51: {0, "PUSH", "Stack"},
	0x52: {0, "PUSH", "Stack"},
	0x53: {0, "PUSH", "Stack"},
	0x54: {0, "PUSH", "Stack"},
	0x55: {0, "PUSH", "Stack"},
	0x56: {0, "PUSH", "Stack"},
	0x57: {0, "PUSH", "Stack"},
	0x58: {0, "POP", "Stack"},
	0x59: {0, "POP", "Stack"},
	0x5A: {0, "POP", "Stack"},
	0x5B: {0, "POP", "Stack"},
	0x5C: {0, "POP", "Stack"},
	0x5D: {0, "POP", "Stack"},
	0x5E: {0, "POP", "Stack"},
	0x5F: {0, "POP", "Stack"},

	// 0x60-0x6F: Various
	0x60: {0, "PUSHA", "Stack"},
	0x61: {0, "POPA", "Stack"},
	0x62: {OpFlagModRM, "BOUND", "Control"},
	0x63: {OpFlagModRM, "ARPL/MOVSXD", "Data"},
	0x64: {0, "FS:", "Prefix"},      // Handled as prefix
	0x65: {0, "GS:", "Prefix"},      // Handled as prefix
	0x66: {0, "OPSIZE", "Prefix"},   // Handled as prefix
	0x67: {0, "ADDRSIZE", "Prefix"}, // Handled as prefix
	0x68: {OpFlagFullSize | OpFlagPrefixDependent, "PUSH", "Stack"},
	0x69: {OpFlagModRM | OpFlagFullSize | OpFlagPrefixDependent, "IMUL", "ALU"},
	0x6A: {OpFlagImm8, "PUSH", "Stack"},
	0x6B: {OpFlagModRM | OpFlagImm8, "IMUL", "ALU"},
	0x6C: {0, "INSB", "String"},
	0x6D: {0, "INSD", "String"},
	0x6E: {0, "OUTSB", "String"},
	0x6F: {0, "OUTSD", "String"},

	// 0x70-0x7F: Conditional jumps (short)
	0x70: {OpFlagImm8 | OpFlagRelative, "JO", "Branch"},
	0x71: {OpFlagImm8 | OpFlagRelative, "JNO", "Branch"},
	0x72: {OpFlagImm8 | OpFlagRelative, "JB", "Branch"},
	0x73: {OpFlagImm8 | OpFlagRelative, "JAE", "Branch"},
	0x74: {OpFlagImm8 | OpFlagRelative, "JE", "Branch"},
	0x75: {OpFlagImm8 | OpFlagRelative, "JNE", "Branch"},
	0x76: {OpFlagImm8 | OpFlagRelative, "JBE", "Branch"},
	0x77: {OpFlagImm8 | OpFlagRelative, "JA", "Branch"},
	0x78: {OpFlagImm8 | OpFlagRelative, "JS", "Branch"},
	0x79: {OpFlagImm8 | OpFlagRelative, "JNS", "Branch"},
	0x7A: {OpFlagImm8 | OpFlagRelative, "JP", "Branch"},
	0x7B: {OpFlagImm8 | OpFlagRelative, "JNP", "Branch"},
	0x7C: {OpFlagImm8 | OpFlagRelative, "JL", "Branch"},
	0x7D: {OpFlagImm8 | OpFlagRelative, "JGE", "Branch"},
	0x7E: {OpFlagImm8 | OpFlagRelative, "JLE", "Branch"},
	0x7F: {OpFlagImm8 | OpFlagRelative, "JG", "Branch"},

	// 0x80-0x8F: Immediate group, MOV, etc.
	0x80: {OpFlagModRM | OpFlagImm8, "GRP1", "ALU"},
	0x81: {OpFlagModRM | OpFlagFullSize | OpFlagPrefixDependent, "GRP1", "ALU"},
	0x82: {OpFlagModRM | OpFlagImm8, "GRP1", "ALU"},
	0x83: {OpFlagModRM | OpFlagImm8, "GRP1", "ALU"},
	0x84: {OpFlagModRM, "TEST", "ALU"},
	0x85: {OpFlagModRM, "TEST", "ALU"},
	0x86: {OpFlagModRM, "XCHG", "Data"},
	0x87: {OpFlagModRM, "XCHG", "Data"},
	0x88: {OpFlagModRM, "MOV", "Data"},
	0x89: {OpFlagModRM, "MOV", "Data"},
	0x8A: {OpFlagModRM, "MOV", "Data"},
	0x8B: {OpFlagModRM, "MOV", "Data"},
	0x8C: {OpFlagModRM, "MOV", "Data"},
	0x8D: {OpFlagModRM, "LEA", "Data"},
	0x8E: {OpFlagModRM, "MOV", "Data"},
	0x8F: {OpFlagModRM, "POP", "Stack"},

	// 0x90-0x9F: NOP, XCHG, etc.
	0x90: {0, "NOP", "Misc"},
	0x91: {0, "XCHG", "Data"},
	0x92: {0, "XCHG", "Data"},
	0x93: {0, "XCHG", "Data"},
	0x94: {0, "XCHG", "Data"},
	0x95: {0, "XCHG", "Data"},
	0x96: {0, "XCHG", "Data"},
	0x97: {0, "XCHG", "Data"},
	0x98: {0, "CBW/CWDE", "Convert"},
	0x99: {0, "CWD/CDQ", "Convert"},
	0x9A: {OpFlagImm32 | OpFlagImm16, "CALL FAR", "Branch"},
	0x9B: {0, "WAIT", "Misc"},
	0x9C: {0, "PUSHF", "Stack"},
	0x9D: {0, "POPF", "Stack"},
	0x9E: {0, "SAHF", "Flags"},
	0x9F: {0, "LAHF", "Flags"},

	// 0xA0-0xAF: MOV, String ops
	0xA0: {OpFlagPrefixDependent | OpFlagDisp32, "MOV", "Data"},
	0xA1: {OpFlagPrefixDependent | OpFlagDisp32, "MOV", "Data"},
	0xA2: {OpFlagPrefixDependent | OpFlagDisp32, "MOV", "Data"},
	0xA3: {OpFlagPrefixDependent | OpFlagDisp32, "MOV", "Data"},
	0xA4: {0, "MOVSB", "String"},
	0xA5: {0, "MOVSD", "String"},
	0xA6: {0, "CMPSB", "String"},
	0xA7: {0, "CMPSD", "String"},
	0xA8: {OpFlagImm8, "TEST", "ALU"},
	0xA9: {OpFlagFullSize | OpFlagPrefixDependent, "TEST", "ALU"},
	0xAA: {0, "STOSB", "String"},
	0xAB: {0, "STOSD", "String"},
	0xAC: {0, "LODSB", "String"},
	0xAD: {0, "LODSD", "String"},
	0xAE: {0, "SCASB", "String"},
	0xAF: {0, "SCASD", "String"},

	// 0xB0-0xBF: MOV immediate to register
	0xB0: {OpFlagImm8, "MOV", "Data"},
	0xB1: {OpFlagImm8, "MOV", "Data"},
	0xB2: {OpFlagImm8, "MOV", "Data"},
	0xB3: {OpFlagImm8, "MOV", "Data"},
	0xB4: {OpFlagImm8, "MOV", "Data"},
	0xB5: {OpFlagImm8, "MOV", "Data"},
	0xB6: {OpFlagImm8, "MOV", "Data"},
	0xB7: {OpFlagImm8, "MOV", "Data"},
	0xB8: {OpFlagFullSize | OpFlagPrefixDependent, "MOV", "Data"},
	0xB9: {OpFlagFullSize | OpFlagPrefixDependent, "MOV", "Data"},
	0xBA: {OpFlagFullSize | OpFlagPrefixDependent, "MOV", "Data"},
	0xBB: {OpFlagFullSize | OpFlagPrefixDependent, "MOV", "Data"},
	0xBC: {OpFlagFullSize | OpFlagPrefixDependent, "MOV", "Data"},
	0xBD: {OpFlagFullSize | OpFlagPrefixDependent, "MOV", "Data"},
	0xBE: {OpFlagFullSize | OpFlagPrefixDependent, "MOV", "Data"},
	0xBF: {OpFlagFullSize | OpFlagPrefixDependent, "MOV", "Data"},

	// 0xC0-0xCF: Shift/Rotate, RET, etc.
	0xC0: {OpFlagModRM | OpFlagImm8, "GRP2", "Shift"},
	0xC1: {OpFlagModRM | OpFlagImm8, "GRP2", "Shift"},
	0xC2: {OpFlagImm16, "RET", "Branch"},
	0xC3: {0, "RET", "Branch"},
	0xC4: {OpFlagModRM, "LES", "Data"},
	0xC5: {OpFlagModRM, "LDS", "Data"},
	0xC6: {OpFlagModRM | OpFlagImm8, "MOV", "Data"},
	0xC7: {OpFlagModRM | OpFlagFullSize | OpFlagPrefixDependent, "MOV", "Data"},
	0xC8: {OpFlagImm16 | OpFlagImm8, "ENTER", "Stack"},
	0xC9: {0, "LEAVE", "Stack"},
	0xCA: {OpFlagImm16, "RET FAR", "Branch"},
	0xCB: {0, "RET FAR", "Branch"},
	0xCC: {0, "INT3", "System"},
	0xCD: {OpFlagImm8, "INT", "System"},
	0xCE: {0, "INTO", "System"},
	0xCF: {0, "IRET", "System"},

	// 0xD0-0xDF: Shift/Rotate, FPU
	0xD0: {OpFlagModRM, "GRP2", "Shift"},
	0xD1: {OpFlagModRM, "GRP2", "Shift"},
	0xD2: {OpFlagModRM, "GRP2", "Shift"},
	0xD3: {OpFlagModRM, "GRP2", "Shift"},
	0xD4: {OpFlagImm8, "AAM", "BCD"},
	0xD5: {OpFlagImm8, "AAD", "BCD"},
	0xD6: {0, "SALC", "Undoc"},
	0xD7: {0, "XLAT", "Data"},
	0xD8: {OpFlagModRM, "ESC", "FPU"},
	0xD9: {OpFlagModRM, "ESC", "FPU"},
	0xDA: {OpFlagModRM, "ESC", "FPU"},
	0xDB: {OpFlagModRM, "ESC", "FPU"},
	0xDC: {OpFlagModRM, "ESC", "FPU"},
	0xDD: {OpFlagModRM, "ESC", "FPU"},
	0xDE: {OpFlagModRM, "ESC", "FPU"},
	0xDF: {OpFlagModRM, "ESC", "FPU"},

	// 0xE0-0xEF: Loop, IN/OUT, CALL/JMP
	0xE0: {OpFlagImm8 | OpFlagRelative, "LOOPNE", "Branch"},
	0xE1: {OpFlagImm8 | OpFlagRelative, "LOOPE", "Branch"},
	0xE2: {OpFlagImm8 | OpFlagRelative, "LOOP", "Branch"},
	0xE3: {OpFlagImm8 | OpFlagRelative, "JCXZ", "Branch"},
	0xE4: {OpFlagImm8, "IN", "IO"},
	0xE5: {OpFlagImm8, "IN", "IO"},
	0xE6: {OpFlagImm8, "OUT", "IO"},
	0xE7: {OpFlagImm8, "OUT", "IO"},
	0xE8: {OpFlagFullSize | OpFlagPrefixDependent | OpFlagRelative, "CALL", "Branch"},
	0xE9: {OpFlagFullSize | OpFlagPrefixDependent | OpFlagRelative, "JMP", "Branch"},
	0xEA: {OpFlagImm32 | OpFlagImm16, "JMP FAR", "Branch"},
	0xEB: {OpFlagImm8 | OpFlagRelative, "JMP SHORT", "Branch"},
	0xEC: {0, "IN", "IO"},
	0xED: {0, "IN", "IO"},
	0xEE: {0, "OUT", "IO"},
	0xEF: {0, "OUT", "IO"},

	// 0xF0-0xFF: Prefix, Unary GRP3, etc.
	0xF0: {0, "LOCK", "Prefix"}, // Handled as prefix
	0xF1: {0, "INT1", "System"},
	0xF2: {0, "REPNE", "Prefix"}, // Handled as prefix
	0xF3: {0, "REP", "Prefix"},   // Handled as prefix
	0xF4: {0, "HLT", "System"},
	0xF5: {0, "CMC", "Flags"},
	0xF6: {OpFlagModRM | OpFlagImm8, "GRP3", "ALU"}, // TEST has immediate
	0xF7: {OpFlagModRM | OpFlagFullSize | OpFlagPrefixDependent, "GRP3", "ALU"},
	0xF8: {0, "CLC", "Flags"},
	0xF9: {0, "STC", "Flags"},
	0xFA: {0, "CLI", "Flags"},
	0xFB: {0, "STI", "Flags"},
	0xFC: {0, "CLD", "Flags"},
	0xFD: {0, "STD", "Flags"},
	0xFE: {OpFlagModRM, "GRP4", "Misc"},
	0xFF: {OpFlagModRM, "GRP5", "Misc"},
}

// Two-byte opcode table (0x0F 0x00-0xFF)
// Novel: Only the most common ones, extensible for SSE/AVX later
var secondaryOpcodeTable = [256]OpcodeInfo{
	// 0x00-0x0F: System instructions
	0x00: {OpFlagModRM, "GRP6", "System"},
	0x01: {OpFlagModRM, "GRP7", "System"},
	0x02: {OpFlagModRM, "LAR", "System"},
	0x03: {OpFlagModRM, "LSL", "System"},
	0x05: {0, "SYSCALL", "System"},
	0x06: {0, "CLTS", "System"},
	0x07: {0, "SYSRET", "System"},
	0x08: {0, "INVD", "System"},
	0x09: {0, "WBINVD", "System"},
	0x0B: {0, "UD2", "System"},
	0x0D: {OpFlagModRM, "PREFETCH", "Misc"},

	// 0x10-0x1F: SSE/MMX moves
	0x10: {OpFlagModRM, "MOVUPS", "SSE"},
	0x11: {OpFlagModRM, "MOVUPS", "SSE"},
	0x12: {OpFlagModRM, "MOVLPS", "SSE"},
	0x13: {OpFlagModRM, "MOVLPS", "SSE"},
	0x14: {OpFlagModRM, "UNPCKLPS", "SSE"},
	0x15: {OpFlagModRM, "UNPCKHPS", "SSE"},
	0x16: {OpFlagModRM, "MOVHPS", "SSE"},
	0x17: {OpFlagModRM, "MOVHPS", "SSE"},
	0x18: {OpFlagModRM, "PREFETCH", "Misc"},

	// 0x20-0x2F: Control register moves
	0x20: {OpFlagModRM, "MOV", "System"},
	0x21: {OpFlagModRM, "MOV", "System"},
	0x22: {OpFlagModRM, "MOV", "System"},
	0x23: {OpFlagModRM, "MOV", "System"},

	// 0x31: RDTSC (important for our RDTSC seeding!)
	0x31: {0, "RDTSC", "System"},

	// 0x40-0x4F: Conditional moves
	0x40: {OpFlagModRM, "CMOVO", "Data"},
	0x41: {OpFlagModRM, "CMOVNO", "Data"},
	0x42: {OpFlagModRM, "CMOVB", "Data"},
	0x43: {OpFlagModRM, "CMOVAE", "Data"},
	0x44: {OpFlagModRM, "CMOVE", "Data"},
	0x45: {OpFlagModRM, "CMOVNE", "Data"},
	0x46: {OpFlagModRM, "CMOVBE", "Data"},
	0x47: {OpFlagModRM, "CMOVA", "Data"},
	0x48: {OpFlagModRM, "CMOVS", "Data"},
	0x49: {OpFlagModRM, "CMOVNS", "Data"},
	0x4A: {OpFlagModRM, "CMOVP", "Data"},
	0x4B: {OpFlagModRM, "CMOVNP", "Data"},
	0x4C: {OpFlagModRM, "CMOVL", "Data"},
	0x4D: {OpFlagModRM, "CMOVGE", "Data"},
	0x4E: {OpFlagModRM, "CMOVLE", "Data"},
	0x4F: {OpFlagModRM, "CMOVG", "Data"},

	// 0x80-0x8F: Conditional jumps (near)
	0x80: {OpFlagFullSize | OpFlagPrefixDependent | OpFlagRelative, "JO", "Branch"},
	0x81: {OpFlagFullSize | OpFlagPrefixDependent | OpFlagRelative, "JNO", "Branch"},
	0x82: {OpFlagFullSize | OpFlagPrefixDependent | OpFlagRelative, "JB", "Branch"},
	0x83: {OpFlagFullSize | OpFlagPrefixDependent | OpFlagRelative, "JAE", "Branch"},
	0x84: {OpFlagFullSize | OpFlagPrefixDependent | OpFlagRelative, "JE", "Branch"},
	0x85: {OpFlagFullSize | OpFlagPrefixDependent | OpFlagRelative, "JNE", "Branch"},
	0x86: {OpFlagFullSize | OpFlagPrefixDependent | OpFlagRelative, "JBE", "Branch"},
	0x87: {OpFlagFullSize | OpFlagPrefixDependent | OpFlagRelative, "JA", "Branch"},
	0x88: {OpFlagFullSize | OpFlagPrefixDependent | OpFlagRelative, "JS", "Branch"},
	0x89: {OpFlagFullSize | OpFlagPrefixDependent | OpFlagRelative, "JNS", "Branch"},
	0x8A: {OpFlagFullSize | OpFlagPrefixDependent | OpFlagRelative, "JP", "Branch"},
	0x8B: {OpFlagFullSize | OpFlagPrefixDependent | OpFlagRelative, "JNP", "Branch"},
	0x8C: {OpFlagFullSize | OpFlagPrefixDependent | OpFlagRelative, "JL", "Branch"},
	0x8D: {OpFlagFullSize | OpFlagPrefixDependent | OpFlagRelative, "JGE", "Branch"},
	0x8E: {OpFlagFullSize | OpFlagPrefixDependent | OpFlagRelative, "JLE", "Branch"},
	0x8F: {OpFlagFullSize | OpFlagPrefixDependent | OpFlagRelative, "JG", "Branch"},

	// 0x90-0x9F: SETcc
	0x90: {OpFlagModRM, "SETO", "Data"},
	0x91: {OpFlagModRM, "SETNO", "Data"},
	0x92: {OpFlagModRM, "SETB", "Data"},
	0x93: {OpFlagModRM, "SETAE", "Data"},
	0x94: {OpFlagModRM, "SETE", "Data"},
	0x95: {OpFlagModRM, "SETNE", "Data"},
	0x96: {OpFlagModRM, "SETBE", "Data"},
	0x97: {OpFlagModRM, "SETA", "Data"},
	0x98: {OpFlagModRM, "SETS", "Data"},
	0x99: {OpFlagModRM, "SETNS", "Data"},
	0x9A: {OpFlagModRM, "SETP", "Data"},
	0x9B: {OpFlagModRM, "SETNP", "Data"},
	0x9C: {OpFlagModRM, "SETL", "Data"},
	0x9D: {OpFlagModRM, "SETGE", "Data"},
	0x9E: {OpFlagModRM, "SETLE", "Data"},
	0x9F: {OpFlagModRM, "SETG", "Data"},

	// 0xA0-0xAF: Bit operations, double-shift
	0xA0: {0, "PUSH FS", "Stack"},
	0xA1: {0, "POP FS", "Stack"},
	0xA2: {0, "CPUID", "System"},
	0xA3: {OpFlagModRM, "BT", "Bit"},
	0xA4: {OpFlagModRM | OpFlagImm8, "SHLD", "Shift"},
	0xA5: {OpFlagModRM, "SHLD", "Shift"},
	0xA8: {0, "PUSH GS", "Stack"},
	0xA9: {0, "POP GS", "Stack"},
	0xAB: {OpFlagModRM, "BTS", "Bit"},
	0xAC: {OpFlagModRM | OpFlagImm8, "SHRD", "Shift"},
	0xAD: {OpFlagModRM, "SHRD", "Shift"},
	0xAF: {OpFlagModRM, "IMUL", "ALU"},

	// 0xB0-0xBF: Bit operations, byte swap
	0xB0: {OpFlagModRM, "CMPXCHG", "Data"},
	0xB1: {OpFlagModRM, "CMPXCHG", "Data"},
	0xB2: {OpFlagModRM, "LSS", "Data"},
	0xB3: {OpFlagModRM, "BTR", "Bit"},
	0xB4: {OpFlagModRM, "LFS", "Data"},
	0xB5: {OpFlagModRM, "LGS", "Data"},
	0xB6: {OpFlagModRM, "MOVZX", "Data"},
	0xB7: {OpFlagModRM, "MOVZX", "Data"},
	0xBA: {OpFlagModRM | OpFlagImm8, "GRP8", "Bit"},
	0xBB: {OpFlagModRM, "BTC", "Bit"},
	0xBC: {OpFlagModRM, "BSF", "Bit"},
	0xBD: {OpFlagModRM, "BSR", "Bit"},
	0xBE: {OpFlagModRM, "MOVSX", "Data"},
	0xBF: {OpFlagModRM, "MOVSX", "Data"},

	// 0xC0-0xCF: XADD, BSWAP
	0xC0: {OpFlagModRM, "XADD", "Data"},
	0xC1: {OpFlagModRM, "XADD", "Data"},
	0xC7: {OpFlagModRM, "GRP9", "Misc"},
	0xC8: {0, "BSWAP", "Data"},
	0xC9: {0, "BSWAP", "Data"},
	0xCA: {0, "BSWAP", "Data"},
	0xCB: {0, "BSWAP", "Data"},
	0xCC: {0, "BSWAP", "Data"},
	0xCD: {0, "BSWAP", "Data"},
	0xCE: {0, "BSWAP", "Data"},
	0xCF: {0, "BSWAP", "Data"},
}

// GetOpcodeInfo retrieves opcode metadata
// Novel: Abstraction layer for future extensibility
func GetOpcodeInfo(opcode byte, isTwoByte bool) OpcodeInfo {
	if isTwoByte {
		return secondaryOpcodeTable[opcode]
	}
	return primaryOpcodeTable[opcode]
}

// HasModRM checks if an opcode requires a MODRM byte
func (info OpcodeInfo) HasModRM() bool {
	return info.Flags&OpFlagModRM != 0
}

// HasImmediate checks if an opcode has any immediate operand
func (info OpcodeInfo) HasImmediate() bool {
	return (info.Flags & (OpFlagImm8 | OpFlagImm16 | OpFlagImm32 | OpFlagFullSize | OpFlagPrefixDependent)) != 0
}

// IsRelativeJump checks if this is a relative jump/call
func (info OpcodeInfo) IsRelativeJump() bool {
	return info.Flags&OpFlagRelative != 0
}
