package reflect

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeBitset(t *testing.T) {
	s := &fieldBitset{}
	for i := uint16(0); i < ^uint16(0); i++ {
		if i%2 == 0 {
			s.set(i)
		}
		if i%4 == 0 {
			s.unset(i)
		}
	}
	for i := uint16(0); i < ^uint16(0); i++ {
		if i%4 == 0 {
			require.False(t, s.test(i))
		} else if i%2 == 0 {
			require.True(t, s.test(i))
		} else {
			require.False(t, s.test(i))
		}
	}
}
