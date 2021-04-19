package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/chzyer/readline"
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
	var fileName string
	flag.BoolVar(&usePiTurtle, "pi", false, "Use the pi turtle")
	flag.StringVar(&fileName, "file", "", "Run this program")
	flag.Parse()
	log.Print("Welcome to jlogo!")
	var turtle Turtle
	if usePiTurtle {
		log.Print("Using pi turtle!")
		turtle = InitPiTurtle()
		defer turtle.Close()
	} else {
		log.Print("Using text turtle!")
		turtle = NewTextTurtle(os.Stdout)
	}

	if fileName != "" {
		runProgramFromFile(fileName, turtle)
	} else {
		runProgramFromStdin(turtle)
	}
}

func runProgramFromStdin(turtle Turtle) {
	rl, err := readline.New("> ")
	if err != nil {
		panic(err)
	}
	defer rl.Close()
	program := &Program{}

	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}
		line += "\n"
		err = basicParser.ParseString("", line, program) //program, err := Parse(strings.NewReader(line))
		if err != nil {
			log.Fatalf("Error parsing line [%v], got %v", line, err)
		}
		funcs := map[string]Function{}
		err = program.Evaluate(turtle, os.Stdin, os.Stdout, funcs)
		if err != nil {
			log.Fatalf("Error running program, got %v", err)
		}
	}
}

func runProgramFromFile(fileName string, turtle Turtle) {
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

	funcs := map[string]Function{}
	err = program.Evaluate(turtle, os.Stdin, os.Stdout, funcs)
	if err != nil {
		log.Fatalf("Error running program, got %v", err)
	}
}
