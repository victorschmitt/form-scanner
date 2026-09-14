package main

import (
	"context"
	"fmt"
	"math"

	vision "cloud.google.com/go/vision/v2/apiv1"
	pb "cloud.google.com/go/vision/v2/apiv1/visionpb"
)

type ocrWord struct {
	Text string
	Box  box
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
	var out []ocrWord
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
					out = append(out, ocrWord{Text: text, Box: polyBox(word.BoundingBox)})
				}
			}
		}
	}
	return out
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
