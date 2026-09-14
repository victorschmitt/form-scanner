package main

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

type labelPattern struct {
	name     string
	tokens   []string
	tokenLen int
}

var fieldPatterns = buildPatterns([]struct {
	name    string
	phrases []string
}{
	{"nom", []string{"nom", "nom de famille"}},
	{"prenom", []string{"prenom", "prenoms"}},
	{"date_naissance", []string{"date de naissance", "ne le"}},
	{"lieu_naissance", []string{"lieu de naissance", "ne a"}},
	{"adresse", []string{"adresse", "domicile"}},
	{"code_postal", []string{"code postal", "cp"}},
	{"ville", []string{"ville", "commune"}},
	{"telephone", []string{"telephone", "tel", "portable"}},
	{"email", []string{"email", "e mail", "courriel", "mail"}},
	{"nationalite", []string{"nationalite"}},
	{"sexe", []string{"sexe", "genre"}},
	{"signature", []string{"signature"}},
})

func buildPatterns(defs []struct {
	name    string
	phrases []string
}) []labelPattern {
	var out []labelPattern
	for _, d := range defs {
		for _, p := range d.phrases {
			toks := strings.Fields(normalize(p))
			if len(toks) == 0 {
				continue
			}
			out = append(out, labelPattern{name: d.name, tokens: toks, tokenLen: len(toks)})
		}
	}
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].tokenLen > out[i].tokenLen {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

func matchFields(words []ocrWord) []field {
	norms := make([]string, len(words))
	for i, w := range words {
		norms[i] = normalize(w.Text)
	}
	used := make([]bool, len(words))
	var fields []field

	for _, p := range fieldPatterns {
		for i := 0; i <= len(words)-p.tokenLen; i++ {
			ok := true
			for k, tok := range p.tokens {
				if used[i+k] || norms[i+k] != tok {
					ok = false
					break
				}
			}
			if !ok {
				continue
			}
			b := words[i].Box
			raw := words[i].Text
			used[i] = true
			for k := 1; k < p.tokenLen; k++ {
				b = unionBox(b, words[i+k].Box)
				raw += " " + words[i+k].Text
				used[i+k] = true
			}
			fields = append(fields, field{Name: p.name, MatchedText: raw, Box: b})
			break
		}
	}
	return fields
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
