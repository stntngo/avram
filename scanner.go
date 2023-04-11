package avram

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"unicode/utf8"
)

const eof = -1

// NewScanner constructs a new avram Scanner from the
// provided input string.
func NewScanner(input string) *Scanner {
	return &Scanner{
		input: input,
	}
}

// Scanner is responsible for maintaining the iterative state through
// which the constructed parser moves.
type Scanner struct {
	input string // the string being lexed
	start int    // location of the end of the last emitted token
	pos   int    // current position of the lexer in the input
	width []int  // width history of read but un-emitted runes from the input
	line  int    // current line number within the source input
}

// ReadRune reads a single rune from the input text.
// This method implements the io.RuneReader interface.
func (s *Scanner) ReadRune() (rune, int, error) {
	if int(s.pos) >= len(s.input) {
		s.width = nil

		return -1, -1, io.EOF
	}

	r, w := utf8.DecodeRuneInString(s.input[s.pos:])

	s.width = append(s.width, w)
	s.pos += w

	if r == '\n' {
		s.line++
	}

	return r, w, nil
}

// UnreadRune unreads the last read rune, the next call
// to ReadRune will return the just unread rune.
// This method implements the io.RuneScanner interface along
// with ReadRune.
func (s *Scanner) UnreadRune() error {
	if len(s.width) < 1 {
		return errors.New("no runes to unread")
	}

	var w int
	w, s.width = s.width[len(s.width)-1], s.width[:len(s.width)-1]

	s.pos -= w

	return nil
}

// Match attempts to match the provided regex from the current
// location of the scanner, returning the first matched
// instance of the regex as a string if a match is found and
// an error otherwise.
func (s *Scanner) Match(re *regexp.Regexp) (string, error) {
	start := s.pos

	m := re.FindReaderIndex(s)
	if m == nil {
		return "", fmt.Errorf("scanner does not match %q at position %v", re.String(), start)
	}

	if m[0] != 0 {
		return "", fmt.Errorf("scanner does not match %q at position %v", re.String(), start)
	}

	s.pos = start + m[1]
	return s.input[start+m[0] : start+m[1]], nil
}

// MatchString attempts to match the provided target string
// rune-by-rune exactly as specified, returning the target string
// if matched or an error if it was unable to match the string.
func (s *Scanner) MatchString(target string) (string, error) {
	for _, r := range target {
		o, _, err := s.ReadRune()
		if err != nil {
			return "", err
		}

		if r != o {
			return "", fmt.Errorf("scanner does not contain %q at position %v", target, s.pos)
		}
	}

	return target, nil
}

// Remaining returns the remaining unread portion of the input string.
func (s *Scanner) Remaining() string {
	return s.input[s.pos:]
}
