package main

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

type labelPattern struct {
	name       string
	tokens     []string
	extraLines int
}

var fieldPatterns = buildPatterns([]struct {
	name       string
	phrases    []string
	extraLines int
}{
	{"nom", []string{"nom", "nom de famille"}, 0},
	{"prenom", []string{"prenom", "prenoms"}, 0},
	{"date_naissance", []string{"date de naissance", "ne le"}, 0},
	{"adresse", []string{"adresse postale", "adresse", "domicile"}, 1},
	{"telephone", []string{"telephone", "tel", "portable"}, 0},
	{"email", []string{"adresse email", "email", "e mail", "courriel", "mail"}, 0},
})

func buildPatterns(defs []struct {
	name       string
	phrases    []string
	extraLines int
}) []labelPattern {
	var out []labelPattern
	for _, d := range defs {
		for _, p := range d.phrases {
			toks := strings.Fields(normalize(p))
			if len(toks) == 0 {
				continue
			}
			out = append(out, labelPattern{name: d.name, tokens: toks, extraLines: d.extraLines})
		}
	}
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if len(out[j].tokens) > len(out[i].tokens) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

func matchFields(words []ocrWord) []field {
	lines := make([][]ocrWord, 0)
	for _, word := range words {
		for len(lines) <= word.Line {
			lines = append(lines, nil)
		}
		lines[word.Line] = append(lines[word.Line], word)
	}

	seen := make(map[string]bool)
	var fields []field
	for lineIndex, line := range lines {
		pattern, valueStart, ok := matchLabel(line)
		if !ok || seen[pattern.name] {
			continue
		}

		valueWords := line[valueStart:]
		boxes := []box{lineBox(line)}
		for extra := 1; extra <= pattern.extraLines && lineIndex+extra < len(lines); extra++ {
			next := lines[lineIndex+extra]
			if len(next) == 0 {
				continue
			}
			if _, _, isLabel := matchLabel(next); isLabel {
				break
			}
			valueWords = append(valueWords, ocrWord{Text: "\n"})
			valueWords = append(valueWords, next...)
			boxes = append(boxes, lineBox(next))
		}

		value := joinValue(valueWords)
		if value == "" {
			continue
		}
		fields = append(fields, field{
			Name:  pattern.name,
			Value: value,
			Boxes: boxes,
		})
		seen[pattern.name] = true
	}
	return fields
}

func matchLabel(line []ocrWord) (labelPattern, int, bool) {
	for _, pattern := range fieldPatterns {
		token := 0
		for i, word := range line {
			normalized := normalize(word.Text)
			if normalized == "" {
				continue
			}
			if token >= len(pattern.tokens) || normalized != pattern.tokens[token] {
				break
			}
			token++
			if token == len(pattern.tokens) {
				return pattern, skipSeparators(line, i+1), true
			}
		}
	}
	return labelPattern{}, 0, false
}

func skipSeparators(line []ocrWord, start int) int {
	for start < len(line) && normalize(line[start].Text) == "" {
		start++
	}
	return start
}

func lineBox(line []ocrWord) box {
	b := line[0].Box
	for _, word := range line[1:] {
		b = unionBox(b, word.Box)
	}
	return b
}

func joinValue(words []ocrWord) string {
	var lines []string
	var current []string
	for _, word := range words {
		if word.Text == "\n" {
			if len(current) > 0 {
				lines = append(lines, strings.Join(current, " "))
				current = nil
			}
			continue
		}
		current = append(current, word.Text)
	}
	if len(current) > 0 {
		lines = append(lines, strings.Join(current, " "))
	}
	return strings.Join(lines, "\n")
}

func normalize(s string) string {
	s = strings.ToLower(norm.NFD.String(s))
	var b strings.Builder
	for _, r := range s {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else if unicode.IsSpace(r) {
			b.WriteByte(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
