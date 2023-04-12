# avram

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/stntngo/avram?style=flat-square)
![GitHub release (latest by date including pre-releases)](https://img.shields.io/github/v/release/stntngo/avram?include_prereleases&style=flat-square)
[![GoDoc Reference](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://pkg.go.dev/github.com/stntngo/avram)

avram is a generic parser-combinator library inspired by the likes of [ocaml's Angstrom](https://github.com/inhabitedtype/angstrom) and [Haskell's Parsec](https://hackage.haskell.org/package/parsec) ported to Go and making use of Go's generic data types.

avram is an early work in progress and has not been extensively performance profiled or tuned in any way. The initial goal of this project was to provide the comfortable developer ergonomics of angstrom within the Go ecosystem. The API is very similar to the angstrom API as a result, replacing ocaml's custom infix operators with traditional Go functions where required.

## Usage

This example is a direct port of the provided usage example in the [angstrom documentation](https://github.com/inhabitedtype/angstrom#usage).

```go
package main

import (
	"strconv"

	. "github.com/stntngo/avram"
	"github.com/stntngo/avram/result"
)

var expr = Finish(Fix(func(expr Parser[int]) Parser[int] {
	add := DiscardLeft(SkipWS(Rune('+')), Return(func(a, b int) int { return a + b }))
	sub := DiscardLeft(SkipWS(Rune('-')), Return(func(a, b int) int { return a - b }))
	mul := DiscardLeft(SkipWS(Rune('*')), Return(func(a, b int) int { return a * b }))
	div := DiscardLeft(SkipWS(Rune('/')), Return(func(a, b int) int { return a / b }))

	integer := result.Unwrap(Lift(
		result.Lift(strconv.Atoi),
		TakeWhile1(Runes('0', '1', '2', '3', '4', '5', '6', '7', '8', '9')),
	))

	factor := Or(Wrap(Rune('('), expr, Rune(')')), integer)
	term := ChainL1(factor, Or(mul, div))

	return ChainL1(term, Or(add, sub))
}))

func Eval(s string) (int, error) {
	out, err := expr(NewScanner(s))
	if err != nil {
		return 0, err
	}

	return out, nil
}
```
