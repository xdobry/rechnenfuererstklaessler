package main

import (
	"image"
	"image/color"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

type Abacus struct {
	red  int
	blue int
	red2 int
}

func (exercise *Exercise) displayAbacus() {
	exercise.showAbacus = true
	switch exercise.task_type {
	case Add10:
		fallthrough
	case Add20:
		fallthrough
	case Sub10:
		exercise.abacus1.setRedBlueRed(exercise.task.parameter1, 0, 0)
		exercise.abacus2.setRedBlueRed(0, exercise.task.parameter2, 0)
	case AddQuest10:
		fallthrough
	case AddQuest20:
		exercise.displayAbacusMinus(exercise.task.task_result, exercise.task.parameter1)
	case Sub20:
		exercise.displayAbacusMinus(exercise.task.parameter1, exercise.task.task_result)
	case SubQuest10:
		fallthrough
	case SubQuest20:
		exercise.displayAbacusMinus(exercise.task.parameter1, exercise.task.task_result)
	}
}

func (exercise *Exercise) abacusHelp() {
	exercise.hasAbacusHelp = true
	switch exercise.task_type {
	case Sub10:
		exercise.displayAbacusMinus(exercise.task.parameter1, exercise.task.task_result)
		return
	case Add10:
	case Add20:
	default:
		exercise.displayAbacus()
		return
	}
	if exercise.task.parameter1 <= 5 && exercise.task.parameter2 <= 5 {
		if exercise.task.parameter1 >= exercise.task.parameter2 {
			toMove := min(5-exercise.task.parameter1, exercise.task.parameter2)
			exercise.abacus1.setRedBlueRed(exercise.task.parameter1, toMove, 0)
			exercise.abacus2.setRedBlueRed(0, exercise.task.parameter2-toMove, 0)
		} else {
			toMove := min(5-exercise.task.parameter2, exercise.task.parameter1)
			exercise.abacus1.setRedBlueRed(exercise.task.parameter1-toMove, 0, 0)
			exercise.abacus2.setRedBlueRed(0, exercise.task.parameter2, toMove)
		}
	} else {
		if exercise.task.parameter1 >= exercise.task.parameter2 {
			toMove := min(10-exercise.task.parameter1, exercise.task.parameter2)
			exercise.abacus1.setRedBlueRed(exercise.task.parameter1, toMove, 0)
			exercise.abacus2.setRedBlueRed(0, exercise.task.parameter2-toMove, 0)
		} else {
			toMove := min(10-exercise.task.parameter2, exercise.task.parameter1)
			exercise.abacus1.setRedBlueRed(exercise.task.parameter1-toMove, 0, 0)
			exercise.abacus2.setRedBlueRed(0, exercise.task.parameter2, toMove)
		}
	}
}

func (abacus *Abacus) setRedBlueRed(red int, blue int, red2 int) {
	abacus.red = red
	abacus.blue = blue
	abacus.red2 = red2
}

func (exercise *Exercise) displayAbacusMinus(sum int, par int) {
	exercise.abacus1.setRedBlueRed(par, min(sum-par, 10-par), 0)
	exercise.abacus2.setRedBlueRed(0, max(sum-10, 0), 0)
}

func drawAbacus(gtx layout.Context, exercise *Exercise) layout.Dimensions {
	gap := gtx.Metric.Dp(unit.Dp(8))
	radius := gtx.Metric.Dp(unit.Dp(32))
	size := image.Pt(10*radius+9*gap+gap, radius*2+gap)
	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
	event.Op(gtx.Ops, &exercise.abacusClick)
	for {
		ev, ok := gtx.Source.Event(pointer.Filter{
			Target: &exercise.abacusClick,
			Kinds:  pointer.Press | pointer.Release,
		})
		if !ok {
			break
		}
		if x, ok := ev.(pointer.Event); ok {
			switch x.Kind {
			case pointer.Press:
				exercise.abacusClick = true
			case pointer.Release:
				exercise.abacusClick = false
			}
		}
	}

	colorBlue := color.NRGBA{B: 0xFF, A: 0xFF}
	colorRed := color.NRGBA{R: 0xFF, A: 0xFF}
	colorBlack := color.NRGBA{A: 0xFF}
	drawAbacusRow := func(abacus *Abacus, yoffset int) {
		for x := 0; x < 10; x++ {
			var color color.NRGBA
			var addGap = 0
			if x >= 5 {
				addGap = gap
			}
			offset := op.Offset(image.Point{X: x*(radius+gap) + addGap, Y: yoffset}).Push(gtx.Ops)
			circle := clip.Ellipse{Max: image.Pt(radius, radius)}
			area := circle.Push(gtx.Ops)
			var fill = true
			if x < abacus.red {
				color = colorRed
			} else if x < abacus.blue+abacus.red {
				color = colorBlue
			} else if x < abacus.blue+abacus.red+abacus.red2 {
				color = colorRed
			} else {
				color = colorBlack
				fill = false
			}
			if fill {
				paint.ColorOp{Color: color}.Add(gtx.Ops)
				paint.PaintOp{}.Add(gtx.Ops)
			} else {
				paint.FillShape(gtx.Ops, colorBlack,
					clip.Stroke{
						Path:  circle.Path(gtx.Ops),
						Width: 4,
					}.Op(),
				)
			}
			area.Pop()
			offset.Pop()
		}
	}
	drawAbacusRow(&exercise.abacus1, 0)
	drawAbacusRow(&exercise.abacus2, radius+gap)
	return layout.Dimensions{Size: size}
}
