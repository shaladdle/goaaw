package vector

import (
	"testing"
)

func TestAppendGetHas(t *testing.T) {
	pairs := []struct {
		key   string
		value int
	}{
		{"a", 1},
		{"p", 2},
		{"o", 3},
		{"k", 4},
		{"l", 5},
		{"j", 6},
		{"e", 7},
		{"d", 8},
		{"x", 9},
		{"z", 10},
	}

	v := New()

	for _, p := range pairs {
		v.Append(p.key, p.value)
		if got := v.Get(p.key).(int); got != p.value {
			t.Errorf("get right after append: got %v, want %v", got, p.value)
		}
	}

	i := 0
	for value := range v.Iter() {
		if got := pairs[i].value; got != value {
			t.Errorf("iter: got %v, want %v", got, pairs[i].value)
		}
		i++
	}

	for _, p := range pairs {
		if !v.Has(p.key) {
			t.Errorf("has returned false for key '%v'", p.key)
		}
	}
}

func TestDelete(t *testing.T) {
	v := New()
	v.Append("a", 123)
	v.Delete("a")
	if v.Has("a") {
		t.Errorf("v.Has returned true, but the element should have been deleted")
	}
}
