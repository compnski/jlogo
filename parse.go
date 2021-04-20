package main

import (
	"io"

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
		{"EOL", `[\n\r\d]+`, nil},
		{"whitespace", `[ \t]+`, nil},
	})

	basicParser = participle.MustBuild(&Program{},
		participle.Lexer(basicLexer),
		participle.CaseInsensitive("Ident"),
		participle.Unquote("String"),
		participle.UseLookahead(2),
	)
)

func Parse(r io.Reader) (*Program, error) {
	program := &Program{}
	err := basicParser.Parse("", r, program)
	if err != nil {
		return nil, err
	}

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

type Sleep struct {
	Pos lexer.Position

	Expression Expression `("SLEEP" | "SP") @@`
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
	Commands []Command   `"[" @@+ "]"`
}

type Comment struct {
	Pos      lexer.Position
	Commands []Command `"#" @@*`
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
	Sleep    *Sleep    ` @@ |`
	Comment  *Comment  ` @@ |`
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

/* FORWARD FD */
/* BACK BK */
/* RIGHT RT */
/* LEFT LT */
