package avramx_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stntngo/avram/avramx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLift(t *testing.T) {
	tests := []struct {
		name      string
		tokens    []token
		transform func(token) (string, error)
		wantErr   bool
		want      string
	}{
		{
			name:   "lift success",
			tokens: []token{"hello"},
			transform: func(t token) (string, error) {
				return string(t) + " world", nil
			},
			want: "hello world",
		},
		{
			name:   "lift transform error",
			tokens: []token{"hello"},
			transform: func(t token) (string, error) {
				return "", errors.New("transform error")
			},
			wantErr: true,
		},
		{
			name:   "lift parser error",
			tokens: []token{"world"},
			transform: func(t token) (string, error) {
				return string(t), nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := createIterator(tt.tokens)
			parser := avramx.Lift(tt.transform, avramx.Match(match("hello")))
			result, err := avramx.Parse(it, parser)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestLift2(t *testing.T) {
	tests := []struct {
		name      string
		tokens    []token
		transform func(token, token) (string, error)
		wantErr   bool
		want      string
	}{
		{
			name:   "lift2 success",
			tokens: []token{"hello", "world"},
			transform: func(a, b token) (string, error) {
				return string(a) + " " + string(b), nil
			},
			want: "hello world",
		},
		{
			name:   "lift2 transform error",
			tokens: []token{"hello", "world"},
			transform: func(a, b token) (string, error) {
				return "", errors.New("transform error")
			},
			wantErr: true,
		},
		{
			name:   "lift2 first parser error",
			tokens: []token{"bye", "world"},
			transform: func(a, b token) (string, error) {
				return string(a) + " " + string(b), nil
			},
			wantErr: true,
		},
		{
			name:   "lift2 second parser error",
			tokens: []token{"hello", "bye"},
			transform: func(a, b token) (string, error) {
				return string(a) + " " + string(b), nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := createIterator(tt.tokens)
			parser := avramx.Lift2(
				tt.transform,
				avramx.Match(match("hello")),
				avramx.Match(match("world")),
			)
			result, err := avramx.Parse(it, parser)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestLift3(t *testing.T) {
	tests := []struct {
		name      string
		tokens    []token
		transform func(token, token, token) (string, error)
		wantErr   bool
		want      string
	}{
		{
			name:   "lift3 success",
			tokens: []token{"a", "b", "c"},
			transform: func(a, b, c token) (string, error) {
				return string(a) + string(b) + string(c), nil
			},
			want: "abc",
		},
		{
			name:   "lift3 transform error",
			tokens: []token{"a", "b", "c"},
			transform: func(a, b, c token) (string, error) {
				return "", errors.New("transform error")
			},
			wantErr: true,
		},
		{
			name:   "lift3 first parser error",
			tokens: []token{"x", "b", "c"},
			transform: func(a, b, c token) (string, error) {
				return string(a) + string(b) + string(c), nil
			},
			wantErr: true,
		},
		{
			name:   "lift3 second parser error",
			tokens: []token{"a", "x", "c"},
			transform: func(a, b, c token) (string, error) {
				return string(a) + string(b) + string(c), nil
			},
			wantErr: true,
		},
		{
			name:   "lift3 third parser error",
			tokens: []token{"a", "b", "x"},
			transform: func(a, b, c token) (string, error) {
				return string(a) + string(b) + string(c), nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := createIterator(tt.tokens)
			parser := avramx.Lift3(
				tt.transform,
				avramx.Match(match("a")),
				avramx.Match(match("b")),
				avramx.Match(match("c")),
			)
			result, err := avramx.Parse(it, parser)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestLift4(t *testing.T) {
	tests := []struct {
		name      string
		tokens    []token
		transform func(token, token, token, token) (string, error)
		wantErr   bool
		want      string
	}{
		{
			name:   "lift4 success",
			tokens: []token{"a", "b", "c", "d"},
			transform: func(a, b, c, d token) (string, error) {
				return string(a) + string(b) + string(c) + string(d), nil
			},
			want: "abcd",
		},
		{
			name:   "lift4 transform error",
			tokens: []token{"a", "b", "c", "d"},
			transform: func(a, b, c, d token) (string, error) {
				return "", errors.New("transform error")
			},
			wantErr: true,
		},
		{
			name:   "lift4 first parser error",
			tokens: []token{"x", "b", "c", "d"},
			transform: func(a, b, c, d token) (string, error) {
				return string(a) + string(b) + string(c) + string(d), nil
			},
			wantErr: true,
		},
		{
			name:   "lift4 second parser error",
			tokens: []token{"a", "x", "c", "d"},
			transform: func(a, b, c, d token) (string, error) {
				return string(a) + string(b) + string(c) + string(d), nil
			},
			wantErr: true,
		},
		{
			name:   "lift4 third parser error",
			tokens: []token{"a", "b", "x", "d"},
			transform: func(a, b, c, d token) (string, error) {
				return string(a) + string(b) + string(c) + string(d), nil
			},
			wantErr: true,
		},
		{
			name:   "lift4 fourth parser error",
			tokens: []token{"a", "b", "c", "x"},
			transform: func(a, b, c, d token) (string, error) {
				return string(a) + string(b) + string(c) + string(d), nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := createIterator(tt.tokens)
			parser := avramx.Lift4(
				tt.transform,
				avramx.Match(match("a")),
				avramx.Match(match("b")),
				avramx.Match(match("c")),
				avramx.Match(match("d")),
			)
			result, err := avramx.Parse(it, parser)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestLiftWithNumbers(t *testing.T) {
	// Test Lift with actual number parsing
	it := createIterator([]token{"123"})

	parser := avramx.Lift(
		func(t token) (int, error) {
			return strconv.Atoi(string(t))
		},
		avramx.Match(func(t token) error {
			_, err := strconv.Atoi(string(t))
			return err
		}),
	)

	result, err := avramx.Parse(it, parser)
	require.NoError(t, err)
	assert.Equal(t, 123, result)
}

func TestLift2WithMath(t *testing.T) {
	// Test Lift2 with addition
	it := createIterator([]token{"10", "20"})

	parser := avramx.Lift2(
		func(a, b token) (int, error) {
			numA, err := strconv.Atoi(string(a))
			if err != nil {
				return 0, err
			}
			numB, err := strconv.Atoi(string(b))
			if err != nil {
				return 0, err
			}
			return numA + numB, nil
		},
		avramx.Match(func(t token) error {
			_, err := strconv.Atoi(string(t))
			return err
		}),
		avramx.Match(func(t token) error {
			_, err := strconv.Atoi(string(t))
			return err
		}),
	)

	result, err := avramx.Parse(it, parser)
	require.NoError(t, err)
	assert.Equal(t, 30, result)
}
