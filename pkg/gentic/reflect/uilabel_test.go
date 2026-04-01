package reflect

import "testing"

func TestReflectUILabelTotal(t *testing.T) {
	tests := []struct {
		maxReflection int
		want          int
	}{
		{-1, 1},
		{0, 1},
		{1, 1},
		{2, 3},
		{3, 4},
	}
	for _, tt := range tests {
		if g := ReflectUILabelTotal(tt.maxReflection); g != tt.want {
			t.Fatalf("ReflectUILabelTotal(%d) = %d, want %d", tt.maxReflection, g, tt.want)
		}
	}
}

func TestReflectUILabelPair(t *testing.T) {
	cur, tot := reflectUILabelPair(0, 1)
	if cur != 1 || tot != 1 {
		t.Fatalf("pair(0,1) = (%d,%d), want (1,1)", cur, tot)
	}
	cur, tot = reflectUILabelPair(1, 1)
	if cur != 1 || tot != 1 {
		t.Fatalf("pair(1,1) = (%d,%d), want (1,1)", cur, tot)
	}
	cur, tot = reflectUILabelPair(0, 2)
	if cur != 1 || tot != 3 {
		t.Fatalf("pair(0,2) = (%d,%d), want (1,3)", cur, tot)
	}
	cur, tot = reflectUILabelPair(2, 2)
	if cur != 3 || tot != 3 {
		t.Fatalf("pair(2,2) = (%d,%d), want (3,3)", cur, tot)
	}
}
