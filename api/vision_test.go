package main

import (
	"math"
	"strings"
	"testing"
)

// skewedWord place un mot sur une rangée inclinée: la rangée y0 est définie
// avant rotation, puis tournée de angle autour de l'origine.
func skewedWord(text string, x, y0, width, height, angle float64) placedWord {
	sin, cos := math.Sincos(angle)
	cx := x + width/2
	return placedWord{
		ocrWord: ocrWord{Text: text},
		cx:      cx*cos - y0*sin,
		cy:      cx*sin + y0*cos,
		height:  height,
		dx:      width * cos,
		dy:      width * sin,
	}
}

func TestGroupRowsMergesColumnsOnSkewedPage(t *testing.T) {
	const angle = 6 * math.Pi / 180
	words := []placedWord{
		// Libellé à gauche et valeur à droite: même rangée malgré l'écart.
		skewedWord("Email", 60, 150, 120, 35, angle),
		skewedWord(":", 190, 150, 15, 30, angle),
		skewedWord("nom.prenom@gmail.com", 600, 150, 580, 35, angle),
		skewedWord("Nom", 60, 300, 95, 35, angle),
		skewedWord("DUPONT", 600, 300, 175, 35, angle),
	}

	got := groupRows(words)
	if len(got) != len(words) {
		t.Fatalf("got %d words, want %d", len(got), len(words))
	}

	lines := map[int][]string{}
	for _, w := range got {
		lines[w.Line] = append(lines[w.Line], w.Text)
	}
	want := map[int]string{
		0: "Email : nom.prenom@gmail.com",
		1: "Nom DUPONT",
	}
	if len(lines) != len(want) {
		t.Fatalf("got %d lines, want %d: %v", len(lines), len(want), lines)
	}
	for line, text := range want {
		if joined := strings.Join(lines[line], " "); joined != text {
			t.Errorf("line %d = %q, want %q", line, joined, text)
		}
	}
}

func TestGroupRowsKeepsDistinctRowsApart(t *testing.T) {
	words := []placedWord{
		skewedWord("Nom", 60, 100, 95, 35, 0),
		skewedWord("Prenom", 60, 140, 150, 35, 0),
	}
	got := groupRows(words)
	if got[0].Line == got[1].Line {
		t.Fatalf("rangées fusionnées à tort: %+v", got)
	}
}
