package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math/rand/v2"
	"os"
	"strconv"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/go-mp3"
)

//go:embed resources/title.png
var pngData []byte

//go:embed resources/begin.mp3
var beginMp3 []byte

//go:embed resources/finish.mp3
var finishMp3 []byte

//go:embed resources/neg1.mp3
var neg1Mp3 []byte

//go:embed resources/neg2.mp3
var neg2Mp3 []byte

//go:embed resources/neg3.mp3
var neg3Mp3 []byte

//go:embed resources/pos1.mp3
var pos1Mp3 []byte

//go:embed resources/pos2.mp3
var pos2Mp3 []byte

//go:embed resources/pos3.mp3
var pos3Mp3 []byte

//go:embed resources/applause.mp3
var applauseMp3 []byte

const (
	AUDIO_BEGIN = iota
	AUDIO_FINISH
	AUDIO_NEG1
	AUDIO_NEG2
	AUDIO_NEG3
	AUDIO_POS1
	AUDIO_POS2
	AUDIO_POS3
	AUDIO_APPLOAUSE
)

func main() {
	audio_chan := make(chan int)
	go playAudio(audio_chan)
	go func() {
		window := new(app.Window)
		window.Option(app.Title("Rechnen für Erstklässler"))
		err := run(window, audio_chan)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

const (
	TIME_DISPLAY_ABACUS = time.Second * 5
	EXERCISES_COUNT     = 9
)

var exercisesTitles = [EXERCISES_COUNT]string{"1+2", "6+7", "1+?=4", "6+?=14", "5-2", "16-7", "10-?=4", "14-?=6", "Mix"}

type ScoreType uint8

const (
	score_perfect = 0
	score_second  = 1
	score_wrong   = 3
)

type Exercise struct {
	task_type     TaskType
	task          Task
	tries         int
	okResults     int
	wrongResults  int
	userResult    int
	resultTime    int64
	startTaskTime int64
	numberButtons [10]widget.Clickable
	eraseButton   widget.Clickable
	homeButton    widget.Clickable
	resultHistory []ScoreType
	abacus1       Abacus
	abacus2       Abacus
	hasAbacusHelp bool
	showAbacus    bool
	abacusClick   bool
	showDialog    bool
	dialogMessage string
	dialogClick   widget.Clickable
}

func playAudio(audio_chan chan int) {
	oto_op := &oto.NewContextOptions{}
	oto_op.SampleRate = 44100
	oto_op.ChannelCount = 2
	oto_op.Format = oto.FormatSignedInt16LE

	audio_bytes := [...][]byte{beginMp3, finishMp3, neg1Mp3, neg2Mp3, neg3Mp3, pos1Mp3, pos2Mp3, pos3Mp3, applauseMp3}

	context, readyChan, err := oto.NewContext(oto_op)
	if err != nil {
		panic(err)
	}
	<-readyChan

	var players [len(audio_bytes)]*oto.Player
	// TODO it decode the mp3 again and again during every play
	// so it has performance issue. But Seek and Reset does not work
	// I could not find any better solution for quick
	for index := range audio_chan {
		var myplayer = players[index]
		if myplayer == nil || !myplayer.IsPlaying() {
			decodedMp3, err := mp3.NewDecoder(bytes.NewReader(audio_bytes[index]))
			if err != nil {
				panic("mp3.NewDecoder failed: " + err.Error())
			}
			myplayer = context.NewPlayer(decodedMp3)
			players[index] = myplayer
			myplayer.Play()
		}
	}
}

func run(window *app.Window, audio_chan chan int) error {
	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		log.Fatalf("failed to decode PNG: %v", err)
	}
	tex := paint.NewImageOp(img)

	theme := material.NewTheme()
	theme.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	var ops op.Ops
	var exerciseClickables [EXERCISES_COUNT]widget.Clickable
	var isTitle = true
	var excercise Exercise
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			if isTitle {
				for i := 0; i < len(exerciseClickables); i++ {
					if exerciseClickables[i].Clicked(gtx) {
						isTitle = false
						excercise.startExercise(TaskType(i), window, gtx)
						audio_chan <- AUDIO_BEGIN
						break
					}
				}
				title := material.H4(theme, "Rechnen für Erstklässler")
				title.Alignment = text.Middle
				layout.Inset{
					Top:    unit.Dp(16),
					Bottom: unit.Dp(16),
					Right:  unit.Dp(16),
					Left:   unit.Dp(16),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								tex.Add(gtx.Ops)
								paint.PaintOp{}.Add(gtx.Ops)
								return layout.Dimensions{Size: img.Bounds().Size()}
							})
						},
						),
						layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
						layout.Rigid(title.Layout),
						layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							var children []layout.FlexChild
							for i := 0; i < len(exerciseClickables); i++ {
								children = append(children, layout.Rigid(material.Button(theme, &exerciseClickables[i], exercisesTitles[i]).Layout))
							}
							return layout.Flex{Spacing: layout.SpaceAround}.Layout(gtx, children...)
						}))
				})
			} else {
				if exerciseUi(&excercise, gtx, theme, window, audio_chan) {
					isTitle = true
				}
			}
			e.Frame(gtx.Ops)
		}
	}
}

