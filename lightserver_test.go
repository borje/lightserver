package main

import "testing"

func Test(t *testing.T) {
	t.Errorf("asdaf")
	t.AssertEq(1, 2)
}
