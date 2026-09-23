package main

import (
	"context"
	"fmt"
	"math"
	"sort"

	vision "cloud.google.com/go/vision/v2/apiv1"
	pb "cloud.google.com/go/vision/v2/apiv1/visionpb"
)

type ocrWord struct {
	Text string `json:"text"`
	Box  box    `json:"box"`
	Line int    `json:"line"`
}

func detectWords(ctx context.Context, client *vision.ImageAnnotatorClient, jpeg []byte) ([]ocrWord, error) {
	resp, err := client.BatchAnnotateImages(ctx, &pb.BatchAnnotateImagesRequest{
		Requests: []*pb.AnnotateImageRequest{{
			Image:    &pb.Image{Content: jpeg},
			Features: []*pb.Feature{{Type: pb.Feature_DOCUMENT_TEXT_DETECTION}},
		}},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Responses) == 0 {
		return nil, nil
	}
	r := resp.Responses[0]
	if e := r.GetError(); e != nil && e.Code != 0 {
		return nil, fmt.Errorf("vision: %s", e.Message)
	}
	return wordsFromAnnotation(r.GetFullTextAnnotation()), nil
}

func wordsFromAnnotation(ann *pb.TextAnnotation) []ocrWord {
	if ann == nil {
		return nil
	}
	var words []placedWord
	for _, page := range ann.Pages {
		for _, block := range page.Blocks {
			for _, para := range block.Paragraphs {
				for _, word := range para.Words {
					text := ""
					for _, s := range word.Symbols {
						text += s.Text
					}
					if text == "" {
						continue
					}
					words = append(words, place(text, word.BoundingBox))
				}
			}
		}
	}
	return groupRows(words)
}

// rowTolerance: écart vertical toléré au sein d'une rangée, en fraction de la
// hauteur médiane des mots.
const rowTolerance = 0.6

type placedWord struct {
	ocrWord
	cx, cy float64
	height float64
	dx, dy float64 // arête supérieure, orientée dans le sens de lecture
}

func place(text string, p *pb.BoundingPoly) placedWord {
	b := polyBox(p)
	w := placedWord{
		ocrWord: ocrWord{Text: text, Box: b},
		cx:      float64(b.X) + float64(b.Width)/2,
		cy:      float64(b.Y) + float64(b.Height)/2,
		height:  float64(b.Height),
	}
	if p != nil && len(p.Vertices) == 4 {
		v := p.Vertices
		w.dx = float64(v[1].X - v[0].X)
		w.dy = float64(v[1].Y - v[0].Y)
		w.height = math.Hypot(float64(v[3].X-v[0].X), float64(v[3].Y-v[0].Y))
	}
	return w
}

// groupRows annule l'inclinaison médiane du texte, puis découpe les mots en
// rangées le long de l'axe perpendiculaire aux lignes.
func groupRows(words []placedWord) []ocrWord {
	if len(words) == 0 {
		return nil
	}

	// Somme vectorielle des arêtes: les mots longs, dont l'orientation est la
	// plus fiable, pèsent naturellement plus que les mots d'un caractère.
	var sumX, sumY float64
	heights := make([]float64, len(words))
	for i, w := range words {
		sumX += w.dx
		sumY += w.dy
		heights[i] = w.height
	}
	angle := math.Atan2(sumY, sumX)
	if math.Abs(angle) > math.Pi/4 {
		angle = 0
	}
	tol := rowTolerance * median(heights)

	sin, cos := math.Sincos(angle)
	along := make([]float64, len(words))
	depth := make([]float64, len(words))
	order := make([]int, len(words))
	for i, w := range words {
		along[i] = w.cx*cos + w.cy*sin
		depth[i] = w.cy*cos - w.cx*sin
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool { return depth[order[a]] < depth[order[b]] })

	out := make([]ocrWord, 0, len(words))
	var row []int
	rowDepth := 0.0
	line := 0
	flush := func() {
		sort.SliceStable(row, func(a, b int) bool { return along[row[a]] < along[row[b]] })
		for _, i := range row {
			w := words[i].ocrWord
			w.Line = line
			out = append(out, w)
		}
		line++
		row = row[:0]
	}
	for _, i := range order {
		if len(row) > 0 && depth[i]-rowDepth > tol {
			flush()
		}
		if len(row) == 0 {
			rowDepth = depth[i]
		}
		row = append(row, i)
	}
	flush()
	return out
}

func median(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	s := append([]float64(nil), v...)
	sort.Float64s(s)
	return s[len(s)/2]
}

func polyBox(p *pb.BoundingPoly) box {
	if p == nil || len(p.Vertices) == 0 {
		return box{}
	}
	minX, minY := math.MaxInt32, math.MaxInt32
	maxX, maxY := 0, 0
	for _, v := range p.Vertices {
		x, y := int(v.X), int(v.Y)
		if x < minX {
			minX = x
		}
		if y < minY {
			minY = y
		}
		if x > maxX {
			maxX = x
		}
		if y > maxY {
			maxY = y
		}
	}
	return box{X: minX, Y: minY, Width: maxX - minX, Height: maxY - minY}
}

func unionBox(a, b box) box {
	minX := min(a.X, b.X)
	minY := min(a.Y, b.Y)
	maxX := max(a.X+a.Width, b.X+b.Width)
	maxY := max(a.Y+a.Height, b.Y+b.Height)
	return box{X: minX, Y: minY, Width: maxX - minX, Height: maxY - minY}
}
