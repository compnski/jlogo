package main

import (
	"fmt"
	"io"
	"math"

	"github.com/alecthomas/repr"

	"github.com/alecthomas/participle/v2"
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
			ctx.Turtle.Rotate(value.(float64))
		case cmd.Left != nil:
			cmd := cmd.Left
			value, err := cmd.Expression.Evaluate(ctx)
			if err != nil {
				return err
			}
			ctx.Turtle.Rotate(-value.(float64))
		case cmd.PenUp != nil:
			ctx.Turtle.PenUp(true)
		case cmd.PenDown != nil:
			ctx.Turtle.PenUp(false)
		case cmd.Repeat != nil:
			cmd := cmd.Repeat
			value, err := cmd.Times.Evaluate(ctx)
			if err != nil {
				return err
			}
			for i := 0.0; i < value.(float64); i++ {
				RunCommandList(cmd.Commands, ctx)
			}
			//fmt.Fprintf(ctx.Output, "repeat %v\n", value)
		default:
			panic("unsupported command " + repr.String(cmd))
		}

		index++
	}
	return nil
}

func (v *Value) Evaluate(ctx *Context) (interface{}, error) {
	switch {
	case v.Number != nil:
		return *v.Number, nil
	case v.String != nil:
		return *v.String, nil
	case v.Variable != nil:
		value, ok := ctx.Vars[*v.Variable]
		if !ok {
			return nil, fmt.Errorf("unknown variable %q", *v.Variable)
		}
		return value, nil
	case v.Subexpression != nil:
		return v.Subexpression.Evaluate(ctx)
		// case v.Call != nil:
		// 	return v.Call.Evaluate(ctx)
	}
	panic("unsupported value type" + repr.String(v))
}

func (f *Factor) Evaluate(ctx *Context) (interface{}, error) {
	base, err := f.Base.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	if f.Exponent == nil {
		return base, nil
	}
	baseNum, exponentNum, err := evaluateFloats(ctx, base, f.Exponent)
	if err != nil {
		return nil, participle.Errorf(f.Pos, "invalid factor: %s", err)
	}
	return math.Pow(baseNum, exponentNum), nil
}

func (o *OpFactor) Evaluate(ctx *Context, lhs interface{}) (interface{}, error) {
	lhsNumber, rhsNumber, err := evaluateFloats(ctx, lhs, o.Factor)
	if err != nil {
		return nil, participle.Errorf(o.Pos, "invalid arguments for %s: %s", o.Operator, err)
	}
	switch o.Operator {
	case "*":
		return lhsNumber * rhsNumber, nil
	case "/":
		return lhsNumber / rhsNumber, nil
	}
	panic("unreachable")
}

func (t *Term) Evaluate(ctx *Context) (interface{}, error) {
	lhs, err := t.Left.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	for _, right := range t.Right {
		rhs, err := right.Evaluate(ctx, lhs)
		if err != nil {
			return nil, err
		}
		lhs = rhs
	}
	return lhs, nil
}

func (o *OpTerm) Evaluate(ctx *Context, lhs interface{}) (interface{}, error) {
	lhsNumber, rhsNumber, err := evaluateFloats(ctx, lhs, o.Term)
	if err != nil {
		return nil, participle.Errorf(o.Pos, "invalid arguments for %s: %s", o.Operator, err)
	}
	switch o.Operator {
	case "+":
		return lhsNumber + rhsNumber, nil
	case "-":
		return lhsNumber - rhsNumber, nil
	}
	panic("unreachable")
}

func (c *Cmp) Evaluate(ctx *Context) (interface{}, error) {
	lhs, err := c.Left.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	for _, right := range c.Right {
		rhs, err := right.Evaluate(ctx, lhs)
		if err != nil {
			return nil, err
		}
		lhs = rhs
	}
	return lhs, nil
}

func (o *OpCmp) Evaluate(ctx *Context, lhs interface{}) (interface{}, error) {
	rhs, err := o.Cmp.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	switch lhs := lhs.(type) {
	case float64:
		rhs, ok := rhs.(float64)
		if !ok {
			return nil, participle.Errorf(o.Pos, "rhs of %s must be a number", o.Operator)
		}
		switch o.Operator {
		case "=":
			return lhs == rhs, nil
		case "!=":
			return lhs != rhs, nil
		case "<":
			return lhs < rhs, nil
		case ">":
			return lhs > rhs, nil
		case "<=":
			return lhs <= rhs, nil
		case ">=":
			return lhs >= rhs, nil
		}
	case string:
		rhs, ok := rhs.(string)
		if !ok {
			return nil, participle.Errorf(o.Pos, "rhs of %s must be a string", o.Operator)
		}
		switch o.Operator {
		case "=":
			return lhs == rhs, nil
		case "!=":
			return lhs != rhs, nil
		case "<":
			return lhs < rhs, nil
		case ">":
			return lhs > rhs, nil
		case "<=":
			return lhs <= rhs, nil
		case ">=":
			return lhs >= rhs, nil
		}
	default:
		return nil, participle.Errorf(o.Pos, "lhs of %s must be a number or string", o.Operator)
	}
	panic("unreachable")
}

func (e *Expression) Evaluate(ctx *Context) (interface{}, error) {
	lhs, err := e.Left.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	for _, right := range e.Right {
		rhs, err := right.Evaluate(ctx, lhs)
		if err != nil {
			return nil, err
		}
		lhs = rhs
	}
	return lhs, nil
}

// func (c *Call) Evaluate(ctx *Context) (interface{}, error) {
// 	function, ok := ctx.Functions[c.Name]
// 	if !ok {
// 		return nil, participle.Errorf(c.Pos, "unknown function %q", c.Name)
// 	}
// 	args := []interface{}{}
// 	for _, arg := range c.Args {
// 		value, err := arg.Evaluate(ctx)
// 		if err != nil {
// 			return nil, err
// 		}
// 		args = append(args, value)
// 	}

// 	value, err := function(args...)
// 	if err != nil {
// 		return nil, participle.Errorf(c.Pos, "call to %s() failed", c.Name)
// 	}
// 	return value, nil
// }

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

func evaluateFloats(ctx *Context, lhs interface{}, rhsExpr Evaluatable) (float64, float64, error) {
	rhs, err := rhsExpr.Evaluate(ctx)
	if err != nil {
		return 0, 0, err
	}
	lhsNumber, ok := lhs.(float64)
	if !ok {
		return 0, 0, fmt.Errorf("lhs must be a number")
	}
	rhsNumber, ok := rhs.(float64)
	if !ok {
		return 0, 0, fmt.Errorf("rhs must be a number")
	}
	return lhsNumber, rhsNumber, nil
}
