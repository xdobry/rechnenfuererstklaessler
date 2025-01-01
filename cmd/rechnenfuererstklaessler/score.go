package main

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

func drawResult(gtx layout.Context, exercise *Exercise) layout.Dimensions {
	const height = 32
	const width = 32
	const gap = 8
	mygap := 0
	colorYellow := color.NRGBA{R: 0xFF, G: 0xFF, A: 0xFF}
	colorRed := color.NRGBA{R: 0xFF, A: 0xFF}
	colorBlack := color.NRGBA{A: 0xFF}
	if len(exercise.resultHistory) > 0 {
		mygap = gap * (len(exercise.resultHistory) - 1)
	}
	size := image.Pt(len(exercise.resultHistory)*width+mygap, height)
	for index, score := range exercise.resultHistory {
		offset := op.Offset(image.Point{X: index * (width + gap)}).Push(gtx.Ops)
		if score == score_perfect {
			drawStar(gtx, width/2, width/2, width/2, 5, colorYellow)
		} else if score == score_second {
			drawStar(gtx, width/2, width/2, width/2, 5, colorBlack)
		} else {
			drawDiagonalCross(gtx, width, height, colorRed)
		}
		offset.Pop()
	}

	return layout.Dimensions{Size: size}
}

func drawStar(gtx layout.Context, cx, cy, radius float32, points int, col color.NRGBA) {
	if points < 5 {
		points = 5 // Ensure at least a star shape
	}

	// Create a path
	var path clip.Path
	path.Begin(gtx.Ops)

	// Calculate the star's points
	innerRadius := radius / 2 // Inner radius of the star
	angleStep := math.Pi / float32(points)

	for i := 0; i < points*2; i++ {
		angle := float32(i) * angleStep
		r := radius
		if i%2 == 1 {
			r = innerRadius // Alternate between outer and inner radius
		}

		x := cx + float32(float64(r)*math.Cos(float64(angle)))
		y := cy + float32(float64(r)*math.Sin(float64(angle)))

		if i == 0 {
			path.MoveTo(f32.Pt(x, y)) // Move to the first point
		} else {
			path.LineTo(f32.Pt(x, y)) // Draw a line to the next point
		}
	}

	path.Close() // Close the path

	// Fill the star
	paint.FillShape(gtx.Ops, col, clip.Outline{Path: path.End()}.Op())
}

func drawDiagonalCross(gtx layout.Context, width, height int, col color.NRGBA) {
	// Convert width and height to float32 for Gio operations
	w := float32(width)
	h := float32(height)

	// Create a new path
	var path clip.Path
	path.Begin(gtx.Ops)

	// Draw the first diagonal (top-left to bottom-right)
	path.MoveTo(f32.Pt(0, 0))
	path.LineTo(f32.Pt(w, h))

	// Draw the second diagonal (top-right to bottom-left)
	path.MoveTo(f32.Pt(w, 0))
	path.LineTo(f32.Pt(0, h))

	// Stroke the paths
	outline := clip.Stroke{
		Path:  path.End(),
		Width: 4, // Thickness of the lines
	}.Op()

	// Fill the shape with the given color
	paint.FillShape(gtx.Ops, col, outline)
}
