package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	program, err := parse("++ hello world [>+<-]")
	require.NoError(t, err)
	require.Len(t, program, 3)
	assert.Equal(t, increment, program[0].op)
	assert.Equal(t, increment, program[1].op)
	assert.Equal(t, loop, program[2].op)
	require.Len(t, program[2].body, 4)
	assert.Equal(t, []operation{moveRight, increment, moveLeft, decrement}, []operation{
		program[2].body[0].op,
		program[2].body[1].op,
		program[2].body[2].op,
		program[2].body[3].op,
	})
}

func TestParseRejectsUnbalancedLoops(t *testing.T) {
	for _, source := range []string{"[+", "+]"} {
		t.Run(source, func(t *testing.T) {
			_, err := parse(source)
			require.Error(t, err)
		})
	}
}

func TestRun(t *testing.T) {
	for _, tt := range []struct {
		name   string
		source string
		input  string
		output string
	}{
		{
			name:   "compute and write A",
			source: "++++++++[>++++++++<-]>+.",
			output: "A",
		},
		{
			name:   "read and echo",
			source: ",.",
			input:  "Z",
			output: "Z",
		},
		{
			name:   "wrap pointer left",
			source: "<+.",
			output: "\x01",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			err := run(tt.source, strings.NewReader(tt.input), &output)
			require.NoError(t, err)
			assert.Equal(t, tt.output, output.String())
		})
	}
}
