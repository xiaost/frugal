package reflect

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func init() {
	testhack()
}

func initTestTypesForBenchmark() *TestTypesForBenchmark {
	b1 := true
	s1 := "hello"
	ret := NewTestTypesForBenchmark()
	ret.B1 = &b1
	ret.Str1 = &s1
	ret.Msg1 = &Msg{Type: 1}
	ret.M0 = map[int32]int32{
		1: 2,
		2: 3,
		3: 4,
	}
	ret.M1 = map[string]*Msg{
		"k1":    &Msg{Type: 1},
		"k1231": &Msg{Type: 2},
		"k233":  &Msg{Type: 3},
		"k12":   &Msg{Type: 4},
	}
	ret.L0 = []int32{1, 2, 3}
	ret.L1 = []*Msg{{Type: 1}, {Type: 2}, {Type: 3}}
	ret.Set0 = []int32{1, 2, 3}
	ret.Set1 = []string{"AAAA", "BB", "CCCCC"}
	return ret
}

func BenchmarkEncode(b *testing.B) {
	p := initTestTypesForBenchmark()
	n := EncodedSize(p)
	buf := make([]byte, n)
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		Encode(buf, p)
	}
}

func BenchmarkEncodedSize(b *testing.B) {
	p := initTestTypesForBenchmark()
	_ = EncodedSize(p) // pretouch
	for i := 0; i < b.N; i++ {
		EncodedSize(p)
	}
}

func BenchmarkDecode(b *testing.B) {
	p := initTestTypesForBenchmark()
	n := EncodedSize(p)
	if n <= 0 {
		b.Fatal(n)
	}
	buf := make([]byte, n)
	b.SetBytes(int64(n))
	_, err := Encode(buf, p)
	require.NoError(b, err)

	for i := 0; i < b.N; i++ {
		p.InitDefault()
		Decode(buf, p)
	}
}
