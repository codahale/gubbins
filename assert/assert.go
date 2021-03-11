// Package assert contains assorted functions for asserting test invariants.
package assert

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Equal validates that the want and got variables are equal using go-cmp. If they are different, an
// informative error is registered in the given test context.
func Equal(tb testing.TB, name string, want, got interface{}, opts ...cmp.Option) {
	tb.Helper()

	if diff := cmp.Diff(want, got, opts...); diff != "" {
		tb.Errorf("%s mismatch (-want +got):\n%s", name, diff)
	}
}
