package avram

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"unicode/utf8"

	"go.uber.org/multierr"
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
//
// Scanner exposes three matching primitives for the parsers to use:
//
// 1. MatchRegexp - Matching on compiled regular expressions
// 2. MatchString - Matching on concrete strings
// 3. MatchRune   - Matching on an individual rune
//
// Each of these matching methods can potentially advance the state of the
// the scanner along the input if they successfully find a match given
// the provided matching criteria.
//
// If no match is found in any of these matching primitives, the state
// of the scanner is not advanced.
type Scanner struct {
	// OPTIM: (niels) Use something more nuanced like an io.RuneScanner
	// that we track the position of as we read / unread runes rather
	// than requiring the full string to be available before beginning
	// to scan the input.
	input string // the string being lexed
	start int    // location of the end of the last emitted token
	pos   int    // current position of the lexer in the input
	width []int  // width history of read but un-emitted runes from the input
	line  int    // current line number within the source input
}

// ReadRune reads a single rune from the input text.
//
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
//
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

// MatchRegexp attempts to match the provided regex from the current
// location of the scanner, returning the first matched
// instance of the regex as a string if a match is found and
// an error otherwise.
//
// NOTE: MatchRegexp only advances the scanner position if a valid
// match is successfully found.
func (s *Scanner) MatchRegexp(re *regexp.Regexp) (string, error) {
	start := s.pos

	m := re.FindReaderIndex(s)
	if m == nil {
		s.pos = start
		return "", fmt.Errorf("scanner does not match %q at position %v", re.String(), start)
	}

	if m[0] != 0 {
		s.pos = start
		return "", fmt.Errorf("scanner does not match %q at position %v", re.String(), start)
	}

	s.pos = start + m[1]
	return s.input[start+m[0] : start+m[1]], nil
}

// MatchString attempts to match the provided target string
// rune-by-rune exactly as specified, returning the target string
// if matched or an error if it was unable to match the string.
//
// NOTE: MatchString only advances the scanner position if a valid
// match is successfully found.
func (s *Scanner) MatchString(target string) (string, error) {
	checkpoint := s.pos

	for _, r := range target {
		o, _, err := s.ReadRune()
		if err != nil {
			s.pos = checkpoint
			return "", err
		}

		if r != o {
			s.pos = checkpoint
			return "", fmt.Errorf("scanner does not contain %q at position %v", target, s.pos)
		}
	}

	return target, nil
}

// MatchRune attempts to match the provided predicate function
// with the next rune in the scanner's input stream.
//
// NOTE: MatchRune only advances the scanner position if a valid
// match is successfully found.
func (s *Scanner) MatchRune(match func(rune) error) (r rune, err error) {
	r, _, err = s.ReadRune()
	if err != nil {
		return -1, err
	}

	if err := match(r); err != nil {
		return -1, multierr.Append(err, s.UnreadRune())
	}

	return r, nil
}

// Remaining returns the remaining unread portion of the input string.
func (s *Scanner) Remaining() string {
	return s.input[s.pos:]
}

// Finish meta-parser ensures that the completed parser has successfully
// parsed the entirety of the input string contained in the scanner.
func Finish[A any](p Parser[A]) Parser[A] {
	return func(s *Scanner) (A, error) {
		parsed, err := p(s)
		if err != nil {
			var zero A
			return zero, err
		}

		if rem := s.Remaining(); len(rem) > 0 {
			var zero A
			return zero, fmt.Errorf("unparsed input: %q", rem)
		}

		return parsed, nil
	}
}

// Location meta-parser tracks the start and end location of successfully parsed
// inputs, allowing the start and end location to be combined into the
// parsed value through the provided function.
func Location[A, B any](p Parser[A], f func(start int, end int, parsed A) B) Parser[B] {
	return func(s *Scanner) (B, error) {
		start := s.pos

		a, err := p(s)
		if err != nil {
			var zero B
			return zero, err
		}

		end := s.pos

		return f(start, end, a), nil
	}
}
