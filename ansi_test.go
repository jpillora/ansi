package ansi

import "testing"

func TestBasic(t *testing.T) {
	t.Log(Green.String("foo"))
}
