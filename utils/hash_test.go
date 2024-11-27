package utils

import "testing"

func TestCompute(t *testing.T) {
	s := ComputeCapacity(15)
	t.Log(s)
	//
	t.Log(s >> 1)
	s = s | s>>1
	t.Log(1 & 1)
	t.Log(s)
}

// fnv1a
func TestFnv(t *testing.T) {
	t.Log(Fnv32("abc1"))
}
