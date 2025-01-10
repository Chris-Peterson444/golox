package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

var hadError bool = false

type TokenType int

const (
	// Single-character tokens.
	LEFT_PAREN TokenType = iota
	RIGHT_PAREN
	LEFT_BRACE
	RIGHT_BRACE
	COMMA
	DOT
	MINUS
	PLUS
	SEMICOLON
	SLASH
	STAR

	// One or two character tokens.
	BANG
	BANG_EQUAL
	EQUAL
	EQUAL_EQUAL
	GREATER
	GREATER_EQUAL
	LESS
	LESS_EQUAL

	// Literals.
	IDENTIFIER
	STRING
	NUMBER

	// Keywords.
	AND
	CLASS
	ELSE
	FALSE
	FUN
	FOR
	IF
	NIL
	OR
	PRINT
	RETURN
	SUPER
	THIS
	TRUE
	VAR
	WHILE

	// EOF token
	EOF
)

var keywords = map[string]TokenType{
	"and":    AND,
	"class":  CLASS,
	"else":   ELSE,
	"false":  FALSE,
	"for":    FOR,
	"fun":    FUN,
	"if":     IF,
	"nil":    NIL,
	"or":     OR,
	"print":  PRINT,
	"return": RETURN,
	"super":  SUPER,
	"this":   THIS,
	"true":   TRUE,
	"var":    VAR,
	"while":  WHILE,
}

type LoxLiteral interface {
	RawPrint() string
}

type LoxString struct {
	value string
}

func (s LoxString) RawPrint() string {
	return fmt.Sprintf("%q", s.value)
}

type LoxNumber struct {
	value float64
}

func (n LoxNumber) RawPrint() string {
	return fmt.Sprintf("%v", n.value)
}

type LoxEmptyLiteral struct{}

func (e LoxEmptyLiteral) RawPrint() string {
	return "null"
}

type Token struct {
	_type   TokenType
	lexeme  string
	literal LoxLiteral
	line    int
}

func (tok *Token) String() string {
	return fmt.Sprintf("%s %s %s", tok._type, tok.lexeme, tok.literal.RawPrint())
}

type Scanner struct {
	source  string
	tokens  []Token
	start   int
	current int
	line    int
}

func NewScanner(source string) Scanner {
	return Scanner{
		source: source,
		line:   1,
	}
}

func (scan *Scanner) isAtEnd() bool {
	return scan.current >= len(scan.source)
}

func (scan *Scanner) ScanTokens() []Token {
	for !scan.isAtEnd() {
		scan.start = scan.current
		scan.scanToken()
	}
	tok := Token{
		_type:   EOF,
		lexeme:  "",
		literal: LoxEmptyLiteral{},
		line:    scan.line,
	}
	scan.tokens = append(scan.tokens, tok)
	return scan.tokens
}

func (scan *Scanner) scanToken() {
	var char byte = scan.advance()
	switch char {
	case '(':
		scan.addToken(LEFT_PAREN)
	case ')':
		scan.addToken(RIGHT_PAREN)
	case '{':
		scan.addToken(LEFT_BRACE)
	case '}':
		scan.addToken(RIGHT_BRACE)
	case ',':
		scan.addToken(COMMA)
	case '.':
		scan.addToken(DOT)
	case '-':
		scan.addToken(MINUS)
	case '+':
		scan.addToken(PLUS)
	case ';':
		scan.addToken(SEMICOLON)
	case '*':
		scan.addToken(STAR)
	case '!':
		if scan.match('=') {
			scan.addToken(BANG_EQUAL)
		} else {
			scan.addToken(BANG)
		}
	case '=':
		if scan.match('=') {
			scan.addToken(EQUAL_EQUAL)
		} else {
			scan.addToken(EQUAL)
		}
	case '<':
		if scan.match('=') {
			scan.addToken(LESS_EQUAL)
		} else {
			scan.addToken(LESS)
		}
	case '>':
		if scan.match('=') {
			scan.addToken(GREATER_EQUAL)
		} else {
			scan.addToken(GREATER)
		}
	case '/':
		if scan.match('/') {
			for scan.peek() != '\n' && !scan.isAtEnd() {
				scan.advance()
			}
		} else {
			scan.addToken(SLASH)
		}
	case ' ', '\r', '\t':
		// Do nothing, skip
	case '\n':
		scan.line++
	case '"':
		scan.parseString()
	default:
		if scan.isDigit(char) {
			scan.number()
		} else if scan.isAlpha(char) {
			scan.identifier()
		} else {
			message := fmt.Sprintf("Unexpected character: %c", char)
			LoxError(scan.line, message)
		}
	}
}

