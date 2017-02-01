package main

import (
	"fmt"

	"github.com/t-mrt/regular-expression-study"
)

func main() {

	regExp := `(あ|い)*うえ*(お|)`
	r := regexp.NewRegexp(regExp)

	str := "あうえ"
	if r.Match(str) {
		fmt.Printf("%v match to %v\n", str, regExp)
	} else {
		fmt.Printf("%v does not match to %v\n", str, regExp)
	}

	rs := regexp.NewRandString(regExp)

	fmt.Printf(rs.Generate())
}
