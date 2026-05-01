package radix

import "testing"

func TestRadixLongestPrefix(t *testing.T) {
	tree := New(map[string]string{
		"demo":                      "all",
		"demo/prod":                 "prod",
		"demo/prod/catalog-service": "catalog",
	})
	got, ok := tree.LongestPrefix("demo/prod/catalog-service/api")
	if !ok || got != "catalog" {
		t.Fatalf("expected catalog route, got %q ok=%v", got, ok)
	}
	got, ok = tree.LongestPrefix("demo/prod/other")
	if !ok || got != "prod" {
		t.Fatalf("expected prod route, got %q ok=%v", got, ok)
	}
}

func TestRadixNilReceiverIsSafe(t *testing.T) {
	var tree *Radix[string]
	got, ok := tree.LongestPrefix("demo/prod")
	if ok || got != "" {
		t.Fatalf("expected empty miss, got %q ok=%v", got, ok)
	}
}
