package reflect

import (
	"bytes"
	"go/format"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenThrift(t *testing.T) {
	fd, err := createOrGetFieldDesc(reflect.TypeOf(TestTypesForBenchmark{}))
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	buf.WriteString(`
package reflect

import (
  "encoding/binary"
  "unsafe"
)

`)

	g := ThriftCodeGenerator{}
	g.GenEncode(buf, fd)

	fd, err = createOrGetFieldDesc(reflect.TypeOf(Msg{}))
	require.NoError(t, err)

	g.GenEncode(buf, fd)

	b, err := format.Source(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile("gen_testdata_test.go", b, 0644)
}
