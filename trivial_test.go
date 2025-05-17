package xaws

import "testing"

func TestKBConstant(t *testing.T) {
	if KB != 1024 {
		t.Errorf("KB constant is %d, want 1024", KB)
	}
}