func randRange(min int, max int) int {
	return rand.IntN(max-min+1) + min
}

func exerciseUi(exercise *Exercise, gtx layout.Context, theme *material.Theme, window *app.Window, audio_chan chan int) bool {
	var par2 string
	var result string
	var exerciseFinished = false
	if exercise.resultTime > 0 {
		diff := gtx.Now.UnixMilli() - exercise.resultTime
		if diff >= 1000 {
			if exercise.wrongResults+exercise.okResults >= 10 {
				if !exercise.showDialog {
					if exercise.wrongResults == 0 {
						audio_chan <- AUDIO_APPLOAUSE
					}
					exercise.showDialog = true
					exercise.dialogMessage = exercise.getFinalMessage()
				}
			} else {
				exercise.newTask(window, gtx)
			}
		} else {
			inv := op.InvalidateCmd{At: gtx.Now.Add(time.Duration(1000-diff) * time.Millisecond)}
			gtx.Execute(inv)
		}
	}
	if exercise.task.task_type.isParameterQuest() {
		if exercise.userResult > 0 {
			par2 = strconv.Itoa(exercise.userResult)
			if exercise.isPartiallyResult() {
				par2 = par2 + "?"
			}
		} else {
			par2 = "?"
		}
		result = strconv.Itoa(exercise.task.task_result)
	} else {
		par2 = strconv.Itoa(exercise.task.parameter2)
		if exercise.userResult > 0 {
			result = strconv.Itoa(exercise.userResult)
			if exercise.isPartiallyResult() {
				result = result + "?"
			}
		} else {
			result = "?"
		}
	}
	if exercise.showDialog && exercise.dialogClick.Clicked(gtx) {
		exercise.showDialog = false
		exerciseFinished = true
		gtx.Execute(op.InvalidateCmd{})
	}
	isWrong := exercise.userResult > 0 && !exercise.task.checkResult(exercise.userResult) && exercise.isFinalResult()
	if exercise.userResult == 0 || isWrong || exercise.isPartiallyResult() {
		for i := 0; i < 10; i++ {
			if exercise.numberButtons[i].Clicked(gtx) {
				var number = i + 1
				if exercise.task.task_type.getMaxSum() > 10 && number == 10 {
					number = 0
				}
				genNewDelayed := true
				if exercise.wasPartiallyResult() {
					exercise.userResult = exercise.userResult*10 + number
				} else {
					exercise.userResult = number
				}
				if exercise.task.checkResult(exercise.userResult) {
					audio_chan <- randRange(AUDIO_POS1, AUDIO_POS3)
					if exercise.tries == 0 {
						exercise.okResults++
						var curr_score ScoreType
						if !exercise.showAbacus && !exercise.hasAbacusHelp {
							curr_score = score_perfect
						} else {
							curr_score = score_second
						}
						exercise.resultHistory = append(exercise.resultHistory, curr_score)
					}
				} else {
					if exercise.isFinalResult() {
						audio_chan <- randRange(AUDIO_NEG1, AUDIO_NEG3)
						if exercise.tries == 0 {
							exercise.wrongResults++
							exercise.resultHistory = append(exercise.resultHistory, score_wrong)
							genNewDelayed = false
						} else {
							if exercise.tries < 3 {
								genNewDelayed = false
							} else {
								exercise.userResult = exercise.task.userExpectedResult()
							}
						}
						exercise.tries++
					} else {
						genNewDelayed = false
					}
				}
				if genNewDelayed {
					exercise.resultTime = gtx.Now.UnixMilli()
					inv := op.InvalidateCmd{At: gtx.Now.Add(time.Second)}
					gtx.Execute(inv)
				}
			}
		}
		if exercise.eraseButton.Clicked(gtx) {
			exercise.userResult = 0
		}
		if exercise.abacusClick && !exercise.hasAbacusHelp {
			exercise.abacusHelp()
		}
	}
	if exercise.homeButton.Clicked(gtx) {
		exerciseFinished = true
	}
	layout.Inset{
		Top:    unit.Dp(16),
		Bottom: unit.Dp(16),
		Right:  unit.Dp(16),
		Left:   unit.Dp(16),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				par1_widget := material.H4(theme, strconv.Itoa(exercise.task.parameter1))
				op_widget := material.H4(theme, exercise.task.task_type.taskOp())
				par2_widget := material.H4(theme, par2)
				eq_widget := material.H4(theme, "=")
				result_widget := material.H4(theme, result)
				if isWrong {
					redColor := color.NRGBA{R: 0xFF, A: 0xFF}
					if exercise.task.task_type.isParameterQuest() {
						par2_widget.Color = redColor
					} else {
						result_widget.Color = redColor
					}
				}
				return layout.Flex{}.Layout(gtx,
					layout.Rigid(par1_widget.Layout),
					layout.Rigid(op_widget.Layout),
					layout.Rigid(par2_widget.Layout),
					layout.Rigid(eq_widget.Layout),
					layout.Rigid(result_widget.Layout))
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return drawNumberButtons(gtx, theme, exercise)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return drawAbacus(gtx, exercise)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return drawResult(gtx, exercise)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(32)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Max.X = 64
				homebut := material.Button(theme, &exercise.homeButton, "\u2B05")
				homebut.Background = color.NRGBA{R: 0x70, G: 0x70, B: 0x70, A: 0xFF}
				return homebut.Layout(gtx)
			}),
		)
	})
	if exercise.showDialog {
		drawModal(gtx, theme, &exercise.dialogClick, &exercise.showDialog, exercise.dialogMessage)
	}
	return exerciseFinished
}

