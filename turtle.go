package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"time"
)

type BaseTurtle struct {
	X, Y    float64
	Heading float64
	IsPenUp bool
}

var penStateMap = map[bool]string{
	false: "down",
	true:  "up",
}

type Turtle interface {
	Move(steps float64) (x, y float64, err error)
	Rotate(deg float64) (heading float64, err error)
	PenUp(state bool) (bool, error)
	State() BaseTurtle
	Close()
}

func NewTextTurtle(w io.Writer) *TextTurtle {
	return &TextTurtle{
		Turtle: &BaseTurtle{},
		Output: w,
	}
}

type TextTurtle struct {
	Turtle
	Output io.Writer
}

func deg2rad(deg float64) float64 {
	return deg * math.Pi / 180
}

func (t BaseTurtle) Close() {

}

func (t BaseTurtle) State() BaseTurtle {
	return t
}

// Move steps. If steps is negative, move backward
// Should return the current position on completion.
func (t *BaseTurtle) Move(steps float64) (x, y float64, err error) {
	t.X = steps * math.Cos(deg2rad(t.Heading))
	t.Y = steps * math.Sin(deg2rad(t.Heading))
	return t.X, t.Y, nil
}

// Rotate clockwise. If deg is negative, rotate counter-clockwise.
// Should return the current heading on completion.
func (t *BaseTurtle) Rotate(deg float64) (heading float64, err error) {
	t.Heading = math.Mod(t.Heading+deg, 360)
	return t.Heading, nil
}

// PenUp sets the state of the pen to state. PenUp(true) to stop drawing, PenUp(false) to start again
func (t *BaseTurtle) PenUp(state bool) (bool, error) {
	t.IsPenUp = state
	return t.IsPenUp, nil
}

// Move steps. If steps is negative, move backward
// Should return the current position on completion.
func (t *TextTurtle) Move(steps float64) (x, y float64, err error) {
	defer func() { fmt.Fprintf(t.Output, "Moved %v steps. Now at (%v, %v)\n", steps, t.State().X, t.State().Y) }()
	return t.Turtle.Move(steps)
}

// Rotate clockwise. If deg is negative, rotate counter-clockwise.
// Should return the current heading on completion.
func (t *TextTurtle) Rotate(deg float64) (heading float64, err error) {
	defer func() { fmt.Fprintf(t.Output, "Rotated %v degrees. Now facing %v\n", deg, t.State().Heading) }()
	return t.Turtle.Rotate(deg)
}

// PenUp sets the state of the pen to state. PenUp(true) to stop drawing, PenUp(false) to start again
func (t *TextTurtle) PenUp(state bool) (bool, error) {
	defer func() { fmt.Fprintf(t.Output, "Pen is now %v\n", penStateMap[t.State().IsPenUp]) }()
	return t.Turtle.PenUp(state)
}

type Stepper interface {
	Step(n int) error
	StepOne(dir int) error
}

type Servo interface {
	Angle(deg float64) error
}

type ServoPen struct {
	Servo              Servo
	UpAngle, DownAngle float64
}

func (p *ServoPen) Up() error {
	return p.Servo.Angle(p.UpAngle)
}

func (p *ServoPen) Down() error {
	return p.Servo.Angle(p.DownAngle)
}

type PiTurtle struct {
	Turtle
	Pen                   ServoPen
	LeftWheel, RightWheel Stepper
	Sleep                 func(time.Duration)
	Delay                 time.Duration
}

const (
	PinPenServo = 18
)

var (
	PinsLeftWheel  = []int{6, 13, 19, 26}
	PinsRightWheel = []int{21, 20, 16, 12}
)

func NewPiTurtle(w io.Writer, pen ServoPen, leftWheel, rightWheel Stepper) *PiTurtle {
	return &PiTurtle{
		Turtle:     NewTextTurtle(w),
		Pen:        pen,
		LeftWheel:  leftWheel,
		RightWheel: rightWheel,
		Sleep:      time.Sleep,
		Delay:      time.Millisecond * 2,
	}
}

func InitPiTurtle() *PiTurtle {
	pwm, err := NewPiBlaster(PinPenServo)
	if err != nil {
		log.Fatal(err)
	}
	servo, err := NewPWMServo(pwm, 0, 90, 0.05, 0.2)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Servo: %+v", servo)
	pen := ServoPen{
		Servo:     servo,
		UpAngle:   0,
		DownAngle: 90,
	}

	leftWheel, err := NewWheel(PinsLeftWheel)
	if err != nil {
		log.Fatal(err)
	}
	rightWheel, err := NewWheel(PinsRightWheel)
	if err != nil {
		log.Fatal(err)
	}

	return NewPiTurtle(os.Stdout,
		pen,
		leftWheel,
		rightWheel)
}

func (t *PiTurtle) Close() {
	t.PenUp(true)
	// 	t.LeftWheel.Close()
	// 	t.RightWheel.Close()
	// 	t.Pen.Close()
}

// Move steps. If steps is negative, move backward
// Should return the current position on completion.
func (t *PiTurtle) Move(steps float64) (x, y float64, err error) {
	//steps *= -1
	var dir = 1
	if steps < 0 {
		dir = -1
		steps = -steps
	}
	stepperSteps := steps * 100
	// TODO Figure out mapping of steps to stepper steps
	// TODO Figure float -> int issues here
	for step := 0.0; step < stepperSteps; step++ {
		err = t.LeftWheel.StepOne(dir)
		if err == nil {
			err = t.RightWheel.StepOne(dir)
		}
		if err != nil {
			break
		}
		t.Sleep(t.Delay)
	}
	if err != nil {
		return t.Turtle.State().X, t.Turtle.State().Y, err
	}
	return t.Turtle.Move(steps)
}

// Rotate clockwise. If deg is negative, rotate counter-clockwise.
// Should return the current heading on completion.
func (t *PiTurtle) Rotate(deg float64) (heading float64, err error) {
	var dir = 1
	if deg < 0 {
		dir = -1
		deg = -deg
	}
	stepperSteps := deg * 23
	// TODO Figure out mapping of deg to stepper steps
	// TODO Figure float -> int issues here
	for step := 0.0; step < stepperSteps; step++ {
		err = t.LeftWheel.StepOne(dir)
		if err == nil {
			err = t.RightWheel.StepOne(-dir)
		}
		if err != nil {
			break
		}
		t.Sleep(t.Delay)
	}
	if err != nil {
		return t.Turtle.State().Heading, err
	}
	return t.Turtle.Rotate(deg)
}

// PenUp sets the state of the pen to state. PenUp(true) to stop drawing, PenUp(false) to start again
func (t *PiTurtle) PenUp(state bool) (bool, error) {
	var err error
	if state {
		err = t.Pen.Up()
	} else {
		err = t.Pen.Down()
	}
	if err != nil {
		return t.Turtle.State().IsPenUp, err
	}
	t.Sleep(t.Delay)
	return t.Turtle.PenUp(state)
}
