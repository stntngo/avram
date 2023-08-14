package avram_test

import (
	"regexp"
	"strconv"
	"testing"

	. "github.com/stntngo/avram"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type JSONNode interface {
	json()
}

func (Number) json() {}
func (String) json() {}
func (Array) json()  {}
func (Object) json() {}
func (Null) json()   {}

type Number float64

type String string

type Array []JSONNode

type Object map[String]JSONNode

type Null struct{}

func jsonify[T JSONNode](p Parser[T]) Parser[JSONNode] {
	return func(s *Scanner) (JSONNode, error) {
		v, err := p(s)
		if err != nil {
			return nil, err
		}

		return v, nil
	}
}

func tointerface(n JSONNode) any {
	switch v := n.(type) {
	case Number:
		return float64(v)
	case String:
		return string(v)
	case Array:
		out := make([]any, len(v))
		for i, e := range v {
			out[i] = tointerface(e)
		}

		return out
	case Null:
		return nil
	case Object:
		out := make(map[string]any)
		for key, value := range v {
			out[string(key)] = tointerface(value)
		}

		return out
	default:
		panic("unknown type")
	}
}

var parsejson = Finish(Fix(
	func(json Parser[JSONNode]) Parser[JSONNode] {
		parsenumber := Lift(
			func(s string) Number {
				f, err := strconv.ParseFloat(s, 64)
				if err != nil {
					panic(err)
				}

				return Number(f)
			},

			MatchRegexp(regexp.MustCompile(`[-+]?([0-9]*\.[0-9]+|[0-9]+)`)),
		)

		parsenull := DiscardLeft(MatchString("null"), Return(Null{}))

		parsequoted := Wrap(
			Rune('"'),
			func(s *Scanner) (string, error) {
				var escaped bool

				var out string

				for {
					r, _, err := s.ReadRune()
					if err != nil {
						for range out {
							if err := s.UnreadRune(); err != nil {
								return "", err
							}
						}

						return "", err
					}

					if !escaped && r == '"' {
						if err := s.UnreadRune(); err != nil {
							return "", err
						}

						break
					}

					out += string(r)

					if !escaped && r == '\\' {
						escaped = true
					} else {
						escaped = false
					}
				}

				return out, nil
			},
			Rune('"'),
		)

		parsestring := Lift(func(s string) String { return String(s) }, parsequoted)

		parseArray := Lift(
			func(arr []JSONNode) JSONNode {
				return Array(arr)
			},
			Wrap(Rune('['), SepBy(Rune(','), json), Rune(']')),
		)

		parseObject := Lift(
			func(pairs []Pair[String, JSONNode]) JSONNode {
				object := make(Object)
				for _, pair := range pairs {
					object[pair.Left] = pair.Right
				}

				return object
			},
			Wrap(
				Rune('{'),
				SepBy(
					Rune(','),
					SkipWS(Both(
						parsestring,
						DiscardLeft(
							Rune(':'),
							json,
						),
					)),
				),
				Rune('}'),
			),
		)

		return SkipWS(TryChoice(
			"json object",
			jsonify(parsenull),
			jsonify(parsestring),
			jsonify(parsenumber),
			parseArray,
			parseObject,
		))
	},
))

func TestJSON(t *testing.T) {
	for _, tt := range []struct {
		name     string
		raw      string
		expected any
	}{
		{
			"simple string",
			`"test string"`,
			"test string",
		},
		{
			"simple number",
			`10`,
			float64(10),
		},
		{
			"simple array",
			`[1, 2, 3, 4]`,
			[]any{1.0, 2.0, 3.0, 4.0},
		},
		{
			"simple object",
			`{"key_one": "value", "some_number": 10}`,
			map[string]any{
				"key_one":     "value",
				"some_number": 10.0,
			},
		},
		{
			"complex nested types",
			`{"test": ["abc123", 3.14, [4.5, "23"]], "two": null}`,
			map[string]any{
				"test": []any{
					"abc123",
					3.14,
					[]any{4.5, "23"},
				},
				"two": nil,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parsejson(NewScanner(tt.raw))
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tointerface(parsed))
		})
	}
}