func drawNumberButtons(gtx layout.Context, theme *material.Theme, exercise *Exercise) layout.Dimensions {
	gap := gtx.Metric.Dp(unit.Dp(16))
	width := gtx.Metric.Dp(unit.Dp(64))
	height := gtx.Metric.Dp(unit.Dp(64))
	size := image.Pt(11*width+9*gap, height*2+gap)
	gtx.Constraints.Min.X = width
	gtx.Constraints.Max.X = width
	gtx.Constraints.Min.Y = height
	gtx.Constraints.Max.Y = height
	for i := 0; i < 5; i++ {
		offset := op.Offset(image.Point{X: i * (width + gap)}).Push(gtx.Ops)
		material.Button(theme, &exercise.numberButtons[i], strconv.Itoa(i+1)).Layout(gtx)
		offset.Pop()
	}
	for i := 0; i < 5; i++ {
		offset := op.Offset(image.Point{X: i * (width + gap), Y: height + gap}).Push(gtx.Ops)
		var dnum int
		if i == 4 && exercise.task.task_type.getMaxSum() > 10 {
			dnum = 0
		} else {
			dnum = i + 6
		}
		material.Button(theme, &exercise.numberButtons[i+5], strconv.Itoa(dnum)).Layout(gtx)
		offset.Pop()
	}
	if exercise.task.task_type.getMaxSum() > 10 {
		offset := op.Offset(image.Point{X: 5 * (width + gap), Y: height + gap}).Push(gtx.Ops)
		ebutton := material.Button(theme, &exercise.eraseButton, "\u232B")
		ebutton.Background = color.NRGBA{R: 0xE0, G: 0xE0, A: 0xFF}
		ebutton.Layout(gtx)
		offset.Pop()
	}
	return layout.Dimensions{Size: size}
}

func (exercise *Exercise) startExercise(taskType TaskType, window *app.Window, gtx layout.Context) {
	exercise.task_type = taskType
	exercise.okResults = 0
	exercise.wrongResults = 0
	exercise.resultHistory = exercise.resultHistory[:0]
	exercise.newTask(window, gtx)
}

