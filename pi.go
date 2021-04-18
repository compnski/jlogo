package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type PWM interface {
	DutyCycle(dc float64) error
	Release() error
}

type PWMServo struct {
	PWM                        PWM
	MinAngle, MaxAngle         float64
	MinDutyCycle, MaxDutyCycle float64
	angleFactor                float64
}

var ErrRange = errors.New("value out of range")
var ErrAngle = errors.New("Max angle not greater than min")
var ErrDutyCycle = errors.New("Max duty cycle not greater than min")

func NewPWMServo(p PWM, minAngle, maxAngle, minDutyCycle, maxDutyCycle float64) (*PWMServo, error) {
	if maxAngle <= minAngle {
		return nil, ErrAngle
	}
	if maxDutyCycle <= minDutyCycle {
		return nil, ErrAngle
	}
	return &PWMServo{
		PWM:          p,
		MinAngle:     minAngle,
		MaxAngle:     maxAngle,
		MinDutyCycle: minDutyCycle,
		MaxDutyCycle: maxDutyCycle,
		angleFactor:  (maxAngle - minAngle) / (maxDutyCycle - minDutyCycle),
	}, nil
}

func (s *PWMServo) Angle(deg float64) error {
	if deg < s.MinAngle || deg > s.MaxAngle {
		return ErrRange
	}
	return s.PWM.DutyCycle(s.MinDutyCycle + deg*s.angleFactor)
}

type PiBlaster struct {
	Pin    int
	Handle *os.File
}

func NewPiBlaster(pin int) (*PiBlaster, error) {
	h, err := os.OpenFile("/dev/pi-blaster", os.O_RDWR, 0666)
	return &PiBlaster{Pin: pin, Handle: h}, err
}

func (p *PiBlaster) DutyCycle(dc float64) error {
	_, err := fmt.Fprintf(p.Handle, fmt.Sprintf("%d=%f", p.Pin, dc))
	return err
}

func (p *PiBlaster) Release() error {
	_, err := fmt.Fprintf(p.Handle, fmt.Sprintf("release %d", p.Pin))
	return err

}

type PiGPIO struct {
	Pin    int
	Handle *os.File
}

var ErrUnitialized = errors.New("Pin not initialized")
var ErrNoGPIO = errors.New("gpio not available. Please install wiringpi")

func InitGPIOPins(pins []int) (gpio []GPIO, err error) {
	for _, pin := range pins {
		var p GPIO
		p, err = NewGPIO(pin)
		if err != nil {
			return
		}
		gpio = append(gpio, p)
	}
	return
}
func NewGPIO(pin int) (*PiGPIO, error) {
	cmd := exec.Command("gpio", "export", strconv.Itoa(pin), "out")
	err := cmd.Run()
	if err != nil {
		//out, _ := cmd.CombinedOutput()
		log.Printf("Failed to run gpio: %v\n%v", err, cmd)
		return nil, ErrNoGPIO
	}
	h, err := os.OpenFile(fmt.Sprintf("/sys/class/gpio/gpio%d/value", pin), os.O_RDWR, 770)
	if err != nil {
		log.Printf("Failed to open gpio pin: %v", err)
		return nil, ErrUnitialized
	}

	return &PiGPIO{
		Pin:    pin,
		Handle: h,
	}, nil
}

var enableMap = map[bool]string{
	true:  "1",
	false: "0",
}

func (g *PiGPIO) Enable(b bool) error {
	if g.Handle == nil {
		return ErrUnitialized
	}
	_, err := fmt.Fprintf(g.Handle, enableMap[b])
	return err
}

type GPIO interface {
	Enable(bool) error
}

type GPIOStepper struct {
	Pins        []GPIO
	Delay       time.Duration
	Pattern     [][]bool
	currentStep int
	Sleep       func(time.Duration)
}

var ErrNoPins = errors.New("No pins")
var ErrBadPattern = errors.New("Pattern doesn't match pins")

var StandardStepperPattern = [][]bool{
	{false, false, false, true},
	{false, false, true, true},
	{false, false, true, false},
	{false, true, true, false},
	{false, true, false, false},
	{true, true, false, false},
	{true, false, false, false},
	{true, false, false, true},
	{false, false, false, false},
}

func NewGPIOStepper(delay time.Duration, pins []GPIO, pattern [][]bool) (*GPIOStepper, error) {
	if len(pins) == 0 {
		return nil, ErrNoPins
	}
	if len(pattern) == 0 {
		return nil, ErrBadPattern
	}
	for _, p := range pattern {
		if len(p) != len(pins) {
			return nil, ErrBadPattern
		}
	}
	return &GPIOStepper{
		Pins:    pins,
		Delay:   delay,
		Pattern: pattern,
		Sleep:   time.Sleep,
	}, nil
}

func (s *GPIOStepper) Forward() error {
	return s.Step(1)
}
func (s *GPIOStepper) Backward() error {
	return s.Step(-1)
}

func (s *GPIOStepper) StepOne(dir int) error {
	s.currentStep = (s.currentStep + dir) % len(s.Pattern)
	for pin := range s.Pins {
		if err := s.Pins[pin].Enable(s.Pattern[s.currentStep][pin]); err != nil {
			return err
		}
	}
	return nil
}

func (s *GPIOStepper) Step(n int) error {
	var dir = 1
	if n < 0 {
		dir = -1
		n = -n
	}
	for step := 0; step < n; step++ {
		if err := s.StepOne(dir); err != nil {
			return err
		}
		s.Sleep(s.Delay)
	}
	return nil
}
