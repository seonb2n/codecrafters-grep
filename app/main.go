package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

var _ = bytes.ContainsAny

// Usage: echo <input_text> | your_program.sh -E <pattern>
func main() {
	if len(os.Args) < 3 || os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
		os.Exit(2) // 1 means no lines were selected, >1 means error
	}

	pattern := os.Args[2]

	line, err := io.ReadAll(os.Stdin) // assume we're only dealing with a single line
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
		os.Exit(2)
	}

	ok, err := matchLine(line, pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	if !ok {
		os.Exit(1)
	}

}

func matchLine(line []byte, pattern string) (bool, error) {
	if pattern == "\\d" {
		return matchDigit(line)
	} else if len(pattern) > 2 && pattern[0] == '[' && pattern[len(pattern)-1] == ']' {
		if len(pattern) > 3 && pattern[1] == '^' {
			return matchOnlyLiteralCharacter(line, pattern[2:len(pattern)-1])
		} else {
			return matchLiteralCharacter(line, pattern[1:len(pattern)-1])
		}
	} else if isSimpleLiteral(pattern) {
		return matchLiteralCharacter(line, pattern)
	}
	return matchPattern(string(line), pattern), nil
}

// 단순 리터럴 패턴 여부 확인
func isSimpleLiteral(pattern string) bool {
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '\\' || pattern[i] == '[' {
			return false
		}
	}
	return true
}

func matchPattern(text string, pattern string) bool {
	for i := 0; i < len(text); i++ {
		if matchHere(text[i:], pattern) {
			return true
		}
	}
	return false
}

// 현재 위치에서 패턴 매치 여부 확인
func matchHere(text string, pattern string) bool {
	if len(pattern) == 0 {
		return true
	}

	if len(text) == 0 {
		return false
	}

	// \d  패턴 처리
	if len(pattern) >= 2 && pattern[0] == '\\' && pattern[1] == 'd' {
		if isDigit(text[0]) {
			return matchHere(text[1:], pattern[2:])
		}
		return false
	}

	// \w 패턴 처리
	if len(pattern) >= 2 && pattern[0] == '\\' && pattern[1] == 'w' {
		if isWordChar(text[0]) {
			return matchHere(text[1:], pattern[2:])
		}
		return false
	}

	// .패턴 처리
	if pattern[0] == '.' {
		return matchHere(text[1:], pattern[1:])
	}

	if pattern[0] == text[0] {
		return matchHere(text[1:], pattern[1:])
	}

	return false
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '_'
}

func matchLiteralCharacter(line []byte, pattern string) (bool, error) {
	return bytes.ContainsAny(line, pattern), nil
}

func matchOnlyLiteralCharacter(line []byte, pattern string) (bool, error) {
	for _, b := range line {
		if !bytes.ContainsAny([]byte(pattern), string(b)) {
			// 패턴에 없는 문자를 찾음
			return true, nil
		}
	}
	return false, nil
}

func matchDigit(line []byte) (bool, error) {
	return bytes.ContainsAny(line, "0123456789"), nil
}
