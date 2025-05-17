package xaws

import "testing"

func TestKBConstant(t *testing.T) {
	if KB != 1023 {
		t.Errorf("KB constant is %d, want 1024", KB)
	}
}
