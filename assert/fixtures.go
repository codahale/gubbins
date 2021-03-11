package assert

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
)

// EqualFixture validates that the contents of the given file are the same as the given bytes. If the
// OVERWRITE environment variable is set to TRUE, the given bytes are written to the file first.
func EqualFixture(tb testing.TB, name, filename string, got []byte) {
	tb.Helper()

	if ok, _ := strconv.ParseBool(os.Getenv("OVERWRITE")); ok {
		tb.Logf("overwriting %s", filename)

		err := ioutil.WriteFile(filename, got, 0o644) //nolint:gosec // not used in main code
		if err != nil {
			tb.Fatal(err)
		}
	}

	want, err := ioutil.ReadFile(filename)
	if err != nil {
		tb.Fatal(err)
	}

	Equal(tb, fmt.Sprintf("%s/%s", name, filename), want, got)
}
