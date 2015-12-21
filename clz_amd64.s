// +build amd64

// func clzBytes(val uint64) uint64
TEXT ·clzBytes(SB),$0
        MOVQ    val+0(FP), AX
        BSWAPQ  AX      // Reverse order of val
        BSFQ    AX, AX  // Get index of highest set bit in val
        JZ      zero    // BSFQ returns 0 if no bits are set to 1. In that case, return 8.
        SHRQ    $3, AX  // Divide by 8 to get bytes
        MOVQ    AX, ret+8(FP)
        RET
zero:
        MOVQ    $8, ret+8(FP)
        RET


