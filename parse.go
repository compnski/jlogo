package main

import (
	"io"
	"log"
	"os"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/alecthomas/participle/v2/lexer/stateful"
)

/* Commands:*/
/* SETHEADING SETH */
/* HOME*/
/* PENDOWNP PENDOWN?*/
/* CLEAN*/
/* CLEARSCREEN CS (HOME+CLEAN)*/
/* REPEAT*/
/* REPCOUNT*/
/* IF*/
/* IFELSE*/
/* STOP*/
/* OUTPUT*/

/* Maybes:*/
/* SETPOS*/
/* SETXY */
/* SETX */
/* SETY */
/* TEST*/
/* IFTRUE*/
/* IFFALSE*/
var (
	basicLexer = stateful.MustSimple([]stateful.Rule{
		{"Comment", `(?i)rem[^\n]*`, nil},
		{"String", `"(\\"|[^"])*"`, nil},
		{"Number", `[-+]?(\d*\.)?\d+`, nil},
		{"Ident", `[a-zA-Z_]\w*`, nil},
		{"Punct", `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`, nil},
		{"EOL", `[\n\r]+`, nil},
		{"whitespace", `[ \t]+`, nil},
	})

	basicParser = participle.MustBuild(&Program{},
		participle.Lexer(basicLexer),
		participle.CaseInsensitive("Ident"),
		participle.Unquote("String"),
		participle.UseLookahead(2),
	)
)

func main() {
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

	funcs := map[string]Function{}
	err = program.Evaluate(os.Stdin, os.Stdout, funcs)
	if err != nil {
		log.Fatalf("Error running program, got %v", err)
	}

}

func Parse(r io.Reader) (*Program, error) {
	program := &Program{}
	err := basicParser.Parse("", r, program)
	if err != nil {
		return nil, err
	}
	program.init()
	return program, nil
}

//////////
////////// Commands
//////////
type Forward struct {
	Pos lexer.Position

	Expression Expression `("FORWARD" | "FD") @@`
}

type Backward struct {
	Pos lexer.Position

	Expression Expression `("BACKWARD" | "BK") @@`
}

type Right struct {
	Pos lexer.Position

	Expression Expression `("RIGHT" | "RT") @@`
}

type Left struct {
	Pos lexer.Position

	Expression Expression `("LEFT" | "LT") @@`
}

type PenUp struct {
	Pos   lexer.Position
	Ident bool `("PENUP" | "PU")`
}

type PenDown struct {
	Pos   lexer.Position
	Ident bool `("PENDOWN" | "PD")`
}

type Repeat struct {
	Pos      lexer.Position
	Times    *Expression `"REPEAT" @@`
	Commands []Command   `@@+`
}

type Stop struct {
	Pos   lexer.Position
	Ident bool `"STOP"`
}

/* SETHEADING SETH */
/* HOME*/
/* PENDOWNP PENDOWN?*/
/* CLEAN*/
/* CLEARSCREEN CS (HOME+CLEAN)*/
/* REPEAT*/
/* REPCOUNT*/
/* IF*/
/* IFELSE*/
/* STOP*/
/* OUTPUT*/

type Command struct {
	Pos lexer.Position

	Index int

	//	Line int `@Number`

	Forward  *Forward  `( @@ |`
	Backward *Backward ` @@ |`
	Right    *Right    ` @@ |`
	Left     *Left     ` @@ |`
	PenUp    *PenUp    ` @@ |`
	PenDown  *PenDown  ` @@ |`
	Repeat   *Repeat   ` @@ |`
	Stop     *Stop     ` @@)`

	// 	Remark *Remark `(   @@`
	// 	Input  *Input  `  | @@`
	// 	Let    *Let    `  | @@`
	// 	Goto   *Goto   `  | @@`
	// 	If     *If     `  | @@`
	// 	Print  *Print  `  | @@`
	// 	Call   *Call   `  | @@ ) EOL`
}

///////////////////////////
/////////////////////////// Program Structure
///////////////////////////
type Line struct {
	Command *Command `@@ EOL`
}

type Program struct {
	Pos lexer.Position

	Commands []Command `(@@ EOL)*`
}

func (p *Program) init() {
	// 	p.Table = map[int]*Command{}
	// 	for index, cmd := range p.Commands {
	// 		cmd.Index = index
	// 		p.Table[cmd.Line] = cmd
	// 	}
}

type Operator string

func (o *Operator) Capture(s []string) error {
	*o = Operator(strings.Join(s, ""))
	return nil
}

type Value struct {
	Pos lexer.Position

	Number   *float64 `  @Number`
	Variable *string  `| @Ident`
	String   *string  `| @String`
	//Call          *Call       `| @@`
	Subexpression *Expression `| "(" @@ ")"`
}

type Factor struct {
	Pos lexer.Position

	Base     *Value `@@`
	Exponent *Value `( "^" @@ )?`
}

type OpFactor struct {
	Pos lexer.Position

	Operator Operator `@("*" | "/")`
	Factor   *Factor  `@@`
}

type Term struct {
	Pos lexer.Position

	Left  *Factor     `@@`
	Right []*OpFactor `@@*`
}

type OpTerm struct {
	Pos lexer.Position

	Operator Operator `@("+" | "-")`
	Term     *Term    `@@`
}

type Cmp struct {
	Pos lexer.Position

	Left  *Term     `@@`
	Right []*OpTerm `@@*`
}

type OpCmp struct {
	Pos lexer.Position

	Operator Operator `@("=" | "<" "=" | ">" "=" | "<" | ">" | "!" "=")`
	Cmp      *Cmp     `@@`
}

type Expression struct {
	Pos lexer.Position

	Left  *Cmp     `@@`
	Right []*OpCmp `@@*`
}

/* FORWARD FD */
/* BACK BK */
/* RIGHT RT */
/* LEFT LT */
