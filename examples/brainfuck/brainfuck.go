// Command brainfuck parses and executes Brainfuck programs.
//
// Run it with:
//
//	go run ./examples/brainfuck examples/brainfuck/sample.bf
package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/stntngo/avram/avramx"
)

type operation uint8

const (
	moveLeft operation = iota
	moveRight
	increment
	decrement
	read
	write
	loop
)

type instruction struct {
	op   operation
	body []instruction
}

func Just(want rune) avramx.Parser[rune, rune] {
	return avramx.Match(func(got rune) error {
		if got != want {
			return fmt.Errorf("expected %q, got %q", want, got)
		}
		return nil
	})
}

func parser() avramx.Parser[rune, []instruction] {
	return avramx.Fix(
		func(program avramx.Parser[rune, []instruction]) avramx.Parser[rune, []instruction] {
			simple := Just('<').To(instruction{op: moveLeft}).
				Or(Just('>').To(instruction{op: moveRight})).
				Or(Just('+').To(instruction{op: increment})).
				Or(Just('-').To(instruction{op: decrement})).
				Or(Just(',').To(instruction{op: read})).
				Or(Just('.').To(instruction{op: write}))

			loop := program.
				Between(Just('['), Just(']')).
				Map(func(body []instruction) instruction {
					return instruction{op: loop, body: body}
				}).
				Named("loop")

			return simple.
				Or(loop).
				Many()
		},
	)
}

func end() avramx.Parser[rune, avramx.Unit] {
	return func(scanner *avramx.Scanner[rune]) (avramx.Unit, error) {
		got, err := scanner.Read()
		if errors.Is(err, io.EOF) {
			return avramx.Unit{}, nil
		}
		if err != nil {
			return avramx.Unit{}, err
		}

		switch got {
		case '[':
			return avramx.Unit{}, errors.New("unclosed '['")
		case ']':
			return avramx.Unit{}, errors.New("unmatched ']'")
		default:
			return avramx.Unit{}, fmt.Errorf("unexpected instruction %q", got)
		}
	}
}

func isCommand(r rune) bool {
	switch r {
	case '<', '>', '+', '-', ',', '.', '[', ']':
		return true
	default:
		return false
	}
}

func parse(source string) ([]instruction, error) {
	commands := avramx.Filter(
		avramx.SliceIterator([]rune(source)),
		isCommand,
	)

	return parser().
		ThenIgnore(end()).
		Parse(commands)
}

const tapeLength = 10_000

type machine struct {
	tape    [tapeLength]byte
	pointer int
	input   io.Reader
	output  io.Writer
}

func (m *machine) execute(program []instruction) error {
	for _, instruction := range program {
		switch instruction.op {
		case moveLeft:
			m.pointer = (m.pointer + tapeLength - 1) % tapeLength
		case moveRight:
			m.pointer = (m.pointer + 1) % tapeLength
		case increment:
			m.tape[m.pointer]++
		case decrement:
			m.tape[m.pointer]--
		case read:
			if _, err := io.ReadFull(m.input, m.tape[m.pointer:m.pointer+1]); err != nil {
				return fmt.Errorf("read input: %w", err)
			}
		case write:
			if _, err := m.output.Write(m.tape[m.pointer : m.pointer+1]); err != nil {
				return fmt.Errorf("write output: %w", err)
			}
		case loop:
			for m.tape[m.pointer] != 0 {
				if err := m.execute(instruction.body); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unknown operation %d", instruction.op)
		}
	}

	return nil
}

func run(source string, input io.Reader, output io.Writer) error {
	program, err := parse(source)
	if err != nil {
		return fmt.Errorf("parse program: %w", err)
	}

	return (&machine{input: input, output: output}).execute(program)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: brainfuck <file>")
		os.Exit(2)
	}

	source, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "read program: %v\n", err)
		os.Exit(1)
	}

	if err := run(string(source), os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