func (exercise *Exercise) newTask(window *app.Window, gtx layout.Context) {
	newTaks := genTaks(exercise.task_type)
	var i = 0
	for newTaks == exercise.task && i < 10 {
		newTaks = genTaks(exercise.task_type)
		i++
	}
	exercise.task = newTaks
	exercise.resultTime = 0
	exercise.userResult = 0
	exercise.tries = 0
	exercise.abacus1.setRedBlueRed(0, 0, 0)
	exercise.abacus2.setRedBlueRed(0, 0, 0)
	exercise.abacusClick = false
	exercise.hasAbacusHelp = false
	exercise.showAbacus = false
	exercise.startTaskTime = gtx.Now.Unix()
	exercise.showDialog = false
	exercise.dialogMessage = ""
	go func(currtask Task) {
		time.Sleep(exercise.task.task_type.getTimeDisplayAbacus())
		if currtask == exercise.task && !exercise.showAbacus && !exercise.hasAbacusHelp && !exercise.showDialog {
			exercise.displayAbacus()
			window.Invalidate()
		}
	}(exercise.task)
}

func (exercise *Exercise) getScore() float32 {
	return 100.0 * float32(exercise.okResults) / (float32(exercise.wrongResults + exercise.okResults))
}

func (exercise *Exercise) getFinalMessage() string {
	score := exercise.getScore()
	var message string
	if score >= 100.0 {
		message = "\U0001F600 Perfekt! Kein Fehler"
	} else if score >= 90.0 {
		message = fmt.Sprintf("\U0001F603 Aufgabe beendet richtig: %d falsch: %d", exercise.okResults, exercise.wrongResults)
	} else if score >= 70.0 {
		message = fmt.Sprintf("\U0001F612 Aufgabe beendet richtig: %d falsch: %d", exercise.okResults, exercise.wrongResults)
	} else if score >= 50.0 {
		message = fmt.Sprintf("\U0001F613 Aufgabe beendet richtig: %d falsch: %d", exercise.okResults, exercise.wrongResults)
	} else if score >= 40.0 {
		message = fmt.Sprintf("\U0001F61E Aufgabe beendet richtig: %d falsch: %d", exercise.okResults, exercise.wrongResults)
	} else if score >= 20.0 {
		message = fmt.Sprintf("\U0001F622 Aufgabe beendet richtig: %d falsch: %d", exercise.okResults, exercise.wrongResults)
	} else {
		message = "\U0001F435 War da ein Affe dran?"
	}
	return message
}

func (exercise *Exercise) wasPartiallyResult() bool {
	return exercise.task.task_type.getMaxSum() > 10 &&
		exercise.userResult <= exercise.task.task_type.getMaxSum()/10 &&
		exercise.userResult > 0
}

func (exercise *Exercise) isFinalResult() bool {
	return exercise.task.task_type.getMaxSum() <= 10 ||
		exercise.userResult > exercise.task.task_type.getMaxSum()/10 ||
		(exercise.userResult <= exercise.task.task_type.getMaxSum()/10 && exercise.task.checkResult(exercise.userResult))
}

func (exercise *Exercise) isPartiallyResult() bool {
	return exercise.userResult > 0 &&
		exercise.task.task_type.getMaxSum() > 10 &&
		exercise.userResult <= exercise.task.task_type.getMaxSum()/10 &&
		!exercise.task.checkResult(exercise.userResult)
}

func drawModal(gtx layout.Context, th *material.Theme, okButton *widget.Clickable, showModal *bool, message string) {
	paint.FillShape(gtx.Ops, color.NRGBA{A: 128}, clip.Rect{Max: gtx.Constraints.Max}.Op())

	layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Stack{
			Alignment: layout.Center,
		}.Layout(gtx,
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				return drawRoundedBox(gtx, color.NRGBA{R: 255, G: 255, B: 255, A: 255}, 300, 150)
			}),
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Vertical,
					Alignment: layout.Middle,
					Spacing:   layout.SpaceAround,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Body1(th, message).Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if okButton.Clicked(gtx) {
							*showModal = false // Close the modal
						}
						return material.Button(th, okButton, "OK").Layout(gtx)
					}),
				)
			}),
		)
	})
}

// drawRoundedBox draws a box with rounded corners
func drawRoundedBox(gtx layout.Context, col color.NRGBA, width, height int) layout.Dimensions {
	defer clip.Rect{
		Min: image.Pt(0, 0),
		Max: image.Pt(width, height),
	}.Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, col)
	return layout.Dimensions{
		Size: image.Pt(width, height),
	}
}
