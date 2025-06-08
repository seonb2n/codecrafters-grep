package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
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
	} else if isStartWithAnchor(pattern) {
		return matchStartWith(line, pattern)
	} else if isEndWithAnchor(pattern) {
		return matchEndWith(line, pattern)
	} else if isSimpleLiteral(pattern) {
		return matchLiteralCharacter(line, pattern)
	}
	return matchPattern(string(line), pattern), nil
}

func isSimpleLiteral(pattern string) bool {
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '\\' || pattern[i] == '[' || pattern[i] == '+' ||
			pattern[i] == '?' || pattern[i] == '.' || pattern[i] == '|' ||
			pattern[i] == '(' || pattern[i] == ')' {
			return false
		}
	}
	return true
}

func isStartWithAnchor(pattern string) bool {
	return pattern[0] == '^'
}

func isEndWithAnchor(pattern string) bool {
	return pattern[len(pattern)-1] == '$'
}

func matchPattern(text string, pattern string) bool {
	for i := 0; i < len(text); i++ {
		if matchHere(text[i:], pattern) {
			return true
		}
	}
	return false
}

func matchPlus(text string, char byte, remainingPattern string) bool {
	// 최소 1번은 매치되어야 함
	if len(text) == 0 || !charMatches(text[0], char) {
		return false
	}

	// 1번 매치된 후, 가능한 많이 매치 시도
	i := 1
	for i < len(text) && charMatches(text[i], char) {
		i++
	}

	// 매치된 개수만큼 역순으로 시도 (greedy matching)
	for j := i; j >= 1; j-- {
		if matchHere(text[j:], remainingPattern) {
			return true
		}
	}

	return false
}

func matchQuestion(text string, char byte, remainPattern string) bool {
	// 0번 매치
	if matchHere(text, remainPattern) {
		return true
	}
	// 1번 매치
	if len(text) > 0 && charMatches(text[0], char) {
		return matchHere(text[1:], remainPattern)
	}
	return false
}

func matchOrPattern(text string, pattern string) bool {
	// 첫 번째 괄호의 끝 찾기
	parenEnd := findMatchingParen(pattern, 0)
	if parenEnd == -1 {
		return false
	}

	orPart := pattern[0 : parenEnd+1]
	suffix := pattern[parenEnd+1:]

	// OR 대안들 시도
	innerPattern := orPart[1:parenEnd]
	alternatives := strings.Split(innerPattern, "|")

	for _, alt := range alternatives {
		if matchHere(text, alt+suffix) {
			return true
		}
	}

	return false
}

func charMatches(textChar, patternChar byte) bool {
	if patternChar == '.' {
		return true
	}
	return textChar == patternChar
}

// 현재 위치에서 패턴 매치 여부 확인
func matchHere(text string, pattern string) bool {
	if len(pattern) == 0 {
		return true
	}

	if len(text) == 0 {
		return false
	}

	// or 패턴 처리
	if isOrPattern(pattern) {
		return matchOrPattern(text, pattern)
	}

	// + 패턴 처리 (두 번째 문자가 +인 경우)
	if len(pattern) >= 2 && pattern[1] == '+' {
		return matchPlus(text, pattern[0], pattern[2:])
	}

	// ? 패턴 처리 (두 번째 문자가 ?인 경우)
	if len(pattern) >= 2 && pattern[1] == '?' {
		return matchQuestion(text, pattern[0], pattern[2:])
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

func findMatchingParen(pattern string, start int) int {
	if start >= len(pattern) || pattern[start] != '(' {
		return -1
	}

	count := 1
	for i := start + 1; i < len(pattern); i++ {
		if pattern[i] == '(' {
			count++
		} else if pattern[i] == ')' {
			count--
			if count == 0 {
				return i
			}
		}
	}
	return -1
}

func isOrPattern(pattern string) bool {
	if len(pattern) < 5 { // 최소 "(a|b)" 형태
		return false
	}
	if pattern[0] != '(' {
		return false
	}

	// 매칭되는 닫는 괄호 찾기
	parenEnd := findMatchingParen(pattern, 0)
	if parenEnd == -1 {
		return false
	}

	// 첫 번째 괄호 안에 | 가 있는지 확인
	return strings.Contains(pattern[1:parenEnd], "|")
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

func matchStartWith(line []byte, pattern string) (bool, error) {
	return bytes.HasPrefix(line, []byte(pattern[1:])), nil
}

func matchEndWith(line []byte, pattern string) (bool, error) {
	return bytes.HasSuffix(line, []byte(pattern[0:len(pattern)-1])), nil
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
