package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

//------------------------------------------------------------
// Escape sequences for colors
var colorRed = "\033[31m"
var colorGreen = "\033[32m"
var colorYellow = "\033[33m"
var colorBlue = "\033[34m"
var colorPurple = "\033[35m"
var colorCyan = "\033[36m"
var colorWhite = "\033[37m"

var colorReset = "\033[0m"

/*------------------------------------------------------------*/

// Stack is a Go stack implementation
type Stack struct {
	Value [100000]interface{}
	index int
}

// Push an item onto the stack
func (s *Stack) Push(x interface{}) {
	s.Value[s.index] = x
	s.index++
	if s.index > len(ValueStack.Value)-10 {
		GfError("Call stack overflow at %d entries; resetting.", s.index)
		s.Reset()
	}
}

// Tos is the top of stack item
func (s *Stack) Tos() interface{} {
	if s.index == 0 {
		GfError("Stack is empty!")
		return nil
	}
	return s.Value[s.index-1]
}

// Pop an item off the stack
func (s *Stack) Pop(id string) interface{} {
	var r interface{}
	if s.index == 0 {
		if len(id) > 0 {
			GfError("Error popping value '%s': stack is empty!", id)
		} else {
			GfError("Stack is empty!")
		}
		r = nil
	} else {
		s.index--
		r = s.Value[s.index]
	}
	return r
}

// Print the contents of the stack non-destructively
func (s *Stack) Print() {
	if s.index > -1 {
		fmt.Println("Stack:")
		i := s.index

		count := 0
		for i > 0 && count < 6 {
			i--
			count++
			fmt.Printf(colorYellow+"%d: %v\n", s.index-i-1, s.Value[i])
		}
	}
}

// Reset the stack state
func (s *Stack) Reset() {
	s.index = 0
	for i := 0; i < len(s.Value); i++ {
		s.Value[0] = nil
	}
}

/*------------------------------------------------------------*/

// ReadLn reads a line from standard input
func ReadLn() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	line = string(strings.TrimRight(line, "\r\n"))
	return line, err
}

func ReadChar() (byte, error) {
	var b []byte = make([]byte, 1)
	cnt, err := os.Stdin.Read(b)
	if err != nil {
		return 0, err
	}
	if cnt != 1 {
		return 0, nil
	}
	return b[0], nil
}

