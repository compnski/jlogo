package main

import (
	"flag"
	"log"
	"os"
	"time"
)

const (
	PinPenServo = 18
)

var (
	PinsLeftWheel  = []int{6, 13, 19, 26}
	PinsRightWheel = []int{12, 16, 20, 21}
)

func NewWheel(pins []int) (*GPIOStepper, error) {
	gpioPins, err := InitGPIOPins(pins)
	if err != nil {
		return nil, err
	}
	wheel, err := NewGPIOStepper(
		time.Millisecond,
		gpioPins,
		StandardStepperPattern,
	)
	return wheel, err
}

func main() {
	var usePiTurtle bool
	flag.BoolVar(&usePiTurtle, "pi", false, "Use the pi turtle")
	flag.Parse()

	fileName := "test.logo"
	r, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Error reading file %s, got %v", fileName, err)
	}
	defer r.Close()
	program, err := Parse(r)
	if err != nil {
		log.Fatalf("Error parsing program, got %v", err)
	}
	log.Printf("%+v", program)

	var turtle Turtle
	if usePiTurtle {
		log.Print("Using pi turtle!")
		turtle = InitPiTurtle()
	} else {
		turtle = NewTextTurtle(os.Stdout)
	}

	funcs := map[string]Function{}
	err = program.Evaluate(turtle, os.Stdin, os.Stdout, funcs)
	if err != nil {
		log.Fatalf("Error running program, got %v", err)
	}

}

func InitPiTurtle() *PiTurtle {
	pwm, err := NewPiBlaster(PinPenServo)
	if err != nil {
		log.Fatal(err)
	}
	servo, err := NewPWMServo(pwm, 0, 180, 0.05, 0.2)
	if err != nil {
		log.Fatal(err)
	}
	pen := ServoPen{
		Servo: servo,
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
