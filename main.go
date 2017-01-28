package main

import (
	"fmt"

	"github.com/t-mrt/regular-expression-study/regexp"
)

func main() {
	r := regexp.NewRegexp(`a`)
	fmt.Println(r.Match("a"))
}