func (scan *Scanner) advance() byte {
	ret := scan.source[scan.current]
	scan.current++
	return ret

}

func (scan *Scanner) peek() byte {
	if scan.isAtEnd() {
		return '\000' // Rune literals are three-digit octals
	}

	return scan.source[scan.current]
}

func (scan *Scanner) peekNext() byte {
	if scan.current+1 > len(scan.source) {
		return '\000'
	}
	return scan.source[scan.current+1]
}

func (scan *Scanner) match(expected byte) bool {
	if scan.isAtEnd() {
		return false
	}

	if scan.source[scan.current] != expected {
		return false
	}

	scan.current++

	return true
}

func (scan *Scanner) number() {
	for scan.isDigit(scan.peek()) {
		scan.advance()
	}

	if scan.peek() == '.' && scan.isDigit(scan.peekNext()) {
		// Consume the "."
		scan.advance()

		for scan.isDigit(scan.peek()) {
			scan.advance()
		}
	}
	number, err := strconv.ParseFloat(scan.source[scan.start:scan.current], 64)
	if err != nil {
		//
	}
	numberLiteral := LoxNumber{value: number}
	scan.addTokenAndLiteral(NUMBER, numberLiteral)

}

func (scan *Scanner) identifier() {
	for scan.isAlphaNumeric(scan.peek()) {
		scan.advance()
	}

	text := scan.source[scan.start:scan.current]
	_type, ok := keywords[text]
	if !ok {
		_type = IDENTIFIER
	}
	scan.addToken(_type)
}

func (scan *Scanner) isDigit(char byte) bool {
	return char >= '0' && char <= '9'
}

func (scan *Scanner) isAlpha(char byte) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		char == '_'
}

func (scan *Scanner) isAlphaNumeric(char byte) bool {
	return scan.isAlpha(char) || scan.isDigit(char)
}

func (scan *Scanner) parseString() {
	for scan.peek() != '"' && !scan.isAtEnd() {
		if scan.peek() == '\n' {
			scan.line++
		}
		scan.advance()
	}

	if scan.isAtEnd() {
		LoxError(scan.line, "Unterminated string.")
		return
	}

	// The closing "
	scan.advance()

	// Trim the surrounding quotes
	literal := LoxString{
		value: scan.source[scan.start+1 : scan.current-1],
	}
	scan.addTokenAndLiteral(STRING, literal)
}

func (scan *Scanner) addToken(_type TokenType) {
	scan.addTokenAndLiteral(_type, LoxEmptyLiteral{})
}

func (scan *Scanner) addTokenAndLiteral(_type TokenType, literal LoxLiteral) {
	text := scan.source[scan.start:scan.current]
	tok := Token{
		_type:   _type,
		lexeme:  text,
		literal: literal,
		line:    scan.line,
	}
	scan.tokens = append(scan.tokens, tok)
}

func Run(source string) {
	// fmt.Printf("reading source '%s'\n", source)
	scanner := NewScanner(source)
	tokens := scanner.ScanTokens()

	// For now, just print the Tokens
	for _, tok := range tokens {
		fmt.Printf("%s\n", &tok)
	}
}

func RunFile(path string) {
	fileContents, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}
	Run(string(fileContents))

}

func RunPrompt() {
	reader := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for reader.Scan() {
		line := reader.Text()
		Run(line)
		fmt.Print("> ")
	}
	fmt.Print("\nExit\n")

}

func LoxError(line int, message string) {
	LoxReport(line, "", message)
}

func LoxReport(line int, where string, message string) {
	fmt.Fprintf(os.Stderr, "[line %d] Error%s: %s\n", line, where, message)
	hadError = true
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")
	// RunPrompt()

	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: ./your_program.sh tokenize <filename>")
		os.Exit(1)
	}

	command := os.Args[1]

	if command != "tokenize" {
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}

	filename := os.Args[2]
	RunFile(filename)
	if hadError {
		os.Exit(65)
	}
	// fileContents, err := os.ReadFile(filename)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
	// 	os.Exit(1)
	// }

	// if len(fileContents) > 0 {
	// 	panic("Scanner not implemented")
	// } else {
	// 	fmt.Println("EOF  null") // Placeholder, remove this line when implementing the scanner
	// }
}
