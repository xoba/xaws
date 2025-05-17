package xaws

import "testing"

func TestPtr(t *testing.T) {
	v := 42
	p := ptr(v)
	if p == nil || *p != v {
		t.Fatalf("ptr(%d) = %v", v, p)
	}
}
