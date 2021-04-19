package main

import (
	"io"
	"time"

	"github.com/alecthomas/repr"
)

type Evaluatable interface {
	Evaluate(ctx *Context) (interface{}, error)
}

type Function func(args ...interface{}) (interface{}, error)

type TurtleController interface {
	// Move steps. If steps is negative, move backward
	// Should return the current position on completion.
	Move(steps float64) (x, y float64, err error)
	// Rotate clockwise. If deg is negative, rotate counter-clockwise.
	// Should return the current heading on completion.
	Rotate(deg float64) (heading float64, err error)
	// PenUp sets the state of the pen to state. PenUp(true) to stop drawing, PenUp(false) to start again
	PenUp(state bool) (bool, error)
}

// Context for evaluation.
type Context struct {
	// User-provided functions.
	Functions map[string]Function
	// Vars defined during evaluation.
	Vars map[string]interface{}
	// Turtle for drawing
	Turtle TurtleController
	// Reader from which INPUT is read.
	Input io.Reader
	// Writer where PRINTing will write.
	Output io.Writer
}

func RunCommandList(commands []Command, ctx *Context) error {
	for index := 0; index < len(commands); {
		cmd := commands[index]
		//fmt.Fprintf(ctx.Output, "Got Cmd: %+v\n", cmd.Command)
		switch {
		case cmd.Sleep != nil:
			cmd := cmd.Sleep
			value, err := cmd.Expression.Evaluate(ctx)
			if err != nil {
				return err
			}
			time.Sleep(time.Millisecond * time.Duration(value.(float64)))
		case cmd.Forward != nil:
			cmd := cmd.Forward
			value, err := cmd.Expression.Evaluate(ctx)
			if err != nil {
				return err
			}
			ctx.Turtle.Move(value.(float64))
		case cmd.Backward != nil:
			cmd := cmd.Backward
			value, err := cmd.Expression.Evaluate(ctx)
			if err != nil {
				return err
			}
			ctx.Turtle.Move(-value.(float64))
			//ctx.Vars[cmd.Variable] = value
		case cmd.Right != nil:
			cmd := cmd.Right
			value, err := cmd.Expression.Evaluate(ctx)
			if err != nil {
				return err
			}
			ctx.Turtle.Rotate(-value.(float64))
		case cmd.Left != nil:
			cmd := cmd.Left
			value, err := cmd.Expression.Evaluate(ctx)
			if err != nil {
				return err
			}
			ctx.Turtle.Rotate(value.(float64))
		case cmd.PenUp != nil:
			_, err := ctx.Turtle.PenUp(true)
			if err != nil {
				return err
			}
		case cmd.PenDown != nil:
			_, err := ctx.Turtle.PenUp(false)
			if err != nil {
				return err
			}
		case cmd.Repeat != nil:
			cmd := cmd.Repeat
			value, err := cmd.Times.Evaluate(ctx)
			if err != nil {
				return err
			}
			for i := 0.0; i < value.(float64); i++ {
				err := RunCommandList(cmd.Commands, ctx)
				if err != nil {
					return err
				}
			}
		//fmt.Fprintf(ctx.Output, "repeat %v\n",
		case cmd.Comment != nil:
		default:
			panic("unsupported command " + repr.String(cmd))
		}

		index++
	}
	return nil
}

func (p *Program) Evaluate(turtle Turtle, r io.Reader, w io.Writer, functions map[string]Function) error {
	if len(p.Commands) == 0 {
		return nil
	}

	ctx := &Context{
		Vars:      map[string]interface{}{},
		Functions: functions,
		Input:     r,
		Output:    w,
		Turtle:    turtle,
	}
	return RunCommandList(p.Commands, ctx)
}
