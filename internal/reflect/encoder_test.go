package reflect

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NewMsg().BLength()
const encodedMsgSize = 15

// NewTestTypes().BLength()
const encodedTestTypesSize = 176

// NewTestTypesOptional().BLength()
const encodedTestTypesOptionalSize = 1

// NewTestTypesWithDefault().BLength()
const encodedTestTypesWithDefaul = 25

func TestEncode(t *testing.T) {
	// default NewXXXX cases
	require.Equal(t, encodedTestTypesSize, EncodedSize(NewTestTypes()))
	require.Equal(t, encodedTestTypesWithDefaul, EncodedSize(NewTestTypesWithDefault()))
	require.Equal(t, encodedTestTypesOptionalSize, EncodedSize(NewTestTypesOptional()))

	type testcase struct {
		name   string
		update func(p *TestTypesOptional)
		expect int // Encode or EncodedSize
	}

	b := make([]byte, 1024)

	fhdr := fieldHeaderLen
	lhdr := listHeaderLen
	mhdr := mapHeaderLen
	shdr := strHeaderLen

	testcases := []testcase{
		{
			name:   "case_bool",
			update: func(p *TestTypesOptional) { v := true; p.FBool = &v },
			expect: encodedTestTypesOptionalSize + fhdr + 1,
		},
		{
			name:   "case_string",
			update: func(p *TestTypesOptional) { v := "str"; p.String_ = &v },
			expect: encodedTestTypesOptionalSize + fhdr + shdr + 3,
		},
		{
			name:   "case_map_with_primitive_types",
			update: func(p *TestTypesOptional) { p.M0 = map[int32]int32{1: 2, 3: 4} },
			expect: encodedTestTypesOptionalSize + fhdr + mhdr + 2*(4+4),
		},
		{
			name:   "case_map_with_i32_string",
			update: func(p *TestTypesOptional) { p.M1 = map[int32]string{1: "2", 2: "3"} },
			expect: encodedTestTypesOptionalSize + fhdr + mhdr + 2*(4+(shdr+1)),
		},
		{
			name:   "case_map_with_string_struct",
			update: func(p *TestTypesOptional) { p.M3 = map[string]*Msg{"1": nil, "2": &Msg{Type: 3}} }, // 36
			expect: encodedTestTypesOptionalSize + fhdr + mhdr + (shdr + 1 + 1) + (shdr + 1 + encodedMsgSize),
		},
		{
			name:   "case_map_with_i32_list",
			update: func(p *TestTypesOptional) { p.ML = map[int32][]int32{1: []int32{1, 2}, 2: []int32{3, 4}} },
			expect: encodedTestTypesOptionalSize + fhdr + mhdr + 2*(4+lhdr+2*4),
		},
		{
			name:   "case_list_with_i32",
			update: func(p *TestTypesOptional) { p.L0 = []int32{1, 2} },
			expect: encodedTestTypesOptionalSize + fhdr + lhdr + 2*4,
		},
		{
			name:   "case_list_with_string",
			update: func(p *TestTypesOptional) { p.L1 = []string{"1", "2"} },
			expect: encodedTestTypesOptionalSize + fhdr + lhdr + 2*(shdr+1),
		},
		{
			name:   "case_list_with_struct",
			update: func(p *TestTypesOptional) { p.L2 = []*Msg{{Type: 1}, {Type: 2}} },
			expect: encodedTestTypesOptionalSize + fhdr + lhdr + 2*encodedMsgSize,
		},
		{
			name:   "case_list_with_map",
			update: func(p *TestTypesOptional) { p.LM = []map[int32]int32{map[int32]int32{1: 2}} },
			expect: encodedTestTypesOptionalSize + fhdr + lhdr + mhdr + 4 + 4,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestTypesOptional()
			tc.update(p)
			assert.Equal(t, tc.expect, EncodedSize(p))
			n, err := Encode(b, p)
			if assert.NoError(t, err) {
				assert.Equal(t, tc.expect, n)
			}
		})
	}
}

func TestEncodeStructOther(t *testing.T) {
	assert.Equal(t, encodedMsgSize, EncodedSize(Msg{})) // indirect type
	assert.Equal(t, 1, EncodedSize((*Msg)(nil)))        // nil
}
