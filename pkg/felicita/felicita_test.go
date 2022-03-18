package felicita

import (
	"testing"
)

func TestInit(t *testing.T) {
	f, err := New()
	if err == nil {
		t.Fatalf("instantiation of scale was unexpectedly successful")
	}
	if f != nil {
		t.Fatalf("instantiation of scale unexpectedly returned non-nil instance")
	}
}
