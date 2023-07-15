package pkg_test

import (
	"testing"

	"github.com/cicconee/go-cicd-example/pkg"
)

func TestIncrement(t *testing.T) {
	a1 := 0
	a2 := pkg.Increment(a1)

	if a2 != 1 {
		t.Errorf("expected 1; got %d", a2)
	}
}
