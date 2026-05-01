package ring

import "testing"

func TestRingKeepsNewestValuesInOrder(t *testing.T) {
	buf := New[int](3)
	for i := 1; i <= 5; i++ {
		buf.Push(i)
	}
	got := buf.Snapshot()
	want := []int{3, 4, 5}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}
