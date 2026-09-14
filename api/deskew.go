package main

import (
	"fmt"
	"image"
	"math"

	"gocv.io/x/gocv"
)

func straighten(raw []byte) (jpeg []byte, width, height int, deskewed bool, err error) {
	src, err := gocv.IMDecode(raw, gocv.IMReadColor)
	if err != nil {
		return nil, 0, 0, false, fmt.Errorf("image illisible")
	}
	defer src.Close()
	if src.Empty() {
		return nil, 0, 0, false, fmt.Errorf("image illisible")
	}

	quad, ok := findFormQuad(src)
	out := src
	if ok {
		warped, werr := warp(src, quad)
		if werr != nil {
			return nil, 0, 0, false, werr
		}
		defer warped.Close()
		out = warped
		deskewed = true
	}

	buf, err := gocv.IMEncode(".jpg", out)
	if err != nil {
		return nil, 0, 0, false, fmt.Errorf("encodage de l'image impossible")
	}
	defer buf.Close()

	jpeg = append([]byte(nil), buf.GetBytes()...)
	return jpeg, out.Cols(), out.Rows(), deskewed, nil
}

func findFormQuad(src gocv.Mat) ([]image.Point, bool) {
	const maxW = 800.0
	scale := 1.0
	work := src
	if float64(src.Cols()) > maxW {
		scale = maxW / float64(src.Cols())
		small := gocv.NewMat()
		defer small.Close()
		gocv.Resize(src, &small, image.Point{}, scale, scale, gocv.InterpolationArea)
		work = small
	}

	gray := gocv.NewMat()
	defer gray.Close()
	gocv.CvtColor(work, &gray, gocv.ColorBGRToGray)

	blur := gocv.NewMat()
	defer blur.Close()
	gocv.GaussianBlur(gray, &blur, image.Pt(5, 5), 0, 0, gocv.BorderDefault)

	edges := gocv.NewMat()
	defer edges.Close()
	gocv.Canny(blur, &edges, 50, 150)

	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(5, 5))
	defer kernel.Close()
	gocv.Dilate(edges, &edges, kernel)

	contours := gocv.FindContours(edges, gocv.RetrievalList, gocv.ChainApproxSimple)
	defer contours.Close()

	minArea := float64(work.Cols()*work.Rows()) * 0.2
	bestArea := 0.0
	var best []image.Point

	for _, c := range contours.ToPoints() {
		pv := gocv.NewPointVectorFromPoints(c)
		peri := gocv.ArcLength(pv, true)
		approx := gocv.ApproxPolyDP(pv, 0.02*peri, true)
		pv.Close()
		if approx.Size() == 4 {
			area := gocv.ContourArea(approx)
			if area > minArea && area > bestArea {
				bestArea = area
				best = approx.ToPoints()
			}
		}
		approx.Close()
	}
	if best == nil {
		return nil, false
	}

	if scale != 1 {
		inv := 1 / scale
		for i := range best {
			best[i].X = int(float64(best[i].X) * inv)
			best[i].Y = int(float64(best[i].Y) * inv)
		}
	}
	return orderQuad(best), true
}

func orderQuad(pts []image.Point) []image.Point {
	tl, tr, br, bl := pts[0], pts[0], pts[0], pts[0]
	sumMin, sumMax := math.MaxInt, math.MinInt
	diffMin, diffMax := math.MaxInt, math.MinInt
	for _, p := range pts {
		s := p.X + p.Y
		d := p.X - p.Y
		if s < sumMin {
			sumMin = s
			tl = p
		}
		if s > sumMax {
			sumMax = s
			br = p
		}
		if d > diffMax {
			diffMax = d
			tr = p
		}
		if d < diffMin {
			diffMin = d
			bl = p
		}
	}
	return []image.Point{tl, tr, br, bl}
}

func warp(src gocv.Mat, quad []image.Point) (gocv.Mat, error) {
	tl, tr, br, bl := quad[0], quad[1], quad[2], quad[3]
	w := int(math.Max(dist(tl, tr), dist(bl, br)))
	h := int(math.Max(dist(tl, bl), dist(tr, br)))
	if w < 32 || h < 32 {
		return gocv.Mat{}, fmt.Errorf("formulaire non détecté")
	}

	srcPts := gocv.NewPointVectorFromPoints(quad)
	defer srcPts.Close()
	dstPts := gocv.NewPointVectorFromPoints([]image.Point{
		{0, 0},
		{w - 1, 0},
		{w - 1, h - 1},
		{0, h - 1},
	})
	defer dstPts.Close()

	m := gocv.GetPerspectiveTransform(srcPts, dstPts)
	defer m.Close()

	out := gocv.NewMat()
	gocv.WarpPerspective(src, &out, m, image.Pt(w, h))
	return out, nil
}

func dist(a, b image.Point) float64 {
	dx := float64(a.X - b.X)
	dy := float64(a.Y - b.Y)
	return math.Hypot(dx, dy)
}
