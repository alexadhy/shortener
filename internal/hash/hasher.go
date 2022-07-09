package hash

import (
	"fmt"
	"math/big"

	"github.com/zeebo/blake3"
)

func Hash(s string) (string, string) {
	h := blake3.New()
	h.WriteString(s)
	x := h.Sum(nil)
	return fmt.Sprintf("%x", x), encode(x)[:8]
}

var (
	base36 = []byte{
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
		'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T',
		'U', 'V', 'W', 'X', 'Y', 'Z'}

	bigRadix = big.NewInt(36)
	bigZero  = big.NewInt(0)
)

// encodeBytes encodes a byte slice to base36.
func encodeBytes(b []byte) []byte {
	x := new(big.Int)
	x.SetBytes(b)

	// pre-alloc
	answer := make([]byte, 0, len(b)*136/100)
	for x.Cmp(bigZero) > 0 {
		mod := new(big.Int)
		x.DivMod(x, bigRadix, mod)
		answer = append(answer, base36[mod.Int64()])
	}

	// leading zero bytes
	for _, i := range b {
		if i != 0 {
			break
		}
		answer = append(answer, base36[0])
	}

	// reverse
	alen := len(answer)
	for i := 0; i < alen/2; i++ {
		answer[i], answer[alen-1-i] = answer[alen-1-i], answer[i]
	}

	return answer
}

// encode encodes a byte slice to base36 string.
func encode(b []byte) string {
	return string(encodeBytes(b))
}
