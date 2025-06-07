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
	var ok bool

	if pattern == "\\d" {
		ok, _ = matchDigit(line)
	} else if pattern[0] == '[' && pattern[len(pattern)-1] == ']' {
		if pattern[1] == '^' {
			pattern = pattern[2 : len(pattern)-1]
			ok, _ = matchOnlyLiteralCharacter(line, pattern)
		} else {
			pattern = pattern[1 : len(pattern)-1]
			ok, _ = matchLiteralCharacter(line, pattern)
		}
	} else {
		ok, _ = matchLiteralCharacter(line, pattern)
	}

	return ok, nil
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
