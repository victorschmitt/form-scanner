package main

import "testing"

func TestMatchFields(t *testing.T) {
	words := []ocrWord{
		{Text: "Titre", Box: box{X: 10, Y: 5, Width: 40, Height: 12}, Line: 0},
		{Text: "Nom", Box: box{X: 10, Y: 20, Width: 40, Height: 12}, Line: 1},
		{Text: ":", Box: box{X: 52, Y: 20, Width: 5, Height: 12}, Line: 1},
		{Text: "SCHMITT", Box: box{X: 60, Y: 20, Width: 70, Height: 12}, Line: 1},
		{Text: "Prénom", Box: box{X: 10, Y: 40, Width: 60, Height: 12}, Line: 2},
		{Text: "Victor", Box: box{X: 75, Y: 40, Width: 50, Height: 12}, Line: 2},
		{Text: "Date", Box: box{X: 10, Y: 60, Width: 30, Height: 12}, Line: 3},
		{Text: "de", Box: box{X: 42, Y: 60, Width: 16, Height: 12}, Line: 3},
		{Text: "naissance", Box: box{X: 60, Y: 60, Width: 70, Height: 12}, Line: 3},
		{Text: "05/09/1985", Box: box{X: 135, Y: 60, Width: 80, Height: 12}, Line: 3},
		{Text: "Adresse", Box: box{X: 10, Y: 80, Width: 55, Height: 12}, Line: 4},
		{Text: "postale", Box: box{X: 68, Y: 80, Width: 50, Height: 12}, Line: 4},
		{Text: "5 rue du Clos", Box: box{X: 125, Y: 80, Width: 100, Height: 12}, Line: 4},
		{Text: "21160 Marsannay-la-Côte", Box: box{X: 125, Y: 100, Width: 160, Height: 12}, Line: 5},
		{Text: "Téléphone", Box: box{X: 10, Y: 120, Width: 70, Height: 12}, Line: 6},
		{Text: "06 59 72 36 85", Box: box{X: 85, Y: 120, Width: 100, Height: 12}, Line: 6},
		{Text: "Adresse", Box: box{X: 10, Y: 140, Width: 55, Height: 12}, Line: 7},
		{Text: "Email", Box: box{X: 68, Y: 140, Width: 40, Height: 12}, Line: 7},
		{Text: "victor.schmitt@gmail.com", Box: box{X: 115, Y: 140, Width: 170, Height: 12}, Line: 7},
		{Text: "Signature", Box: box{X: 10, Y: 160, Width: 70, Height: 12}, Line: 8},
	}
	got := matchFields(words)
	want := map[string]string{
		"nom":            "SCHMITT",
		"prenom":         "Victor",
		"date_naissance": "05/09/1985",
		"adresse":        "5 rue du Clos\n21160 Marsannay-la-Côte",
		"telephone":      "06 59 72 36 85",
		"email":          "victor.schmitt@gmail.com",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d fields, want %d: %+v", len(got), len(want), got)
	}
	for _, f := range got {
		value, ok := want[f.Name]
		if !ok {
			t.Fatalf("unexpected field %q", f.Name)
		}
		if f.Value != value {
			t.Errorf("%s value = %q, want %q", f.Name, f.Value, value)
		}
		wantBoxes := 1
		if f.Name == "adresse" {
			wantBoxes = 2
		}
		if len(f.Boxes) != wantBoxes {
			t.Errorf("%s has %d boxes, want %d", f.Name, len(f.Boxes), wantBoxes)
		}
	}
}

func TestNormalize(t *testing.T) {
	if got := normalize("Prénom :"); got != "prenom" {
		t.Fatalf("got %q", got)
	}
}
