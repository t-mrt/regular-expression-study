package main

import (
	"fmt"

	"github.com/t-mrt/regular-expression-study"
)

func main() {

	regExp := `a|b`
	str := "c"

	r := regexp.NewRegexp(regExp)
	if r.Match(str) {
		fmt.Printf("%v is match to %v\n", str, regExp)
	} else {
		fmt.Printf("%v is not match to %v\n", str, regExp)
	}
}
