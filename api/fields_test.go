package main

import "testing"

func TestMatchFields(t *testing.T) {
	words := []ocrWord{
		{Text: "Nom", Box: box{X: 10, Y: 20, Width: 40, Height: 12}},
		{Text: "Prénom", Box: box{X: 10, Y: 40, Width: 60, Height: 12}},
		{Text: "Date", Box: box{X: 10, Y: 60, Width: 30, Height: 12}},
		{Text: "de", Box: box{X: 42, Y: 60, Width: 16, Height: 12}},
		{Text: "naissance", Box: box{X: 60, Y: 60, Width: 70, Height: 12}},
	}
	got := matchFields(words)
	want := map[string]box{
		"nom":            {10, 20, 40, 12},
		"prenom":         {10, 40, 60, 12},
		"date_naissance": {10, 60, 120, 12},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d fields, want %d: %+v", len(got), len(want), got)
	}
	for _, f := range got {
		b, ok := want[f.Name]
		if !ok {
			t.Fatalf("unexpected field %q", f.Name)
		}
		if f.Box != b {
			t.Fatalf("%s box = %+v, want %+v", f.Name, f.Box, b)
		}
	}
}

func TestNormalize(t *testing.T) {
	if got := normalize("Prénom :"); got != "prenom" {
		t.Fatalf("got %q", got)
	}
}