/*------------------------------------------------------------*/
/*
	Intermediate struct used to sort maps - the map gets flattened
	into a list of key/value pairs then that list is sorted.
*/
type Pair struct {
	Key   interface{}
	Value interface{}
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Less(i, j int) bool { return Compare(p[i].Value, p[j].Value) == -1 }

type DescendingPairList []Pair

func (p DescendingPairList) Len() int           { return len(p) }
func (p DescendingPairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p DescendingPairList) Less(i, j int) bool { return Compare(p[i].Value, p[j].Value) == 1 }

/*------------------------------------------------------------*/

// Token represents a token in the language
type Token struct {
	File   string
	Name   string
	Text   string
	Line   int
	Offset int
}

func (tok Token) GetCodeLine() (string, int) {
	if tok.Text != "" {
		text := tok.Text
		if tok.Offset > 0 && tok.Offset < len(text) {
			offset := tok.Offset
			start := offset
			end := offset
			for start > 0 && text[start] != '\n' {
				start--
			}
			if text[start] == '\n' {
				start++
			}
			for end < len(text) && text[end] != '\n' {
				end++
			}
			return string(text[start:end]), offset - start
		}
	}
	return "", 0
}

var lineno int = 1

var currentFile = "<stdin>"

//
// ParseLine parses a string into tokens which will then be compiled into a GoForth lambda
//
func ParseLine(text string) []Token {
	strtemp := ""
	var result []Token
	inString := false
	inRegex := false
	inComment := false
	inChar := false
	braceCount := 0
	squareCount := 0
	quoted := false

	text = strings.TrimSpace(text)

	for offset, chr := range text {

		if inComment {
			strtemp = ""
			if chr == '\n' {
				inComment = false
				lineno++
			}
			continue
		}

		if inChar {
			if quoted {
				switch rune(chr) {
				case 'n':
					strtemp += "\n"
				case '"':
					strtemp += "\""
				case 'r':
					strtemp += "\r"
				case 't':
					strtemp += "\t"
				case 'e':
					strtemp += string(rune(27))
				case '\\':
					strtemp += "\\"
				default:
					strtemp += "\\"
					strtemp += string(rune(chr))
				}
				quoted = false
				continue
			}

			if rune(chr) == '\\' {
				quoted = true
				continue
			}

			strtemp += string(rune(chr))

			if chr != '\'' {
				continue
			}

			if len(strtemp) != 3 {
				activeFunction = op{tok: Token{Text: text, File: currentFile, Name: strtemp, Line: lineno, Offset: offset}, fn: nil}
				GfError("invalid number of characters in a character literal: %s", strtemp)
				continue
			}
			result = append(result, Token{Text: text, File: currentFile, Name: strtemp, Line: lineno, Offset: offset})
			strtemp = ""
			inChar = false
			continue
		}

		if inString {
			if quoted {
				switch rune(chr) {
				case 'n':
					strtemp += "\n"
				case '"':
					strtemp += "\""
				case 'r':
					strtemp += "\r"
				case 't':
					strtemp += "\t"
				case 'e':
					strtemp += string(rune(27))
				case '\\':
					strtemp += "\\"
				default:
					strtemp += "\\"
					strtemp += string(rune(chr))
				}
				quoted = false
				continue
			}

			if rune(chr) == '\\' {
				quoted = true
				continue
			}

			strtemp += string(rune(chr))

			if chr != '"' {
				continue
			}
			result = append(result, Token{Text: text, File: currentFile, Name: strtemp, Line: lineno, Offset: offset})
			strtemp = ""
			inString = false
			continue
		}

		if inRegex {
			if quoted {
				switch rune(chr) {
				case 'n':
					strtemp += "\n"
				case '"':
					strtemp += "\""
				case 'r':
					strtemp += "\r"
				case 't':
					strtemp += "\t"
				case '\\':
					strtemp += "\\"
				case '/':
					strtemp += "/"
				default:
					strtemp += "\\"
					strtemp += string(rune(chr))
				}
				quoted = false
				continue
			}

			if rune(chr) == '\\' {
				quoted = true
				continue
			}

			strtemp += string(rune(chr))
			if chr != '/' {
				continue
			}
			result = append(result, Token{Text: text, File: currentFile, Name: strtemp, Line: lineno, Offset: offset})
			strtemp = ""
			inRegex = false
			continue
		}

		if chr == ' ' || chr == '\r' || chr == '\n' || chr == '\t' {
			if len(strtemp) > 0 {
				if strtemp[0] == ':' {
					// Turn :foobar into "foobar"
					strtemp = "\"" + string(strtemp[1:]) + "\""
				}
				result = append(result, Token{Text: text, File: currentFile, Name: strtemp, Line: lineno, Offset: offset})
			}
			strtemp = ""
			if chr == '\n' {
				lineno++
			}
			continue
		}

		if chr == '\'' {
			if len(strtemp) > 0 {
				result = append(result, Token{Text: text, File: currentFile, Name: strtemp, Line: lineno, Offset: offset})
			}
			strtemp = "'"
			inChar = true
			continue
		}

		if chr == '"' {
			if len(strtemp) > 0 {
				result = append(result, Token{Text: text, File: currentFile, Name: strtemp, Line: lineno, Offset: offset})
			}
			strtemp = "\""
			inString = true
			continue
		}

		if chr == '/' && strtemp == "r" {
			strtemp = "r/"
			inRegex = true
			continue
		}

		if chr == '#' {
			if len(strtemp) > 0 {
				result = append(result, Token{Text: text, File: currentFile, Name: strtemp, Line: lineno, Offset: offset})
			}
			strtemp = "#"
			inComment = true
			continue
		}

		if chr == ';' {
			if len(strtemp) > 0 {
				result = append(result, Token{Text: text, File: currentFile, Name: strtemp, Line: lineno, Offset: offset})
				strtemp = ""
			}
			result = append(result, Token{Text: text, File: currentFile, Name: ";", Line: lineno, Offset: offset})
			continue
		}

		if chr == '[' {
			if len(strtemp) > 0 {
				result = append(result, Token{Text: text, File: currentFile, Name: strtemp, Line: lineno, Offset: offset})
				strtemp = ""
			}
			result = append(result, Token{Text: text, File: currentFile, Name: "[", Line: lineno, Offset: offset})
			squareCount++
			continue
		}

		if chr == ']' {
			if len(strtemp) > 0 {
				result = append(result, Token{Text: text, File: currentFile, Name: strtemp, Line: lineno, Offset: offset})
				strtemp = ""
			}
			squareCount--
			if squareCount < 0 {
				fmt.Println("Too many ']'s. THere must be one '[' for each ']'.")
			} else {
				result = append(result, Token{Text: text, File: currentFile, Name: "]", Line: lineno, Offset: offset})
			}
			continue
		}

		if chr == '{' {
			if len(strtemp) > 0 {
				result = append(result, Token{File: currentFile, Name: strtemp, Line: lineno, Offset: offset})
				strtemp = ""
			}
			result = append(result, Token{Text: text, File: currentFile, Name: "{", Line: lineno, Offset: offset})
			braceCount++
			continue
		}

		if chr == '}' {
			if len(strtemp) > 0 {
				result = append(result, Token{Text: text, File: currentFile, Name: strtemp, Line: lineno, Offset: offset})
				strtemp = ""
			}
			result = append(result, Token{Text: text, File: currentFile, Name: "}", Line: lineno, Offset: offset})
			braceCount--
			if braceCount < 0 {
				fmt.Println("Too many '}'s. THere must be one '{' for each '}'.")
			}
			continue
		}

		strtemp += string(rune(chr))
	}

	if len(strtemp) > 0 {
		result = append(result, Token{Text: text, File: currentFile, Name: strtemp, Line: lineno, Offset: 0})
	}

	if inString {
		GfError("Unterminated string in text")
		result = []Token{}
	}
	if inRegex {
		GfError("Unterminated regex in text")
		result = []Token{}
	}
	if squareCount != 0 {
		GfError("Invalid number of square brackets '[' ']': %d", squareCount)
		result = []Token{}
	}
	if braceCount != 0 {
		GfError("Invalid number of braces '{' '}': %d", braceCount)
		result = []Token{}
	}

	return result
}

// VariableTable is the variable table
var VariableTable = make(map[string]interface{})

// Represents a compiled function in a lambda
type op struct {
	fn  func()
	tok Token
}

// Dictionary of name string to operator functions
var ops map[string]interface{} = make(map[string]interface{})

// ValueStack Holds the values that operations operate on
var ValueStack *Stack = &Stack{}

// OffsetStack tracks the starting stack position of an array literal
var OffsetStack *Stack = &Stack{}

// CallStack tracks the user-level calls
var CallStack *Stack = &Stack{}

// The evaluator keeps evaluating while this is true
var loop bool = true

// The REPL keeps running while this is true
var quit bool = false

// Step the evaluator at each word
var step bool = false

// The currently active function - used for error messages
var activeFunction op

// Eval evaluates a GoForth program
func Eval(funcs []op) {
	for _, f := range funcs {
		oldFunc := activeFunction
		activeFunction = f
		if step {
			fmt.Printf(">>>>>> Next func is '%s' input stack is:\n", f.tok.Name)
			ValueStack.Print()
			fmt.Print("step> ")
			cmd, _ := ReadLn()
			if strings.TrimSpace(cmd) == "q" {
				step = false
			}
		}

		f.fn()
		activeFunction = oldFunc
		if !loop {
			break
		}
	}
}

// Load and evaluate a script file
func LoadFile(fileToRun string) {
	ok, _ := regexp.MatchString("\\.gf$", fileToRun)
	if !ok {
		fileToRun += ".gf"
	}

	oldFile := currentFile
	currentFile = fileToRun
	lineno = 1
	text, err := ioutil.ReadFile(fileToRun)
	if err != nil {
		GfError("Error loading script:", err)
		return
	}
	fields := ParseLine(string(text))
	_, body := Compile(fields, 0, "")
	CallStack.Push(Token{File: fileToRun, Name: fileToRun, Line: 1, Offset: 0})
	Eval(body)
	CallStack.Pop("scriptExit")
	lineno = 1
	currentFile = oldFile
}

//
// Function to dynamically dispatch a function. The function argument
// can be a op, a func() or a string nameing a function or script.
//
func InvokeDynamic(fn interface{}, tok Token) {
	var name string
	switch fn := fn.(type) {
	case op:
		fn.fn()
		return
	case func():
		fn()
		return
	case string:
		name = fn
	default:
		name = fmt.Sprint(fn)
	}

	val, in := VariableTable[name]
	if in {
		switch fn := val.(type) {
		case op:
			fn.fn()
		case func():
			fn()
		default:
			activeFunction = op{fn: nil, tok: tok}
			GfError("argument must be a function, not %t", val)
		}
	}

	opval, in := ops[name]
	if in {
		switch fn := opval.(type) {
		case op:
			fn.fn()
		case func():
			fn()
		default:
			activeFunction = op{fn: nil, tok: tok}
			GfError("argument must be a function, not %t", val)
		}
	}

	LoadFile(name)
}

//
// GfError handles GoForth errors using the value of the activeFunction variable for source context.
//
func GfError(str string, a ...interface{}) {
	fmtStr := colorRed + "Calling '" + activeFunction.tok.Name + "': " + str + colorReset + "\n"
	fmt.Printf(fmtStr, a...)
	fmt.Printf("%sAt: %s:%d\tfunc: '%s'%s\n", colorRed, activeFunction.tok.File, activeFunction.tok.Line, activeFunction.tok.Name, colorReset)
	codeline, pos := activeFunction.tok.GetCodeLine()
	if codeline != "" {
		fmt.Printf(colorRed+">> %s\n"+colorReset, codeline)
		padding := ">>"
		for pos > 0 {
			padding += " "
			pos--
		}
		padding += "^\n"
		fmt.Printf(colorRed+"%s"+colorReset, padding)
	}
	if CallStack.index > 0 {
		for i := CallStack.index - 1; i >= 0; i-- {
			tok := CallStack.Value[i].(Token)
			fmt.Printf("%sAt: %s:%d\tfunc: '%s'%s\n", colorRed, tok.File, tok.Line, tok.Name, colorReset)
		}
	}
	loop = false
}

// Compile a list of tokens into a lambda
func Compile(fields []Token, start int, term string) (int, []op) {

	var result = make([]op, 0, len(fields))
	var funcName string

	var index int = start
	for index < len(fields) {

		f := fields[index]

		if f.Name == "def" || f.Name == "DEFINE" {
			// Defining a function

			index++
			if index >= len(fields) {
				activeFunction = op{fn: nil, tok: f}
				GfError("missing function name after 'def', syntax is: DEFINE <name> == ... ;")
				return 0, nil
			}

			funcName = fields[index].Name

			index++
			if index >= len(fields) && !(fields[index].Name == "=" || fields[index].Name == "==") {
				activeFunction = op{fn: nil, tok: f}
				GfError("missing '==' in function definition; syntax is: DEFINE <name> == ... ;")
				return 0, nil
			}

			// Local vars that get captured in the closure
			var bodyPtr *[]op
			var defToken = fields[index-1]

			ops[funcName] = func() {
				CallStack.Push(defToken)
				Eval(*bodyPtr)
				CallStack.Pop("funcExit")
			}

			index++
			if index >= len(fields) {
				activeFunction = op{fn: nil, tok: f}
				GfError("body for function '%s' is missing", funcName)
				delete(ops, funcName)
				return 0, nil
			}

			offset, body := Compile(fields, index, ";")
			bodyPtr = &body

			index = offset
			continue
		}

		if f.Name == "{" {
			offset, body := Compile(fields, index+1, "}")
			bodyPtr := &body
			wrapper := op{fn: func() { Eval(*bodyPtr) }, tok: f}
			result = append(result, op{fn: func() { ValueStack.Push(wrapper) }, tok: f})
			index = offset
			continue
		}

		if f.Name == term {
			return index + 1, result
		}

		isFloat, _ := regexp.MatchString("^-?[0-9]+\\.[0-9]+(e-?[0-9]+)?$", f.Name)
		isNumber, _ := regexp.MatchString("^-?[0-9,_]+$", f.Name)
		isString := f.Name[0] == '"'
		isVarSet, _ := regexp.MatchString("^\\![^ ]+$", f.Name)
		isVarGet, _ := regexp.MatchString("^[@$][^ ]+$", f.Name)
		isFuncCall, _ := regexp.MatchString("^\\&[^ ]+$", f.Name)
		isRegex := len(f.Name) > 1 && f.Name[0] == 'r' && f.Name[1] == '/'
		isChar := f.Name[0] == '\''
		if isFloat {
			num, _ := strconv.ParseFloat(f.Name, 64)
			result = append(result, op{fn: func() { ValueStack.Push(num) }, tok: f})
		} else if isNumber {
			sval := strings.ReplaceAll(f.Name, ",", "")
			sval = strings.ReplaceAll(sval, "_", "")
			num, _ := strconv.Atoi(sval)
			result = append(result, op{fn: func() {
				ValueStack.Value[ValueStack.index] = num
				ValueStack.index++
			}, tok: f})
		} else if isString {
			str := strings.Trim(f.Name, "\"")
			result = append(result, op{fn: func() { ValueStack.Push(str) }, tok: f})
		} else if isVarSet {
			str := string(f.Name[1:])
			result = append(result, op{fn: (func() {
				VariableTable[str] = ValueStack.Pop("valueToStore")
			}), tok: f})
		} else if isVarGet {
			str := string(f.Name[1:])
			result = append(result, op{fn: (func() {
				val, in := VariableTable[str]
				if !in {
					val, in := ops[str]
					if in {
						ValueStack.Push(val)
					} else {
						activeFunction = op{fn: nil, tok: f}
						GfError("variable '%s' doesn't exist.", f.Name)
					}
				} else {
					ValueStack.Push(val)
				}
			}), tok: f})
		} else if isFuncCall { // dynamic call e.g. &foo calls "foo"
			str := string(f.Name[1:])
			tok := f
			result = append(result, op{fn: func() {
				InvokeDynamic(str, tok)
			}, tok: f})
		} else if isRegex {
			str := strings.Trim(string(f.Name[1:]), "/")
			reLiteral, err := regexp.Compile(str)
			if err != nil {
				activeFunction = op{fn: nil, tok: f}
				GfError("error compiling regex /%s/: %s", str, err)
			} else {
				result = append(result, op{fn: func() {
					ValueStack.Push(reLiteral)
				}, tok: f})
			}
		} else if isChar {
			charToPush := rune(f.Name[1])
			result = append(result, op{fn: func() {
				ValueStack.Push(charToPush)
			}, tok: f})
		} else if f.Name == "quit" {
			result = append(result, op{fn: func() {
				loop = false
				quit = false
			}, tok: f})
		} else if f.Name == "->" {
			index++
			if index >= len(fields) {
				activeFunction = op{fn: nil, tok: f}
				GfError("missing variable name after '->', syntax is: ... -> foo ;")
				return 0, nil
			}
			varName := fields[index].Name
			result = append(result, op{fn: func() {
				val := ValueStack.Pop("valueToStore")
				if !loop {
					return
				}
				VariableTable[varName] = val
			}, tok: f})
		} else if f.Name == "IMPORT" {
			index++
			if index >= len(fields) {
				activeFunction = op{fn: nil, tok: f}
				GfError("missing variable name after 'IMPORT', syntax is: IMPORT foo")
				return 0, nil
			}
			fileName := fields[index].Name
			LoadFile(fileName)

		} else {
			fn, exists := ops[f.Name]
			if !exists {
				activeFunction = op{fn: nil, tok: f}
				GfError("Undefined function '%s'", f.Name)
			} else {
				switch fn := fn.(type) {
				case op:
					result = append(result, fn)
				case func():
					result = append(result, op{fn: fn, tok: f})
				default:
					GfError("compiling '%s': expected func(), not %t", f.Name, fn)
				}
			}
		}
		index++
	}

	return index, result
}

// Compare two items polymorphically
func Compare(v1 interface{}, v2 interface{}) int {
	if v1 == nil {
		if v2 == nil {
			return 0
		}
		return -1
	}

	if v2 == nil {
		if v1 == nil {
			return 0
		}
		return -1
	}

	switch targetType := v2.(type) {
	case reflect.Type:
		switch valueType := v1.(type) {
		case reflect.Type:
			if valueType == targetType {
				return 0
			}
			return 1
		default:
			if reflect.TypeOf(v1) == targetType {
				return 0
			}
			return 1
		}
	}

	switch x := v1.(type) {
	case float64:
		switch y := v2.(type) {
		case string:
			fval, err := strconv.ParseFloat(y, 64)
			if err != nil {
				GfError("comparing %v and %v: %s", v1, v2, err)
				return -1
			}
			if x > fval {
				return 1
			} else if x < fval {
				return -1
			} else {
				return 0
			}
		case int:
			fval := float64(y)
			if x > fval {
				return 1
			} else if x < fval {
				return -1
			} else {
				return 0
			}
		case float64:
			fval := float64(y)
			if x > fval {
				return 1
			} else if x < fval {
				return -1
			} else {
				return 0
			}
		}
	case int:
		switch y := v2.(type) {
		case int:
			if x > y {
				return 1
			} else if x < y {
				return -1
			} else {
				return 0
			}
		case string:
			ival, err := strconv.Atoi(y)
			if err != nil {
				GfError("comparing %v and %v: %s", v1, v2, err)
				return -1
			}
			if x > ival {
				return 1
			} else if x < ival {
				return -1
			} else {
				return 0
			}
		case float64:
			ival := int(y)
			if x > ival {
				return 1
			} else if x < ival {
				return -1
			} else {
				return 0
			}
		}
	case string:
		switch y := v2.(type) {
		case string:
			if x > y {
				return 1
			} else if x < y {
				return -1
			} else {
				return 0
			}
		default:
			sval := fmt.Sprint(y)
			if x > sval {
				return 1
			} else if x < sval {
				return -1
			} else {
				return 0
			}
		}
	case []interface{}:
		switch y := v2.(type) {
		case []interface{}:
			if len(x) > len(y) {
				return 1
			} else if len(x) < len(y) {
				return -1
			} else {
				for i, v := range x {
					r := Compare(v, y[i])
					if r != 0 {
						return r
					}
				}
				return 0
			}
		default:
			GfError("Cannot compare %v and %v\n", v1, v2)
			return -1
		}
	default:
		GfError("Cannot compare %v and %v\n", v1, v2)
		return -1
	}

	GfError("Cannot compare %v and %v\n", v1, v2)
	return -1
}

func isFalse(val interface{}) bool {
	return !isTrue(val)
}

func isTrue(val interface{}) bool {
	if val == nil {
		return false
	}
	switch v := val.(type) {
	case bool:
		return v
	case int:
		return v != 0
	case float64:
		return v != 0
	case []interface{}:
		return len(v) != 0
	case map[interface{}]interface{}:
		return len(v) != 0
	case map[string]interface{}:
		return len(v) != 0
	case string:
		return len(v) != 0
	default:
		GfError("argument must be bool, int, float, array or string, not '%s' [%t]", v, v)
		return false
	}
}

func stringify(val interface{}) string {
	if val == nil {
		return ""
	}

	switch val := val.(type) {
	case int:
		return fmt.Sprintf("%d", val)
	case float64:
		return fmt.Sprintf("%f", val)
	case string:
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}

var listType = reflect.TypeOf(make([]interface{}, 0))

func binRecHelper(val interface{}, ifProg, thenProg, recProg, endProg func()) {
	ValueStack.Push(val)
	ifProg()
	r := ValueStack.Pop("ifProgResult")
	if !loop {
		return
	}
	if isTrue(r) {
		ValueStack.Push(val)
		thenProg()
		return
	}
	ValueStack.Push(val)
	recProg()
	val1 := ValueStack.Pop("rec1Val")
	val2 := ValueStack.Pop("rec2val")
	binRecHelper(val1, ifProg, thenProg, recProg, endProg)
	val1 = ValueStack.Pop("rec1Result")
	if !loop {
		return
	}
	binRecHelper(val2, ifProg, thenProg, recProg, endProg)
	val2 = ValueStack.Pop("rec2Result")
	if !loop {
		return
	}
	ValueStack.Push(val1)
	ValueStack.Push(val2)

	endProg()
}

func main() {

	ops["step"] = func() {
		step = true
	}

	ops["compare"] = func() {
		v2 := ValueStack.Pop("operand2")
		v1 := ValueStack.Pop("operand1")
		if !loop {
			return
		}
		ValueStack.Push(Compare(v1, v2))
	}

	ops["=="] = func() {
		v2 := ValueStack.Pop("operand2")
		v1 := ValueStack.Pop("operand1")
		if !loop {
			return
		}
		ValueStack.Push(Compare(v1, v2) == 0)
	}

	ops["!="] = func() {
		v2 := ValueStack.Pop("operand2")
		v1 := ValueStack.Pop("operand1")
		if !loop {
			return
		}
		ValueStack.Push(Compare(v1, v2) != 0)
	}

	ops[">"] = func() {
		v2 := ValueStack.Pop("operand2")
		v1 := ValueStack.Pop("operand1")
		if !loop {
			return
		}
		ValueStack.Push(Compare(v1, v2) > 0)
	}

	ops[">="] = func() {
		v2 := ValueStack.Pop("operand2")
		v1 := ValueStack.Pop("operand1")
		if !loop {
			return
		}
		ValueStack.Push(Compare(v1, v2) >= 0)
	}

	ops["<"] = func() {
		v2 := ValueStack.Pop("operand2")
		v1 := ValueStack.Pop("operand1")
		if !loop {
			return
		}
		ValueStack.Push(Compare(v1, v2) < 0)
	}

	ops["<="] = func() {
		v2 := ValueStack.Pop("operand2")
		v1 := ValueStack.Pop("operand1")
		if !loop {
			return
		}
		ValueStack.Push(Compare(v1, v2) <= 0)
	}

	ops["and"] = func() {
		v2 := ValueStack.Pop("operand2")
		v1 := ValueStack.Pop("operand1")
		if !loop {
			return
		}
		ValueStack.Push(isTrue(v1) && isTrue(v2))
	}

	ops["or"] = func() {
		v2 := ValueStack.Pop("operand2")
		v1 := ValueStack.Pop("operand1")
		if !loop {
			return
		}
		ValueStack.Push(isTrue(v1) || isTrue(v2))
	}

	ops["keys"] = func() {
		val := ValueStack.Pop("dictionary")
		if !loop {
			return
		}
		switch val := val.(type) {
		case map[interface{}]interface{}:
			keys := make([]interface{}, 0, len(ops))
			for k := range val {
				keys = append(keys, k)
			}
			ValueStack.Push(keys)
		case map[string]interface{}:
			keys := make([]interface{}, 0, len(ops))
			for k := range val {
				keys = append(keys, k)
			}
			ValueStack.Push(keys)
		default:
			GfError("The 'keys' function can only be used on a map, not %v", reflect.TypeOf(val))
		}
	}

	ops["help"] = func() {
		count := 0

		for k := range ops {
			fmt.Printf("%15s", k)
			if count == 6 {
				fmt.Println()
				count = 0
				continue
			}
			count++
		}
		fmt.Println()
	}

	ops["repeat"] = func() {
		val := ValueStack.Pop("program")
		if !loop {
			return
		}
		var body func()
		switch val := val.(type) {
		case op:
			body = val.fn
		case func():
			body = val
		default:
			GfError("The second argument to 'repeat' must be a lambda, to %t", val)
			return
		}

		iterval := ValueStack.Pop("itercount")
		var itercount int
		switch iterval := iterval.(type) {
		case int:
			itercount = iterval
		case float64:
			itercount = int(iterval)
		default:
			GfError("The first argument to 'repeat' must be an integer.")
			return
		}

		oldval, ok := VariableTable["_"]
		for i := 0; i < itercount; i++ {
			if !loop {
				break
			}
			VariableTable["_"] = i // BUGBUGBUG - do we really want to do this?
			body()
		}
		if ok {
			VariableTable["_"] = oldval
		}
	}

	ops["if"] = func() {
		val := ValueStack.Pop("condVal")
		if !loop {
			return
		}
		var body func()
		switch val := val.(type) {
		case op:
			body = val.fn
		case func():
			body = val
		default:
			GfError("The first argument to 'repeat' must be a lambda, to %t", val)
		}

		cond := ValueStack.Pop("thenPart")
		if !loop {
			return
		}

		if isTrue(cond) {
			body()
		}
	}

	ops["ifte"] = func() {
		elsePartVal := ValueStack.Pop("elsePart")
		ifPartVal := ValueStack.Pop("thenPart")
		cond := ValueStack.Pop("condVal")
		if !loop {
			return
		}

		var elsePart func()
		switch val := elsePartVal.(type) {
		case op:
			elsePart = val.fn
		case func():
			elsePart = val
		default:
			GfError("The 'if' 'else' argument must be a lambda, to %t", val)
		}

		var ifPart func()
		switch val := ifPartVal.(type) {
		case op:
			ifPart = val.fn
		case func():
			ifPart = val
		default:
			GfError("The 'if' 'then' argument must be a lambda, to %t", val)
		}

		if isTrue(cond) {
			ifPart()
		} else {
			elsePart()
		}
	}

	ops["case"] = func() {
		pattern := ValueStack.Pop("pattern")
		val := ValueStack.Pop("valToMatch")
		if !loop {
			return
		}

		switch pattern := pattern.(type) {
		case []interface{}:
			isPat := true
			var pe interface{}
			for _, v := range pattern {
				if isPat {
					pe = v
				} else {
					switch pe := pe.(type) {
					case op:
						ValueStack.Push(val)
						pe.fn()
						testResult := ValueStack.Pop("testResult")
						if !loop {
							return
						}
						if isTrue(testResult) {
							switch v := v.(type) {
							case op:
								ValueStack.Push(val)
								v.fn()
							case func():
								ValueStack.Push(val)
								v()
							default:
								ValueStack.Push(v)
							}
							return
						}

					case func():
						ValueStack.Push(val)
						pe()
						testResult := ValueStack.Pop("testResult")
						if !loop {
							return
						}
						if isTrue(testResult) {
							switch v := v.(type) {
							case op:
								ValueStack.Push(val)
								v.fn()
							case func():
								ValueStack.Push(val)
								v()
							default:
								ValueStack.Push(v)
							}
							return
						}
					case *regexp.Regexp:
						if pe.MatchString(fmt.Sprintf("%v", val)) {
							switch v := v.(type) {
							case op:
								ValueStack.Push(val)
								v.fn()
							case func():
								ValueStack.Push(val)
								v()
							default:
								ValueStack.Push(v)
							}
							return
						}
					case reflect.Type:
						if pe == reflect.TypeOf(val) {
							switch v := v.(type) {
							case op:
								ValueStack.Push(val)
								v.fn()
							case func():
								ValueStack.Push(val)
								v()
							default:
								ValueStack.Push(v)
							}
							return
						}
					default:
						if Compare(pe, val) == 0 {
							switch v := v.(type) {
							case op:
								ValueStack.Push(val)
								v.fn()
							case func():
								ValueStack.Push(val)
								v()
							default:
								ValueStack.Push(v)
							}
							return
						}
					}
				}
				isPat = !isPat
			}
		default:
			GfError("The second argument to 'case' must be a list")
		}
	}

	ops["true"] = func() { ValueStack.Push(true) }

	// replaces the TOS with true
	ops["true!"] = func() { ValueStack.Value[ValueStack.index-1] = true }

	ops["false"] = func() { ValueStack.Push(false) }

	// replaces the TOS with false
	ops["true!"] = func() { ValueStack.Value[ValueStack.index-1] = false }

	ops["type"] = func() {
		val := ValueStack.Pop("valToGetTypeOf")
		if !loop {
			return
		}
		ValueStack.Push(reflect.TypeOf(val))
	}

	ops["^int"] = func() {
		ValueStack.Push(reflect.TypeOf(1))
	}

	ops["^float"] = func() {
		ValueStack.Push(reflect.TypeOf(1.0))
	}

	ops["^string"] = func() {
		ValueStack.Push(reflect.TypeOf(""))
	}

	ops["^lambda"] = func() {
		ValueStack.Push(reflect.TypeOf(func() {}))
	}

	ops["^list"] = func() {
		ValueStack.Push(listType)
	}

	ops["^bool"] = func() {
		ValueStack.Push(reflect.TypeOf(true))
	}

	ops["^byte"] = (func() {
		ValueStack.Push(reflect.TypeOf("a"[0]))
	})

	ops["^type"] = func() {
		// Having to do this is crazy...
		ValueStack.Push(reflect.TypeOf(reflect.TypeOf(1)))
	}

	ops["is"] = func() {
		v2 := ValueStack.Pop("targetType")
		v1 := ValueStack.Pop("value")
		if !loop {
			return
		}

		switch v2 := v2.(type) {
		case reflect.Type:
			ValueStack.Push(reflect.TypeOf(v1) == v2)
		default:
			GfError("The second argument to 's' must be a type.")
		}
	}

	ops["&"] = func() {
		val := ValueStack.Pop("programToinvoke")
		if !loop {
			return
		}

		InvokeDynamic(val, activeFunction.tok)
	}

	ops["apply2"] = func() {
		fnval := ValueStack.Pop("programToinvoke")
		if !loop {
			return
		}

		var fn func()
		switch fnval := fnval.(type) {
		case op:
			fn = fnval.fn
		case func():
			fn = fnval
		default:
			GfError("The argument to '&2' must be of type lambda(); not %t\n", fnval)
		}

		val := ValueStack.Pop("secondVal")
		fn()
		ValueStack.Push(val)
		fn()
	}

	ops["apply3"] = func() {
		fnval := ValueStack.Pop("programToinvoke")
		if !loop {
			return
		}

		var fn func()
		switch fnval := fnval.(type) {
		case op:
			fn = fnval.fn
		case func():
			fn = fnval
		default:
			GfError("The argument to 'apply3' must be of type lambda(); not %t\n", fnval)
		}

		val3 := ValueStack.Pop("thirdVal")
		val2 := ValueStack.Pop("secondVal")
		fn()
		ValueStack.Push(val2)
		fn()
		ValueStack.Push(val3)
		fn()
	}

	//C [X Y ..] unstack -> ..Y X
	//C The list [X Y ..] becomes the new stack.
	ops["unstack"] = func() {
		listVal := ValueStack.Pop("listVar")
		if !loop {
			return
		}

		var list []interface{}
		switch listVal := listVal.(type) {
		case []interface{}:
			list = listVal
		default:
			GfError("the argument to 'stack' must be a list of values [ ... ]")
		}

		for _, v := range list {
			ValueStack.Push(v)
		}
	}

	//C .. X Y Z -> .. X Y Z [Z Y X ..]
	//C Pushes the stack as a list.
	ops["stack"] = func() {
		result := make([]interface{}, ValueStack.index)
		for i := ValueStack.index - 1; i >= 0; i-- {
			result[i] = ValueStack.Value[i]
		}
		ValueStack.Push(result)
	}

	ops["over"] = func() {
		if ValueStack.index > 1 {
			ValueStack.Push(ValueStack.Value[ValueStack.index-2])
		} else {
			GfError("there need to be at least 2 elements on the stack to call 'over'.")
		}
	}

	ops["["] = func() {
		OffsetStack.Push(ValueStack.index)
	}

	ops["]"] = func() {

		startIndex := OffsetStack.Pop("offsetOfStartOfList").(int)
		if startIndex < 0 {
			ValueStack.Push(make([]interface{}, 0))
			return
		}

		endIndex := ValueStack.index
		if endIndex <= startIndex {
			ValueStack.Push(make([]interface{}, 0))
			return
		}

		index := endIndex - startIndex
		arr := make([]interface{}, index)
		for index > 0 {
			index--
			arr[index] = ValueStack.Pop("listElement")
		}

		ValueStack.Push(arr)
	}

	ops["+"] = func() {
		v2 := ValueStack.Pop("operand2")
		v1 := ValueStack.Pop("operand1")
		if !loop {
			return
		}
		switch x := v1.(type) {
		case float64:
			switch y := v2.(type) {
			case string:
				fval, err := strconv.ParseFloat(y, 64)
				if err != nil {
					GfError("Can't add %v and %v: %s", v1, v2, err)
					return
				}
				ValueStack.Push(x + fval)
			case int:
				ValueStack.Push(x + float64(y))
			case float64:
				ValueStack.Push(x + y)
			}
		case int:
			switch y := v2.(type) {
			case int:
				ValueStack.Push(x + y)
			case string:
				ival, err := strconv.Atoi(y)
				if err != nil {
					GfError("Can't add %v and %v: %s", v1, v2, err)
					return
				}
				ValueStack.Push(x + ival)
			case float64:
				ValueStack.Push(x + int(y))
			}
		case string:
			switch y := v2.(type) {
			case string:
				ValueStack.Push(x + y)
			default:
				ValueStack.Push(x + fmt.Sprint(y))
			}
		case []interface{}:
			ValueStack.Push(append(x, v2))
		}
	}

	ops["uncons"] = func() {
		vect := ValueStack.Pop("vector")
		if !loop {
			return
		}

		if vect == nil {
			ValueStack.Push(nil)
			ValueStack.Push(make([]interface{}, 0))
		}

		switch x := vect.(type) {
		case []interface{}:
			if len(x) > 0 {
				ValueStack.Push(x[0])
				ValueStack.Push(x[1:])
			} else {
				ValueStack.Push(nil)
				ValueStack.Push(make([]interface{}, 0))
			}
		default:
			ValueStack.Push(vect)
			ValueStack.Push(make([]interface{}, 0))
		}
	}

	ops["append"] = func() {
		v2 := ValueStack.Pop("list2")
		v1 := ValueStack.Pop("list1")
		if !loop {
			return
		}
		switch x := v1.(type) {
		case []interface{}:
			switch y := v2.(type) {
			case []interface{}:
				for _, v := range y {
					x = append(x, v)
				}
				ValueStack.Push(x)
			default:
				ValueStack.Push(append(x, v2))
			}
		default:
			result := []interface{}{v1}
			switch y := v2.(type) {
			case []interface{}:
				for _, v := range y {
					result = append(result, v)
				}
				ValueStack.Push(result)
			default:
				ValueStack.Push(append(result, v2))
			}
		}
	}

	ops["cons"] = func() {
		v2 := ValueStack.Pop("list")
		v1 := ValueStack.Pop("valToCons")
		if !loop {
			return
		}
		switch v2 := v2.(type) {
		case []interface{}:
			v2 = append(v2, 0)
			copy(v2[1:], v2)
			v2[0] = v1
			ValueStack.Push(v2)
		default:
			GfError("the second argument to cons must be a list, not [%t]", v2)
		}
	}

	ops["list:split"] = func() {
		v2 := ValueStack.Pop("progOrValue")
		v1 := ValueStack.Pop("listToSplit")
		if !loop {
			return
		}

		smaller := make([]interface{}, 0)
		larger := make([]interface{}, 0)

		switch x := v1.(type) {
		case []interface{}:
			switch y := v2.(type) {
			case op:
				for _, v := range x {
					ValueStack.Push(v)
					y.fn()
					cond := ValueStack.Pop("progResult")
					switch cond := cond.(type) {
					case bool:
						if cond {
							larger = append(larger, v)
						} else {
							smaller = append(smaller, v)
						}
					default:
						GfError("condition expression should return a boolean, not '%s' [%t]", cond, cond)
					}
				}
				ValueStack.Push(smaller)
				ValueStack.Push(larger)
			case func():
				for _, v := range x {
					ValueStack.Push(v)
					y()
					cond := ValueStack.Pop("progResult")
					switch cond := cond.(type) {
					case bool:
						if cond {
							larger = append(larger, v)
						} else {
							smaller = append(smaller, v)
						}
					default:
						GfError("condition expression should return a boolean, not '%s' [%t]", cond, cond)
					}
				}
				ValueStack.Push(smaller)
				ValueStack.Push(larger)
			case int:
				all := make([]interface{}, 0)
				curr := make([]interface{}, 0, y)
				for _, v := range x {
					if len(curr) > 0 && len(curr)%y == 0 {
						all = append(all, curr)
						curr = make([]interface{}, 0, y)
						curr = append(curr, v)

					} else {
						curr = append(curr, v)
					}
				}
				if len(curr) > 0 {
					all = append(all, curr)
				}
				ValueStack.Push(all)

			default:
				GfError("expression argument should be a lambda, not '%s' [%t]", y, y)
			}
		default:
			GfError("the first argument must be a list ([]interface{}), not '%v' [%t]", x, x)
		}
	}

	ops["false?"] = func() {
		v1 := ValueStack.Pop("valToTest")
		if !loop {
			return
		}
		ValueStack.Push(isFalse(v1))
	}

	ops["true?"] = func() {
		v1 := ValueStack.Pop("valToTest")
		if !loop {
			return
		}
		ValueStack.Push(isTrue(v1))
	}

	ops["true!"] = func() {
		ValueStack.Pop("valToConvert")
		if !loop {
			return
		}
		ValueStack.Push(true)
	}

	ops["*"] = func() {
		v2 := ValueStack.Pop("operand2")
		v1 := ValueStack.Pop("operand1")
		switch x := v1.(type) {
		case int:
			switch y := v2.(type) {
			case int:
				ValueStack.Push(x * y)
			case float64:
				ValueStack.Push(x * int(y))
			default:
				GfError("Cannot multiply '%s' and '%s'", v1, v2)
			}
		case float64:
			switch y := v2.(type) {
			case int:
				ValueStack.Push(x * float64(y))
			case float64:
				ValueStack.Push(x * y)
			default:
				GfError("Cannot multiply '%s' and '%s'", v1, v2)
			}
		case string:
			switch y := v2.(type) {
			case int:
				rstr := ""
				for i := 0; i < y; i++ {
					rstr += x
				}
				ValueStack.Push(rstr)
			default:
				GfError("Cannot multiply '%s' and '%s'", v1, v2)
			}
		case []interface{}:
			switch y := v2.(type) {
			case int:
				result := make([]interface{}, 0, len(x)*y)
				for i := 0; i < y; i++ {
					result = append(result, x)
				}
				ValueStack.Push(result)
			default:
				GfError("Cannot multiply '%s' and '%s'", v1, v2)
			}
		default:
			GfError("Cannot multiply '%s' and '%s'", v1, v2)
		}
	}

	ops["-"] = func() {
		v2 := ValueStack.Pop("operand2")
		v1 := ValueStack.Pop("operand1")
		if !loop {
			return
		}
		switch x := v1.(type) {
		case int:
			switch y := v2.(type) {
			case int:
				ValueStack.Push(x - int(y))
			case float64:
				ValueStack.Push(float64(x) - y)
			default:
				GfError("Can't subtract a value of type [%t] from an integer.")
			}
		case float64:
			switch y := v2.(type) {
			case int:
				ValueStack.Push(x - float64(y))
			case float64:
				ValueStack.Push(x - y)
			default:
				GfError("Can't subtract a value of type [%t] from a floating point number.")
			}
		case string:
			switch y := v2.(type) {
			case int:
				if y < 0 {
					y = len(x) + y
				}
				if y <= len(x) {
					ValueStack.Push(string(x[y:]))
				} else {
					ValueStack.Push("")
				}
			case float64:
				var i int
				if y < 0 {
					i = len(x) + int(y)
				} else {
					i = int(y)
				}
				if i <= len(x) {
					ValueStack.Push(string(x[i:]))
				} else {
					ValueStack.Push("")
				}
			case string:
				start := strings.Index(x, y)
				if start > -1 {
					ValueStack.Push(string(x[0:start]) + string(x[start+len(y):]))
				}
			default:
				GfError("Can't subtract a value of type [%t] from a string.")
			}
		case []interface{}:
			switch y := v2.(type) {
			case int:
				if y < 0 {
					y = len(x) + y
				}
				if y <= len(x) {
					ValueStack.Push((x[y:]))
				} else {
					ValueStack.Push(nil)
				}
			case float64:
				var i int
				if y < 0 {
					i = len(x) + int(y)
				} else {
					i = int(y)
				}
				if i <= len(x) {
					ValueStack.Push(x[i:])
				} else {
					ValueStack.Push(nil)
				}
			default:
				GfError("Can't subtract a value of type [%t] from a list.")
			}

		default:
			GfError("Cannot subtract %s [%t] and %s [%t]", v1, v1, v2, v2)
		}
	}

	ops["/"] = func() {
		v2 := ValueStack.Pop("operand2")
		v1 := ValueStack.Pop("operand1")
		if !loop {
			return
		}
		switch x := v1.(type) {
		case int:
			switch y := v2.(type) {
			case int:
				if y == 0 {
					GfError("division by zero.")
					return
				}
				ValueStack.Push(x / y)
			case float64:
				if y == 0 {
					GfError("division by zero.")
					return
				}
				ValueStack.Push(x / int(y))
			}
		case float64:
			switch y := v2.(type) {
			case int:
				if y == 0 {
					GfError("division by zero.")
					return
				}
				ValueStack.Push(int(x) / y)
			case float64:
				if y == 0 {
					GfError("division by zero.")
					return
				}
				ValueStack.Push(x / y)
			}
		default:
			GfError("Cannot divide %s by %s", v1, v2)
		}
	}

	ops["%"] = func() {
		v2 := ValueStack.Pop("operand2")
		v1 := ValueStack.Pop("operand1")
		if !loop {
			return
		}
		switch x := v1.(type) {
		case int:
			switch y := v2.(type) {
			case int:
				ValueStack.Push(x % y)
			case float64:
				ValueStack.Push(x % int(y))
			default:
				GfError("Cannot use % with operands %s and %s", v1, v2)
			}
		case float64:
			switch y := v2.(type) {
			case int:
				ValueStack.Push(int(x) % y)
			case float64:
				ValueStack.Push(int(x) % int(y))
			default:
				GfError("Cannot use % with operands %s and %s", v1, v2)
			}
		default:
			GfError("Cannot use % with operands %s and %s", v1, v2)
		}
	}

	ops[".."] = func() {
		v2 := ValueStack.Pop("operand2")
		v1 := ValueStack.Pop("operand1")
		if !loop {
			return
		}
		switch v1 := v1.(type) {
		case int:
			switch v2 := v2.(type) {
			case int:
				incr := 1
				if v1 > v2 {
					incr = -1
				}
				size := int(math.Abs(float64(v1) - float64(v2)))
				result := make([]interface{}, 0, size)
				result = append(result, v1)
				for v1 != v2 {
					v1 += incr
					result = append(result, v1)
				}
				ValueStack.Push(result)
				return
			default:
				GfError("two integer arguments are required")
			}
		default:
			GfError("two integer arguments are required")
		}
	}

	ops["..."] = func() {
		v3 := ValueStack.Pop("incr")
		v2 := ValueStack.Pop("finish")
		v1 := ValueStack.Pop("start")
		if !loop {
			return
		}
		var incr int
		switch v3 := v3.(type) {
		case int:
			incr = v3
			if incr < 1 {
				GfError("the range increment must be greater than 0, not %d", incr)
				return
			}
		default:
			GfError("the third range argument 'increment' must be an integer")
			return
		}

		switch v1 := v1.(type) {
		case int:
			switch v2 := v2.(type) {
			case int:
				size := int(math.Abs(float64(v1) - float64(v2)))
				result := make([]interface{}, 0, size)
				result = append(result, v1)
				for size-incr > 0 {
					if v1 < v2 {
						v1 += incr
					} else {
						v1 -= incr
					}
					size -= incr
					result = append(result, v1)
				}
				ValueStack.Push(result)
				return
			default:
				GfError("two integer arguments are required")
			}
		default:
			GfError("two integer arguments are required")
		}
	}

	ops["random"] = (func() {
		ValueStack.Push(rand.Int())
	})

	ops["list:random"] = (func() {
		num := ValueStack.Pop("numValsToGenerate")
		if !loop {
			return
		}

		if reflect.TypeOf(num) != reflect.TypeOf(1) {
			GfError("'list:random' requires an integer argument specifying the number of random numbers to generate, not %T", num)
			return
		}
		result := make([]interface{}, 0, num.(int))
		for n := num.(int); n != 0; n-- {
			result = append(result, rand.Int())
		}
		ValueStack.Push(result)
	})

	ops["list?"] = func() {
		val := ValueStack.Pop("valToTest")
		if !loop {
			return
		}
		switch val.(type) {
		case []interface{}:
			ValueStack.Push(true)
		default:
			ValueStack.Push(false)
		}
	}

	ops["string?"] = func() {
		val := ValueStack.Pop("valToTest")
		if !loop {
			return
		}
		switch val.(type) {
		case string:
			ValueStack.Push(true)
		default:
			ValueStack.Push(false)
		}
	}

	ops["int?"] = func() {
		val := ValueStack.Pop("valToTest")
		if !loop {
			return
		}
		switch val.(type) {
		case int:
			ValueStack.Push(true)
		default:
			ValueStack.Push(false)
		}
	}

	ops["float?"] = func() {
		val := ValueStack.Pop("valToTest")
		if !loop {
			return
		}
		switch val.(type) {
		case float64:
			ValueStack.Push(true)
		default:
			ValueStack.Push(false)
		}
	}

	ops["byte?"] = func() {
		val := ValueStack.Pop("valToTest")
		if !loop {
			return
		}
		switch val.(type) {
		case byte:
			ValueStack.Push(true)
		default:
			ValueStack.Push(false)
		}
	}

	ops["number?"] = func() {
		val := ValueStack.Pop("valToTest")
		if !loop {
			return
		}
		switch val.(type) {
		case float64:
			ValueStack.Push(true)
		case int:
			ValueStack.Push(true)
		case byte:
			ValueStack.Push(true)
		default:
			ValueStack.Push(false)
		}
	}

	ops["rol"] = func() {
		tos := ValueStack.Pop("tos")
		v2 := ValueStack.Pop("v2")
		v1 := ValueStack.Pop("v1")
		if !loop {
			return
		}
		ValueStack.Push(tos)
		ValueStack.Push(v1)
		ValueStack.Push(v2)
	}

	ops["swap"] = func() {
		if ValueStack.index > 1 {
			ValueStack.Value[ValueStack.index-2], ValueStack.Value[ValueStack.index-1] =
				ValueStack.Value[ValueStack.index-1], ValueStack.Value[ValueStack.index-2]
		} else {
			GfError("there must be at least 2 values on the stack to 'swap' them.")
		}
	}

	ops["swapd"] = func() {
		if ValueStack.index > 2 {
			ValueStack.Value[ValueStack.index-3], ValueStack.Value[ValueStack.index-2] =
				ValueStack.Value[ValueStack.index-2], ValueStack.Value[ValueStack.index-3]
		} else {
			GfError("there must be at least 2 values on the stack to 'swap' them.")
		}
	}

	ops["pop"] = func() {
		if ValueStack.index > 0 {
			ValueStack.index--
		} else {
			GfError("there must be at least 1 value on the stack to call 'pop'.")
		}
	}

	ops["popd"] = func() {
		if ValueStack.index > 1 {
			ValueStack.Value[ValueStack.index-2] = ValueStack.Value[ValueStack.index-1]
			ValueStack.index--
		} else {
			GfError("there must be at least 2 values on the stack to call 'popd'.")
		}
	}

	ops["small"] = func() {
		index := ValueStack.index
		if index < 1 {
			GfError("there must be at least 1 value on the stack to call 'dup'.")
			return
		}
		switch val := ValueStack.Value[index-1].(type) {
		case []interface{}:
			ValueStack.Value[index-1] = len(val) < 2
		case string:
			ValueStack.Value[index-1] = len(val) < 2
		case int:
			ValueStack.Value[index-1] = val < 2
		case float64:
			ValueStack.Value[index-1] = val < 2
		case nil:
			ValueStack.Value[index-1] = true
		case bool:
			ValueStack.Value[index-1] = true
		}
	}

	ops["cstk"] = func() { ValueStack.Reset() }

	ops["dup"] = func() {
		index := ValueStack.index
		if index > 0 {
			ValueStack.Value[index] = ValueStack.Value[index-1]
			ValueStack.index++
		} else {
			GfError("there must be at least 1 value on the stack to call 'dup'.")
		}
	}

	ops["dup2"] = func() {
		index := ValueStack.index
		if index > 1 {
			ValueStack.Value[index+1], ValueStack.Value[index] = ValueStack.Value[index-1], ValueStack.Value[index-2]
			ValueStack.index += 2
		} else {
			GfError("there must be at least 2 values on the stack to call 'dup2'.")
		}
	}

	//C X {P1} {P2} cleave -> R1 R2
	//C Executes P1 and P2, each with X on top, producing two results.
	ops["cleave"] = func() {
		prog2val := ValueStack.Pop("prog2val")
		prog1val := ValueStack.Pop("prog1val")
		val := ValueStack.Pop("valToPrint")
		if !loop {
			return
		}

		var prog1 func()
		var prog2 func()

		switch prog1val := prog1val.(type) {
		case op:
			prog1 = prog1val.fn
		case func():
			prog1 = prog1val
		default:
			GfError("the third argument to 'cleave' must be a prog")
		}

		switch prog2val := prog2val.(type) {
		case op:
			prog2 = prog2val.fn
		case func():
			prog2 = prog2val
		default:
			GfError("the second argument to 'cleave' must be a prog")
		}

		if !loop {
			return
		}

		ValueStack.Push(val)
		prog1()
		ValueStack.Push(val)
		prog2()
	}

	ops["."] = func() {
		val := ValueStack.Pop("valToPrint")
		if !loop {
			return
		}
		fmt.Println(val)
	}

	ops[".s"] = ValueStack.Print

	ops[".red"] = func() {
		val := ValueStack.Pop("valToPrint")
		if !loop {
			return
		}
		fmt.Println(colorRed+fmt.Sprintf("%v", val), colorReset)
	}

	ops[".yellow"] = func() {
		val := ValueStack.Pop("valToPrint")
		if !loop {
			return
		}
		fmt.Println(colorYellow+fmt.Sprintf("%v", val), colorReset)
	}

	ops[".green"] = func() {
		val := ValueStack.Pop("valToPrint")
		if !loop {
			return
		}
		fmt.Println(colorGreen+fmt.Sprintf("%v", val), colorReset)
	}

	ops[".blue"] = func() {
		val := ValueStack.Pop("valToPrint")
		if !loop {
			return
		}
		fmt.Println(colorBlue+fmt.Sprintf("%v", val), colorReset)
	}

	ops[".purple"] = func() {
		val := ValueStack.Pop("valToPrint")
		if !loop {
			return
		}
		fmt.Println(colorPurple+fmt.Sprintf("%v", val), colorReset)
	}

	ops[".white"] = func() {
		val := ValueStack.Pop("valToPrint")
		if !loop {
			return
		}
		fmt.Println(colorWhite+fmt.Sprintf("%v", val), colorReset)
	}

	ops[".cyan"] = func() {
		val := ValueStack.Pop("valToPrint")
		if !loop {
			return
		}
		fmt.Println(colorCyan+fmt.Sprintf("%v", val), colorReset)
	}

	ops["console:at"] = func() {
		x := ValueStack.Pop("consoleX")
		y := ValueStack.Pop("consoleY")
		if !loop {
			return
		}
		// <ESC>[{ROW};{COLUMN}f
		fmt.Printf("\033[%d;%df", x, y)
	}

	ops["console:print"] = func() {
		str := ValueStack.Pop("strToPrint")
		y := ValueStack.Pop("consoleY")
		x := ValueStack.Pop("consoleX")
		if !loop {
			return
		}
		// <ESC>[{ROW};{COLUMN}f
		fmt.Printf("\033[%d;%df", x, y)
		fmt.Print(str)
	}

	ops["print"] = func() {
		str := ValueStack.Pop("strToPrint")
		if !loop {
			return
		}
		fmt.Print(str)
	}

	ops["format"] = func() {
		val := ValueStack.Pop("argList")
		str := ValueStack.Pop("formatString")

		var fmtString string
		switch str := str.(type) {
		case string:
			fmtString = str
		default:
			GfError("The first argument to the 'format' function must be a string")
		}

		var args []interface{}
		switch val := val.(type) {
		case []interface{}:
			args = val
		default:
			GfError("The second argument to the 'format' function must be a list of values")
		}

		result := fmt.Sprintf(fmtString, args...)

		ValueStack.Push(result)
	}

	ops["sleep"] = func() {
		duration := ValueStack.Pop("duration")
		if !loop {
			return
		}

		switch duration := duration.(type) {
		case int:
			time.Sleep(time.Millisecond * time.Duration(duration))
		default:
			GfError("The 'sleep' function requires an integer argument.")
		}
	}

	ops["getchar"] = func() {
		chr, _ := ReadChar()
		ValueStack.Push(chr)
	}

	ops["getline"] = func() {
		str, _ := ReadLn()
		ValueStack.Push(str)
	}

	//C datetime -> R1
	//C Put the current date/time object on the stack
	ops["datetime"] = func() {
		ValueStack.Push(time.Now())
	}

	//C S since -> R1
	//C The 'since' function takes a datetime object and calculates the elapsed time since the start time.
	ops["since"] = func() {
		val := ValueStack.Pop("startTime")
		if !loop {
			return
		}

		switch v := val.(type) {
		case time.Time:
			ValueStack.Push(time.Since(v))
		default:
			GfError("Invalid argument is type %t, should be time.Time", v)
		}
	}

	ops["@"] = func() {
		idxVal := ValueStack.Pop("indexVal")
		vect := ValueStack.Pop("vect")
		if !loop {
			return
		}

		idx := 0
		switch idxVal := idxVal.(type) {
		case int:
			idx = idxVal
		case float64:
			idx = int(idxVal)
		}

		switch x := vect.(type) {
		case []interface{}:
			if len(x) == 0 {
				ValueStack.Push(nil)
			} else if idx >= 0 {
				if idx < len(x) {
					ValueStack.Push(x[idx])
				} else {
					ValueStack.Push(x[len(x)-1])
				}
			} else {
				idx = len(x) + idx
				if idx < 0 {
					ValueStack.Push(x[0])
				} else {
					ValueStack.Push(x[idx])
				}
			}
		case map[interface{}]interface{}:
			ValueStack.Push(x[idxVal])
		case map[string]interface{}:
			idxStr := fmt.Sprintf("%v", idxVal)
			ValueStack.Push(x[idxStr])
		case string:
			result := ""
			if len(x) == 0 {
				result = ""
			} else if idx >= 0 {
				if idx <= len(x) {
					result = string(rune(x[idx]))
				} else {
					result = string(rune(x[len(x)-1]))
				}
			} else {
				idx = len(x) + idx
				if idx < 0 {
					result = string(rune(x[0]))
				} else {
					result = string(rune(x[idx]))
				}
			}
			ValueStack.Push(result)
		default:
			GfError("unable to index into a object of type %t using '%v'", x, idx)
		}
	}

	//C V I E !
	// The "!" stores the element E in the Vector V at index I
	ops["!"] = func() {
		newVal := ValueStack.Pop("newval")
		idx := ValueStack.Pop("index")
		vect := ValueStack.Pop("vector")
		if !loop {
			return
		}
		switch x := vect.(type) {
		case []interface{}:
			x[idx.(int)] = newVal
		case map[interface{}]interface{}:
			x[idx] = newVal
		case map[string]interface{}:
			x[idx.(string)] = newVal
		default:
			GfError("unable to index into a object of type %t using '%s'", x, idx)
		}
	}

	ops["first"] = func() {
		vect := ValueStack.Pop("list")
		if !loop {
			return
		}
		if vect == nil {
			ValueStack.Push(nil)
		}
		switch x := vect.(type) {
		case []interface{}:
			if len(x) > 0 {
				ValueStack.Push(x[0])
			} else {
				ValueStack.Push(nil)
			}
		case string:
			if len(x) > 0 {
				ValueStack.Push(string(x[:1]))
			} else {
				ValueStack.Push("")
			}
		default:
			ValueStack.Push(vect)
		}
	}

	ops["rest"] = func() {
		vect := ValueStack.Pop("list")
		if !loop {
			return
		}
		if vect == nil {
			ValueStack.Push(make([]interface{}, 0))
		}
		switch x := vect.(type) {
		case []interface{}:
			if len(x) > 0 {
				ValueStack.Push(x[1:])
			} else {
				ValueStack.Push(make([]interface{}, 0))
			}
		default:
			ValueStack.Push(make([]interface{}, 0))
		}
	}

	ops["skip"] = func() {
		num := ValueStack.Pop("numToSkip")
		vect := ValueStack.Pop("vector")
		if !loop {
			return
		}

		if vect == nil {
			ValueStack.Push(make([]interface{}, 0))
		}

		switch x := vect.(type) {
		case []interface{}:
			switch num := num.(type) {
			case int:
				if len(x) > num {
					ValueStack.Push(x[num:])
				} else {
					ValueStack.Push(make([]interface{}, 0))
				}
			default:
				GfError("the second argument to 'skip' must be an integer")
			}
		case string:
			switch num := num.(type) {
			case int:
				if len(x) > num {
					ValueStack.Push(string(x[num:]))
				} else {
					ValueStack.Push(make([]interface{}, 0))
				}
			default:
				GfError("the second argument to 'skip' must be an integer")
			}
		default:
			GfError("the first argument to 'skip' must be a vector")
		}
	}

	ops["lastn"] = func() {
		num := ValueStack.Pop("num")
		vect := ValueStack.Pop("vector")
		if !loop {
			return
		}

		if vect == nil {
			ValueStack.Push(make([]interface{}, 0))
		}

		switch x := vect.(type) {
		case []interface{}:
			switch num := num.(type) {
			case int:
				if len(x) > num {
					ValueStack.Push(x[len(x)-num:])
				} else {
					ValueStack.Push(x)
				}
			default:
				GfError("the second argument to 'lastn' must be an integer")
			}
		case string:
			switch num := num.(type) {
			case int:
				if len(x) > num {
					ValueStack.Push(string(x[len(x)-num:]))
				} else {
					ValueStack.Push(x)
				}
			default:
				GfError("the second argument to 'lastn' must be an integer")
			}
		default:
			GfError("the first argument to 'lastn' must be a vector")
		}
	}

	ops["last"] = func() {
		vect := ValueStack.Pop("vector")
		if !loop {
			return
		}

		switch x := vect.(type) {
		case []interface{}:
			ValueStack.Push(x[len(x)-1])
		case string:
			ValueStack.Push(string(x[len(x)-1]))
		default:
			GfError("the first argument to 'last' must be a vector")
		}
	}

	ops["nil?"] = func() {
		val := ValueStack.Pop("valueToTest")
		if !loop {
			return
		}
		ValueStack.Push(val == nil)
	}

	ops["nil"] = func() {
		ValueStack.Push(nil)
	}

	ops["empty?"] = func() {
		vect := ValueStack.Pop("valueToTest")
		if !loop {
			return
		}
		if vect == nil {
			ValueStack.Push(true)
			return
		}
		switch x := vect.(type) {
		case []interface{}:
			if len(x) > 0 {
				ValueStack.Push(false)
			} else {
				ValueStack.Push(true)
			}
		case string:
			if len(x) > 0 {
				ValueStack.Push(false)
			} else {
				ValueStack.Push(true)
			}
		default:
			ValueStack.Push(true)
		}
	}

	ops["notempty?"] = func() {
		vect := ValueStack.Pop("valueToTest")
		if !loop {
			return
		}
		if vect == nil {
			ValueStack.Push(false)
			return
		}
		switch x := vect.(type) {
		case []interface{}:
			if len(x) > 0 {
				ValueStack.Push(true)
			} else {
				ValueStack.Push(false)
			}
		case string:
			if len(x) > 0 {
				ValueStack.Push(true)
			} else {
				ValueStack.Push(false)
			}
		default:
			ValueStack.Push(false)
		}
	}

	ops["map"] = func() {
		progVal := ValueStack.Pop("program")
		vect := ValueStack.Pop("vector")
		if !loop {
			return
		}
		if vect == nil {
			ValueStack.Push(make([]interface{}, 0))
			return
		}

		if progVal == nil {
			ValueStack.Push(vect)
			return
		}

		var prog func()
		switch progVal := progVal.(type) {
		case op:
			prog = progVal.fn
		case func():
			prog = (progVal)
		default:
			GfError("invalid prog argument, please provide a lambda, not %t", progVal)
			return
		}

		switch vect := vect.(type) {
		case []interface{}:
			result := make([]interface{}, 0)
			for _, v := range vect {
				ValueStack.Push(v)
				prog()
				val := ValueStack.Pop("progResult")
				if !loop {
					break
				}
				if val != nil {
					result = append(result, val)
				}
			}
			ValueStack.Push(result)
		case map[interface{}]interface{}:
			result := make([]interface{}, 0, len(vect))
			for k, v := range vect {
				pair := make([]interface{}, 2)
				pair[0] = k
				pair[1] = v
				ValueStack.Push(pair)
				prog()
				val := ValueStack.Pop("progResult")
				if !loop {
					break
				}
				result = append(result, val)
			}
			ValueStack.Push(result)
		case map[string]interface{}:
			result := make([]interface{}, 0, len(vect))
			for k, v := range vect {
				pair := make([]interface{}, 2)
				pair[0] = k
				pair[1] = v
				ValueStack.Push(pair)
				prog()
				val := ValueStack.Pop("progResult")
				result = append(result, val)
				if !loop {
					break
				}
			}
			ValueStack.Push(result)
		default:
			ValueStack.Push(vect)
			prog()
		}
	}

	// Apply a function too each list element, returning nothing
	ops["each"] = func() {
		progVal := ValueStack.Pop("program")
		vect := ValueStack.Pop("vector")
		if !loop {
			return
		}
		if vect == nil {
			return
		}

		if progVal == nil {
			return
		}

		var prog func()
		switch progVal := progVal.(type) {
		case op:
			prog = progVal.fn
		case func():
			prog = (progVal)
		default:
			GfError("invalid prog argument, please provide a lambda, not %t", progVal)
			return
		}

		switch vect := vect.(type) {
		case []interface{}:
			for _, v := range vect {
				ValueStack.Push(v)
				prog()
				if !loop {
					break
				}
			}
		case map[interface{}]interface{}:
			for k, v := range vect {
				pair := make([]interface{}, 2)
				pair[0] = k
				pair[1] = v
				ValueStack.Push(pair)
				prog()
				if !loop {
					break
				}
			}
		case map[string]interface{}:
			for k, v := range vect {
				pair := make([]interface{}, 2)
				pair[0] = k
				pair[1] = v
				ValueStack.Push(pair)
				prog()
				if !loop {
					break
				}
			}
		default:
			ValueStack.Push(vect)
			prog()
		}
	}

	ops["filter"] = func() {
		progVal := ValueStack.Pop("program")
		vect := ValueStack.Pop("vector")
		if !loop {
			return
		}

		if vect == nil {
			ValueStack.Push(make([]interface{}, 0))
			return
		}

		var prog func()
		switch progVal := progVal.(type) {
		case op:
			prog = progVal.fn
		case func():
			prog = (progVal)
		default:
			GfError("invalid prog argument, please provide a lambda, not %t", progVal)
			return
		}

		switch vect := vect.(type) {
		case []interface{}:
			result := make([]interface{}, 0)
			for _, v := range vect {
				ValueStack.Push(v)
				prog()
				v2 := ValueStack.Pop("progResult")
				if !loop {
					return
				}
				if isTrue(v2) {
					result = append(result, v)
				}
			}
			ValueStack.Push(result)
		default:
			ValueStack.Push(vect)
			prog()
		}
	}

	ops["sort"] = (func() {
		values := ValueStack.Pop("listToSort")
		if !loop {
			return
		}
		switch values := values.(type) {
		case []interface{}:
			newlist := make([]interface{}, len(values))
			for i, v := range values {
				newlist[i] = v
			}
			sort.Slice(newlist, func(i int, j int) bool {
				return Compare(newlist[i], newlist[j]) == -1
			})
			ValueStack.Push(newlist)
		case map[interface{}]interface{}:
			newlist := make(PairList, len(values))
			index := 0
			for i, v := range values {
				newlist[index] = Pair{i, v}
				index++
			}
			sort.Sort(newlist)
			resultlist := make([]interface{}, 0, len(values))
			for _, v := range newlist {
				resultlist = append(resultlist, v)
			}
			ValueStack.Push(resultlist)
		case []string:
			newlist := make([]interface{}, len(values))
			for i, v := range values {
				newlist[i] = v
			}
			sort.Slice(newlist, func(i int, j int) bool {
				return Compare(newlist[i], newlist[j]) == -1
			})
			ValueStack.Push(newlist)
		case []int:
			newlist := make([]interface{}, len(values))
			for i, v := range values {
				newlist[i] = v
			}
			sort.Slice(newlist, func(i int, j int) bool {
				return Compare(newlist[i], newlist[j]) == -1
			})
			ValueStack.Push(newlist)
		default:
			GfError("this function can only sort []interface{} or []string")
		}
	})

	ops["dsort"] = (func() {
		values := ValueStack.Pop("listToSort")
		if !loop {
			return
		}

		switch values := values.(type) {
		case []interface{}:
			newlist := make([]interface{}, len(values))
			for i, v := range values {
				newlist[i] = v
			}
			sort.Slice(newlist, func(i int, j int) bool {
				return Compare(newlist[i], newlist[j]) == 1
			})
			ValueStack.Push(newlist)
		case map[interface{}]interface{}:
			newlist := make(DescendingPairList, len(values))
			index := 0
			for i, v := range values {
				newlist[index] = Pair{i, v}
				index++
			}
			sort.Sort(newlist)
			resultlist := make([]interface{}, 0, len(values))
			for _, v := range newlist {
				resultlist = append(resultlist, v)
			}
			ValueStack.Push(resultlist)
		case []string:
			newlist := make([]interface{}, len(values))
			for i, v := range values {
				newlist[i] = v
			}
			sort.Slice(newlist, func(i int, j int) bool {
				return Compare(newlist[i], newlist[j]) == 1
			})
			ValueStack.Push(newlist)
		case []int:
			newlist := make([]interface{}, len(values))
			for i, v := range values {
				newlist[i] = v
			}
			sort.Slice(newlist, func(i int, j int) bool {
				return Compare(newlist[i], newlist[j]) == 1
			})
			ValueStack.Push(newlist)
		default:
			GfError("this function can only sort []interface{} or []string")
		}
	})

	ops["dip"] = func() {
		prog := ValueStack.Pop("program")
		v1 := ValueStack.Pop("value")
		if !loop {
			return
		}
		switch prog := prog.(type) {
		case op:
			prog.fn()
			ValueStack.Push(v1)
		case func():
			prog()
			ValueStack.Push(v1)
		default:
			GfError("The first argument to 'dip' bust be a lambda, not %t.", prog)
		}
	}

	ops["primrec"] = func() {
		progVal := ValueStack.Pop("program")
		initProgVal := ValueStack.Pop("result")
		val := ValueStack.Pop("val")
		if !loop {
			return
		}

		var prog func()
		switch progVal := progVal.(type) {
		case op:
			prog = progVal.fn
		case func():
			prog = (progVal)
		default:
			GfError("invalid prog argument, please provide a lambda, not %t", progVal)
			return
		}

		var initProg func()
		switch initProgVal := initProgVal.(type) {
		case op:
			initProg = initProgVal.fn
		case func():
			initProg = initProgVal
		default:
			GfError("invalid init prog argument, please provide a lambda, not %t", progVal)
			return
		}

		initProg()
		result := ValueStack.Pop("initProgResult")
		if !loop {
			return
		}

		for isTrue(val) {
			switch tval := val.(type) {
			case int:
				ValueStack.Push(result)
				val = tval - 1
				ValueStack.Push(tval)
			case float64:
				ValueStack.Push(result)
				val = tval - 1
				ValueStack.Push(tval)
			case string:
				ValueStack.Push(result)
				val = string(tval[1:])
				ValueStack.Push(tval)
			case []interface{}:
				ValueStack.Push(result)
				val = tval[1:]
				ValueStack.Push(tval)
			default:
				GfError("can't use 'primrec' with type %t", val)
				return
			}

			prog()
			result = ValueStack.Pop("progResult")
			if !loop {
				return
			}
		}
		ValueStack.Push(result)
	}

	ops["linrec"] = (func() {
		endProgVal := ValueStack.Pop("endProg")
		rec1progVal := ValueStack.Pop("recProg")
		thenProgVal := ValueStack.Pop("thenProg")
		ifProgVal := ValueStack.Pop("ifProg")
		val := ValueStack.Pop("value")
		if !loop {
			return
		}

		var endProg func()
		switch progVal := endProgVal.(type) {
		case op:
			endProg = progVal.fn
		case func():
			endProg = progVal
		default:
			GfError("invalid end prog argument, please provide a lambda, not %t", progVal)
			return
		}

		var rec1prog func()
		switch progVal := rec1progVal.(type) {
		case op:
			rec1prog = progVal.fn
		case func():
			rec1prog = progVal
		default:
			GfError("invalid rec1 prog argument, please provide a lambda, not %t", progVal)
			return
		}

		var thenProg func()
		switch progVal := thenProgVal.(type) {
		case op:
			thenProg = progVal.fn
		case func():
			thenProg = progVal
		default:
			GfError("invalid then prog argument, please provide a lambda, not %t", progVal)
			return
		}

		var ifProg func()
		switch progVal := ifProgVal.(type) {
		case op:
			ifProg = progVal.fn
		case func():
			ifProg = progVal
		default:
			GfError("invalid if prog argument, please provide a lambda, not %t", progVal)
			return
		}

		count := 0
		for loop {
			ValueStack.Push(val)
			ifProg()
			r := ValueStack.Pop("ifProgResult")
			if !loop {
				return
			}
			if isTrue(r) {
				ValueStack.Push(val)
				thenProg()
				break
			}
			ValueStack.Push(val)
			rec1prog()
			val = ValueStack.Pop("value")
			count++
		}

		count--
		for loop && count >= 0 {
			endProg()
			count--
		}
	})

	ops["binrec"] = (func() {
		endProgVal := ValueStack.Pop("endProg")
		rec1progVal := ValueStack.Pop("recProg")
		thenProgVal := ValueStack.Pop("thenProg")
		ifProgVal := ValueStack.Pop("ifProg")
		val := ValueStack.Pop("value")
		if !loop {
			return
		}

		var endProg func()
		switch progVal := endProgVal.(type) {
		case op:
			endProg = progVal.fn
		case func():
			endProg = progVal
		default:
			GfError("invalid end prog argument, please provide a lambda, not %t", progVal)
			return
		}

		var rec1prog func()
		switch progVal := rec1progVal.(type) {
		case op:
			rec1prog = progVal.fn
		case func():
			rec1prog = progVal
		default:
			GfError("invalid rec1 prog argument, please provide a lambda, not %t", progVal)
			return
		}

		var thenProg func()
		switch progVal := thenProgVal.(type) {
		case op:
			thenProg = progVal.fn
		case func():
			thenProg = progVal
		default:
			GfError("invalid then prog argument, please provide a lambda, not %t", progVal)
			return
		}

		var ifProg func()
		switch progVal := ifProgVal.(type) {
		case op:
			ifProg = progVal.fn
		case func():
			ifProg = progVal
		default:
			GfError("invalid if prog argument, please provide a lambda, not %t", progVal)
			return
		}

		// DEFINE fib == {small} {drop 1} {1 - dup 1 -} {+} binrec;
		// 5
		// 	4 3
		// 4
		// 	3 2
		//    2 1
		//       3
		// 3
		//   2 1
		//		 3
		binRecHelper(val, ifProg, thenProg, rec1prog, endProg)
	})

	ops["float!"] = func() {
		val := ValueStack.Pop("valToConvert")
		if !loop {
			return
		}
		if val == nil {
			ValueStack.Push(0.0)
			return
		}

		switch val := val.(type) {
		case float64:
			ValueStack.Push(val)
		case int:
			ValueStack.Push(float64(val))
		case string:
			num, err := strconv.ParseFloat(val, 64)
			if err != nil {
				GfError("%v", err)
				return
			}
			ValueStack.Push(float64(num))
		default:
			GfError("can't convert '%v' of type %t to float.", val, val)
		}
	}

	ops["int!"] = func() {
		val := ValueStack.Pop("valToConvert")
		if !loop {
			return
		}
		if val == nil {
			ValueStack.Push(0)
			return
		}

		switch val := val.(type) {
		case float64:
			ValueStack.Push(int(val))
		case int:
			ValueStack.Push(val)
		case string:
			num, err := strconv.Atoi(val)
			if err != nil {
				GfError("%v", err)
				return
			}
			ValueStack.Push(float64(num))
		default:
			GfError("can't convert '%v' of type %t to float.", val, val)
		}
	}

	ops["chr!"] = (func() {
		val := ValueStack.Pop("valToConvert")
		if !loop {
			return
		}

		switch val := val.(type) {
		case int:
			ValueStack.Push(string(rune(val)))
		case float64:
			ValueStack.Push(string(rune(int(val))))
		case string:
			ValueStack.Push(string(rune(int(val[0]))))
		default:
			GfError("The argument to 'chr! must be an int, float64 or string; not %t, val")
		}
	})

	ops["str:join"] = (func() {
		val := ValueStack.Pop("listToJoin")
		if !loop {
			ValueStack.Push("")
			return
		}
		switch val := val.(type) {
		case []interface{}:
			// Turn a list into a string
			result := "" // Should be string builder
			for _, v := range val {
				switch v := v.(type) {
				case string:
					result += v
				case int:
					result += string(rune(v))
				case float32:
					result += string(rune(int(v)))
				case []interface{}:
					result += fmt.Sprintf("%v", v)
				}
			}
			ValueStack.Push(result)
		case string:
			ValueStack.Push(val)
		default:
			GfError("The argument 'str:join' must be a list; not [%t].", val)
		}
	})

	ops["str:tolower"] = func() {
		val := ValueStack.Pop("stringToLower")
		if !loop {
			return
		}

		switch val := val.(type) {
		case string:
			ValueStack.Push(strings.ToLower(val))
		default:
			ValueStack.Push(strings.ToLower(fmt.Sprint(val)))
		}
	}

	ops["str:toupper"] = func() {
		val := ValueStack.Pop("stringToUpper")
		if !loop {
			return
		}

		switch val := val.(type) {
		case string:
			ValueStack.Push(strings.ToUpper(val))
		default:
			ValueStack.Push(strings.ToUpper(fmt.Sprint(val)))
		}
	}

	ops["ord"] = (func() {
		val := ValueStack.Pop("stringToGetOrdOf")
		if !loop {
			return
		}
		switch val := val.(type) {
		case rune:
			ValueStack.Push((int(val)))
		case string:
			if len(val) > 0 {
				ValueStack.Push(int(val[0]))
			} else {
				ValueStack.Push(0)
			}
		default:
			GfError("Te 'ord' function can only be user on runes and strings; not %t", val)
		}
	})

	ops["string!"] = func() {
		val := ValueStack.Pop("valueToConvert")
		if !loop {
			return
		}
		if val == nil {
			ValueStack.Push("")
		}

		switch val := val.(type) {
		case []interface{}:
			// Turn a list into a string
			result := "" // Should be string builder
			for _, v := range val {
				result += stringify(v)
			}
			ValueStack.Push(result)
		default:
			ValueStack.Push(stringify(val))
		}
	}

	ops["explode"] = func() {
		val := ValueStack.Pop("stringToExplode")
		if !loop {
			return
		}
		result := make([]interface{}, 0)
		if val == nil {
			ValueStack.Push(result)
		}

		str := fmt.Sprintf("%s", val)
		for _, c := range str {
			result = append(result, string(c))
		}

		ValueStack.Push(result)
	}

	ops["reduce"] = func() {
		progVal := ValueStack.Pop("reduceProg")
		vect := ValueStack.Pop("listToReduce")
		if !loop {
			return
		}
		if vect == nil {
			ValueStack.Push(make([]interface{}, 0))
			return
		}

		if progVal == nil {
			GfError("the second argument must be a program, not %t", vect)
			return
		}

		var prog func()
		switch p := progVal.(type) {
		case op:
			prog = p.fn
		case func():
			prog = p
		default:
			GfError("the second argument must be a program, not %t", vect)
			return
		}

		switch vect := vect.(type) {
		case []interface{}:
			first := true
			for _, v := range vect {
				ValueStack.Push(v)
				if first {
					first = false
					continue
				}
				prog()
			}
		default:
			GfError("the first argument must be a list, not %t", vect)
		}
	}

	ops["while"] = func() {
		bodyExpr := ValueStack.Pop("bodyProgram")
		condExpr := ValueStack.Pop("condProgram")
		if !loop {
			return
		}

		if condExpr == nil {
			return
		}

		if bodyExpr == nil {
			return
		}

		switch condExpr := condExpr.(type) {
		case op:
			switch bodyExpr := bodyExpr.(type) {
			case op:
				for loop {
					condExpr.fn()
					if isFalse(ValueStack.Pop("condProgramResult")) {
						break
					}
					bodyExpr.fn()
				}
				return
			case func():
				for loop {
					condExpr.fn()
					if isFalse(ValueStack.Pop("condProgramResult")) {
						break
					}
					bodyExpr()
				}
				return
			default:
				GfError("the body of a while loop must be a lambda")
			}

		case func():
			switch bodyExpr := bodyExpr.(type) {
			case op:
				for loop {
					condExpr()
					if isFalse(ValueStack.Pop("condProgramResult")) {
						break
					}
					bodyExpr.fn()
				}
				return
			case func():
				for loop {
					condExpr()
					if isFalse(ValueStack.Pop("condProgramResult")) {
						break
					}
					bodyExpr()
				}
				return
			default:
				GfError("the body of a while loop must be a lambda")
			}

		default:
			GfError("the condition part of a while loop must be a lambda")
		}
	}

	ops["len"] = func() {
		val := ValueStack.Pop("valToGetLenOf")
		if !loop {
			return
		}
		switch val := val.(type) {
		case []interface{}:
			ValueStack.Push(len(val))
		case string:
			ValueStack.Push(len(val))
		case map[interface{}]interface{}:
			ValueStack.Push(len(val))
		case map[string]interface{}:
			ValueStack.Push(len(val))
		default:
			GfError("this operator cannot be applied to an object of type %t", val)
		}
	}

	ops["not?"] = func() {
		val := ValueStack.Pop("valToTest")
		if !loop {
			return
		}
		switch val := val.(type) {
		case bool:
			ValueStack.Push(!val)
		case string:
			ValueStack.Push(len(val) == 0)
		case map[interface{}]interface{}:
			ValueStack.Push(len(val) == 0)
		case map[string]interface{}:
			ValueStack.Push(len(val) == 0)
		case []interface{}:
			ValueStack.Push(len(val) == 0)
		default:
			ValueStack.Push(false)
		}
	}

	ops["load"] = func() {
		val := ValueStack.Pop("fileToLoad")
		if !loop {
			return
		}
		switch fileToRun := val.(type) {
		case string:
			CallStack.Push(activeFunction.tok)
			LoadFile(fileToRun)
			CallStack.Pop("exitLoad")
		default:
			GfError("requires a string argument, not '%s' [%t]", val, val)
		}
	}

	ops["eval"] = func() {
		val := ValueStack.Pop("strToEvaluate")
		if !loop {
			return
		}

		var text string
		switch val := val.(type) {
		case string:
			text = val
		default:
			text = fmt.Sprintf("%v", val)
		}

		lineno = 1
		fields := ParseLine(string(text))
		_, body := Compile(fields, 0, "")
		CallStack.Push(Token{
			Text:   activeFunction.tok.Text,
			File:   activeFunction.tok.File,
			Name:   activeFunction.tok.Name,
			Line:   activeFunction.tok.Line,
			Offset: activeFunction.tok.Offset})
		Eval(body)
		CallStack.Pop("evalExit")
		lineno = 1
	}

	ops["file:read"] = func() {
		val := ValueStack.Pop("fileToRead")
		if !loop {
			return
		}

		switch filename := val.(type) {
		case string:
			result, err := ioutil.ReadFile(filename)
			if err != nil {
				GfError("Error reading file: %s", err)
				return
			}
			ValueStack.Push(string(result))
		default:
			GfError("'file:read' requires a string argument, not '%s' [%t]", val, val)
		}
	}

	ops["file:readlines"] = (func() {
		val := ValueStack.Pop("fileToReadLinesFrom")
		if !loop {
			return
		}

		switch filename := val.(type) {
		case string:
			text, err := ioutil.ReadFile(filename)
			if err != nil {
				GfError("Error reading file: %s", err)
				return
			}
			result := make([]interface{}, 0)
			line := strings.Builder{}
			for _, c := range text {
				if c == '\r' {
					continue
				}

				if c == '\n' {
					result = append(result, line.String())
					line.Reset()
					continue
				}

				line.WriteByte(c)
			}

			// handle dangling line fragments
			if line.Len() > 0 {
				result = append(result, line.String())
			}
			ValueStack.Push(result)
		default:
			GfError("'file:read' requires a string argument, not '%s' [%t]", val, val)
		}
	})

	ops["file:readlinesWith"] = (func() {
		progVal := ValueStack.Pop("program")
		val := ValueStack.Pop("fileToReadLinesFrom")
		if !loop {
			return
		}

		var prog func()
		switch progVal := progVal.(type) {
		case op:
			prog = progVal.fn
		case func():
			prog = progVal
		default:
			GfError("the second argument to 'file:readlinesWith' must be a lambda, not %t", progVal)
			return
		}

		switch filename := val.(type) {
		case string:
			text, err := ioutil.ReadFile(filename)
			if err != nil {
				GfError("Error reading file: %s", err)
				return
			}
			result := make([]interface{}, 0)
			line := strings.Builder{}
			for _, c := range text {
				if c == '\r' {
					continue
				}

				if c == '\n' {
					strLine := line.String()
					ValueStack.Push(strLine)
					prog()
					progResult := ValueStack.Pop("progResult")
					if !loop {
						return
					}
					if progResult != nil {
						result = append(result, progResult)
					}
					line.Reset()
					continue
				}

				line.WriteByte(c)
			}

			// handle dangling line fragments
			if line.Len() > 0 {
				strLine := line.String()
				ValueStack.Push(strLine)
				prog()
				progResult := ValueStack.Pop("progResult")
				if !loop {
					return
				}
				if progResult != nil {
					result = append(result, progResult)
				}
			}
			ValueStack.Push(result)
		default:
			GfError("'file:read' requires a string argument, not '%s' [%t]", val, val)
		}
	})

	ops["file:files"] = (func() {
		items, err := os.ReadDir(".")
		if err != nil {
			GfError("error getting directory entries: %s", err)
			return
		}
		result := make([]interface{}, 0)
		for _, val := range items {
			if !val.IsDir() {
				name := val.Name()
				result = append(result, name)
			}
		}
		ValueStack.Push(result)
	})

	ops["file:files/2"] = (func() {
		pat := ValueStack.Pop("filePattern")
		if !loop {
			return
		}

		strpat := fmt.Sprintf("%v", pat)
		items, err := os.ReadDir(strpat)
		if err != nil {
			GfError("error getting directory entries: %s", err)
			return
		}
		result := make([]interface{}, 0)
		for _, val := range items {
			if !val.IsDir() {
				name := val.Name()
				result = append(result, name)
			}
		}
		ValueStack.Push(result)
	})

	ops["file:dirs"] = (func() {
		items, err := os.ReadDir(".")
		if err != nil {
			GfError("error getting directory entries: %s", err)
			return
		}
		result := make([]interface{}, 0)
		for _, val := range items {
			if val.IsDir() {
				name := val.Name()
				result = append(result, name)
			}
		}
		ValueStack.Push(result)
	})

	ops["file:dirs/2"] = (func() {
		pat := ValueStack.Pop("dirPattern")
		if !loop {
			return
		}

		strpat := fmt.Sprintf("%v", pat)
		items, err := os.ReadDir(strpat)
		if err != nil {
			GfError("error getting directory entries: %s", err)
			return
		}
		result := make([]interface{}, 0)
		for _, val := range items {
			if val.IsDir() {
				name := val.Name()
				result = append(result, name)
			}
		}
		ValueStack.Push(result)
	})

	ops["str:split"] = func() {
		sep := ValueStack.Pop("separatorRegex")
		val := ValueStack.Pop("stringsToSplit")
		if !loop {
			return
		}

		if val == nil {
			ValueStack.Push(make([]interface{}, 0, 1000))
			return
		}

		switch val := val.(type) {
		case string:
			switch sep := sep.(type) {
			case *regexp.Regexp:
				pieces := sep.Split(val, -1)
				result := make([]interface{}, 0, len(pieces))
				for _, str := range pieces {
					result = append(result, str)
				}
				ValueStack.Push(result)
			case string:
				pieces := strings.Split(val, sep)
				result := make([]interface{}, 0, len(pieces))
				for _, str := range pieces {
					result = append(result, str)
				}
				ValueStack.Push(result)

			default:
				sepStr := fmt.Sprintf("%s", sep)
				pieces := strings.Split(fmt.Sprintf("%s", val), sepStr)
				result := make([]interface{}, 0, len(pieces))
				for _, str := range pieces {
					result = append(result, str)
				}
				ValueStack.Push(result)
			}
		case []interface{}:
			for _, v := range val {
				val := fmt.Sprintf("%s", v)
				switch sep := sep.(type) {
				case *regexp.Regexp:
					pieces := sep.Split(val, -1)
					result := make([]interface{}, 0, len(pieces))
					for _, str := range pieces {
						result = append(result, str)
					}
					ValueStack.Push(result)
				case string:
					pieces := strings.Split(val, sep)
					result := make([]interface{}, 0, len(pieces))
					for _, str := range pieces {
						result = append(result, str)
					}
					ValueStack.Push(result)
				default:
					sepStr := fmt.Sprintf("%s", sep)
					pieces := strings.Split(fmt.Sprintf("%s", val), sepStr)
					result := make([]interface{}, 0, len(pieces))
					for _, str := range pieces {
						result = append(result, str)
					}
					ValueStack.Push(result)
				}
			}

		default:
			GfError("the argument to this function must be a string, not %t", val)
			return
		}

	}

	ops["take"] = (func() {
		countVal := ValueStack.Pop("numToTake")
		val := ValueStack.Pop("vector")
		if !loop {
			return
		}

		result := make([]interface{}, 0)
		if val == nil {
			ValueStack.Push(result)
			return
		}

		count := 0
		switch countVal := countVal.(type) {
		case int:
			count = countVal
		case float64:
			count = int(countVal)
		default:
			GfError("The count argument to 'take' must be an integer, not %t", countVal)
			return
		}

		switch val := val.(type) {
		case []interface{}:
			if len(val) == 0 {
				ValueStack.Push(result)
				return
			}
			if count >= 0 {
				for _, v := range val {
					if count > 0 {
						result = append(result, v)
					} else {
						break
					}
					count--
				}
			} else {
				for count < 0 {
					result = append(result, val[len(val)+count])
					count++
				}
			}
		default:
			if count != 0 {
				result = append(result, val)
			}
		}
		ValueStack.Push(result)
	})

	ops["str:match"] = (func() {
		pat := ValueStack.Pop("regexToMatch")
		val := ValueStack.Pop("stringsToMatch")
		if !loop {
			return
		}

		switch val := val.(type) {
		case string:
			switch pat := pat.(type) {
			case *regexp.Regexp:
				wasMatch := pat.MatchString(val)
				ValueStack.Push(wasMatch)
			case string:
				ValueStack.Push(pat == val)
			default:
				patStr := fmt.Sprintf("%s", pat)
				ValueStack.Push(patStr == val)
			}
		case []interface{}:
			result := make([]interface{}, 0, len(val))
			for _, v := range val {
				var vstr string
				switch v := v.(type) {
				case string:
					vstr = v
				default:
					vstr = fmt.Sprintf("%v", v)
				}
				switch pat := pat.(type) {
				case *regexp.Regexp:
					wasMatch := pat.MatchString(vstr)
					if wasMatch {
						result = append(result, vstr)
					}
				case string:
					if pat == vstr {
						result = append(result, vstr)
					}
				default:
					patStr := fmt.Sprintf("%v", pat)
					if patStr == vstr {
						result = append(result, vstr)
					}
				}
			}
			ValueStack.Push(result)
			return
		default:
			GfError("the first argument to this function must be a string or list.")

		}
	})

	ops["str:notmatch"] = (func() {
		pat := ValueStack.Pop("regexToMatch")
		val := ValueStack.Pop("stringsToMatch")
		if !loop {
			return
		}

		switch val := val.(type) {
		case string:
			switch pat := pat.(type) {
			case *regexp.Regexp:
				wasMatch := pat.MatchString(val)
				ValueStack.Push(!wasMatch)
			case string:
				ValueStack.Push(pat != val)
			default:
				patStr := fmt.Sprintf("%s", pat)
				ValueStack.Push(patStr != val)
			}
		case []interface{}:
			result := make([]interface{}, 0, len(val))
			for _, v := range val {
				var vstr string
				switch v := v.(type) {
				case string:
					vstr = v
				default:
					vstr = fmt.Sprintf("%v", v)
				}
				switch pat := pat.(type) {
				case *regexp.Regexp:
					wasMatch := pat.MatchString(vstr)
					if !wasMatch {
						result = append(result, vstr)
					}
				case string:
					if pat != vstr {
						result = append(result, vstr)
					}
				default:
					patStr := fmt.Sprintf("%v", pat)
					if patStr != vstr {
						result = append(result, vstr)
					}
				}
			}
			ValueStack.Push(result)
			return
		default:
			GfError("the first argument to this function must be a string or list.")
		}
	})

	ops["str:replace"] = func() {
		repval := ValueStack.Pop("replacementValue")
		pat := ValueStack.Pop("regexPattern")
		val := ValueStack.Pop("stringsToReplace")
		if !loop {
			return
		}

		if val == nil {
			val = ""
		}

		if repval == nil {
			repval = ""
		}

		repstr := fmt.Sprintf("%s", repval)

		switch val := val.(type) {
		case string:
			switch pat := pat.(type) {
			case *regexp.Regexp:
				newStr := string(pat.ReplaceAllString(val, repstr))
				ValueStack.Push(newStr)
			case string:
				ValueStack.Push(strings.ReplaceAll(val, pat, repstr))
			default:
				patStr := fmt.Sprintf("%s", pat)
				ValueStack.Push(strings.ReplaceAll(val, patStr, repstr))
			}
		case []interface{}:
			result := make([]interface{}, 0, len(val))
			for _, v := range val {
				vstr := fmt.Sprintf("%s", v)
				switch pat := pat.(type) {
				case *regexp.Regexp:
					newStr := string(pat.ReplaceAllString(vstr, repstr))
					result = append(result, newStr)
				case string:
					newStr := strings.ReplaceAll(vstr, pat, repstr)
					result = append(result, newStr)
				default:
					patStr := fmt.Sprintf("%s", pat)
					newStr := strings.ReplaceAll(vstr, patStr, repstr)
					result = append(result, newStr)
				}
			}
			ValueStack.Push(result)
		default:
			GfError("the argument to this function must be a string or list")
		}
	}

	ops["set!"] = func() {
		val := ValueStack.Pop("vector")
		if !loop {
			return
		}

		var result map[interface{}]interface{}
		switch val := val.(type) {
		case []interface{}:
			result = make(map[interface{}]interface{}, len(val))
			for _, v := range val {
				result[v] = true
			}
		default:
			result = make(map[interface{}]interface{}, 1)
			result[val] = true
		}
		ValueStack.Push(result)
	}

	// Turn the list into a counted set
	ops["cset!"] = func() {
		val := ValueStack.Pop("vector")
		if !loop {
			return
		}

		var result map[interface{}]interface{}
		switch val := val.(type) {
		case []interface{}:
			result = make(map[interface{}]interface{}, len(val))
			for _, v := range val {
				cv, ok := result[v]
				if ok {
					result[v] = cv.(int) + 1
				} else {
					result[v] = 1
				}
			}
		default:
			result = make(map[interface{}]interface{}, 1)
			result[val] = 1
		}
		ValueStack.Push(result)
	}

	ops["dict!"] = func() {
		val := ValueStack.Pop("vector")
		if !loop {
			return
		}

		var result map[interface{}]interface{}
		switch val := val.(type) {
		case []interface{}:
			if len(val)%2 != 0 {
				GfError("when converting a list to a dictionary, the list length must be even")
				return
			}
			result = make(map[interface{}]interface{}, len(val))
			iskey := true
			var key interface{}
			for _, v := range val {
				if iskey {
					key = v
				} else {
					result[key] = v
				}
				iskey = !iskey
			}
			ValueStack.Push(result)
		default:
			GfError("only a list can be converted to a dictionary")
		}
	}

	ops["regex!"] = func() {
		val := ValueStack.Pop("valueToConvert")
		if !loop {
			return
		}

		switch val := val.(type) {
		case string:
			re, err := regexp.Compile(val)
			if err != nil {
				GfError("error compiling regex /%s/: %s", val, err)
			}
			ValueStack.Push(re)
		default:
			valStr := fmt.Sprintf("%s", val)
			re, err := regexp.Compile(valStr)
			if err != nil {
				GfError("error compiling regex /%s/: %s", valStr, err)
			}
			ValueStack.Push(re)
		}
	}

	ops["ops"] = func() {
		ValueStack.Push(ops)
	}

	ops["vars"] = func() {
		ValueStack.Push(VariableTable)
	}

	ops["shell"] = func() {
		cmdToRun := ValueStack.Pop("cmdToRun")
		if !loop {
			return
		}

		switch cmdToRun := cmdToRun.(type) {
		case string:
			cmd := exec.Command(cmdToRun)
			data, err := cmd.Output()
			if err != nil {
				GfError("error running command '%s': %s", cmd, err)
			}
			ValueStack.Push(string(data))
		case []interface{}:
			argVector := make([]string, 0)
			var cmdName string = ""
			for _, element := range cmdToRun {
				if cmdName == "" {
					cmdName = fmt.Sprintf("%v", element)
				} else {
					argVector = append(argVector, fmt.Sprintf("%v", element))
				}
			}
			cmd := exec.Command(cmdName, argVector...)
			data, err := cmd.Output()
			if err != nil {
				GfError("error running command '%s': %s", cmdName, err)
			}
			ValueStack.Push(string(data))
		default:
			GfError("this command requires either the name of a command to run or a command vector")
		}
	}

	LoadFile("prelude.gf")

	if len(os.Args) == 1 {
		fmt.Println(colorGreen+"Welcome to Go Forth | pid: ", os.Getpid())

		/* Need to extract this and turn it into real tests.

		fmt.Println("minus one  test")
		_, tbody := Compile(ParseLine(`
			"script.gf" file:read r/\n/ str:split r/^def/ str:match .
		`), 0, "")
		Eval(tbody)

		fmt.Println("zero test")
		_, tbody = Compile(ParseLine(`
			"2" [
					1 {drop "one" .red "one"}
					r/2/ {drop "two" .blue "two"}
					3 {drop "three" .green "three"}
				] case .
		`), 0, "")
		Eval(tbody)

		fmt.Println("first test")
		_, tbody = Compile(ParseLine(`[1 2 3 4 5] {+} reduce .green`), 0, "")
		Eval(tbody)

		fmt.Println("secord test")
		_, tbody = Compile(ParseLine("[1 2] [3 4 5] + .green"), 0, "")
		Eval(tbody)

		fmt.Println("third test")
		_, tbody = Compile(ParseLine(" 2 3 4 + * typeof .red"), 0, "")
		Eval(tbody)

		fmt.Println("fourth test")
		_, tbody = Compile(ParseLine("\"script.gf\" load"), 0, "")
		Eval(tbody)

		fmt.Println("fifth test")
		_, tbody = Compile(ParseLine("5 1 {*} primrec ."), 0, "")
		Eval(tbody)

		fmt.Println("sixth test")
		_, tbody = Compile(ParseLine("[1 2 3 4 5] [] {first swap append} primrec ."), 0, "")
		Eval(tbody)

		fmt.Println("seventh test")
		_, tbody = Compile(ParseLine("\"abcdefghi\" \"\" {first swap +} primrec ."), 0, "")
		Eval(tbody)

		fmt.Println("eighth test")
		_, tbody = Compile(ParseLine("10 1 {*} primrec ."), 0, "")
		Eval(tbody)

		*/

		ct := time.Now()
		for !quit {
			loop = true

			fmt.Printf(colorGreen+"\nTime: %s Stack Depth: %d\n", time.Since(ct), ValueStack.index)
			fmt.Print("|> " + colorReset)

			CallStack.Reset()
			line, err := ReadLn()
			if err != nil {
				fmt.Println(err)
			} else {
				if strings.TrimSpace(line) == "quit" {
					break
				}
				fields := ParseLine(line)
				_, body := Compile(fields, 0, "")
				ct = time.Now()
				Eval(body)
			}
		}
	} else {
		fileToRun := os.Args[1]
		LoadFile(fileToRun)
	}
}
