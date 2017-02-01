package regexp

import (
	"strings"
	"testing"

	"github.com/deckarep/golang-set"
)

func TestNondeterministicFiniteAutomaton(t *testing.T) {
	accepts := mapset.NewSetFromSlice([]interface{}{2})

	transition := func(state int, char string) mapset.Set {
		if state == 0 && char == "a" {
			return mapset.NewSetFromSlice([]interface{}{1, 2})
		}
		if state == 1 && char == "b" {
			return mapset.NewSetFromSlice([]interface{}{2})
		}
		if state == 2 && char == "" {
			return mapset.NewSetFromSlice([]interface{}{0})
		}
		return mapset.NewSet()
	}

	nfa := nondeterministicFiniteAutomaton{
		transition: transition,
		start:      0,
		accepts:    accepts,
	}

	if nfa.start != 0 {
		t.Fatal("nfa.start")
	}

	if !nfa.accepts.Equal(accepts) {
		t.Fatal("nfa.accepts")
	}

	if !nfa.transition(0, "a").Equal(mapset.NewSetFromSlice([]interface{}{1, 2})) {
		t.Fatal("nfa.transition")
	}

	if !nfa.transition(0, "").Equal(mapset.NewSetFromSlice([]interface{}{})) {
		t.Fatal("nfa.transition")
	}

	if !nfa.transition(1, "b").Equal(mapset.NewSetFromSlice([]interface{}{2})) {
		t.Fatal("nfa.transition")
	}

	if !nfa.transition(2, "").Equal(mapset.NewSetFromSlice([]interface{}{0})) {
		t.Fatal("nfa.transition")
	}

	if !nfa.epsilonExpand(mapset.NewSetFromSlice([]interface{}{0})).Equal(nfa.epsilonExpand(mapset.NewSetFromSlice([]interface{}{0}))) {
		t.Fatal("nfa.transition")
	}
	if !nfa.epsilonExpand(mapset.NewSetFromSlice([]interface{}{1})).Equal(nfa.epsilonExpand(mapset.NewSetFromSlice([]interface{}{1}))) {
		t.Fatal("nfa.transition")
	}
	if !nfa.epsilonExpand(mapset.NewSetFromSlice([]interface{}{2})).Equal(nfa.epsilonExpand(mapset.NewSetFromSlice([]interface{}{0, 2}))) {
		t.Fatal("nfa.transition")
	}

}

func TestDeterministicFiniteAutomaton(t *testing.T) {
	accepts := mapset.NewSet()
	accepts.Add(mapset.NewSetFromSlice([]interface{}{0, 1, 2}))
	accepts.Add(mapset.NewSetFromSlice([]interface{}{0, 2}))

	transition := func(state mapset.Set, char string) mapset.Set {
		if char == "" {
			panic("空文字")
		}

		if state.Equal(mapset.NewSetFromSlice([]interface{}{0})) && char == "a" {
			return mapset.NewSetFromSlice([]interface{}{0, 1, 2})
		}

		if state.Equal(mapset.NewSetFromSlice([]interface{}{0, 1, 2})) && char == "a" {
			return mapset.NewSetFromSlice([]interface{}{0, 1, 2})
		}

		if state.Equal(mapset.NewSetFromSlice([]interface{}{0, 1, 2})) && char == "b" {
			return mapset.NewSetFromSlice([]interface{}{0, 2})
		}

		if state.Equal(mapset.NewSetFromSlice([]interface{}{0, 2})) && char == "a" {
			return mapset.NewSetFromSlice([]interface{}{0, 1, 2})
		}

		return mapset.NewSet()
	}

	dfa := deterministicFiniteAutomaton{
		transition: transition,
		start:      mapset.NewSetFromSlice([]interface{}{0}),
		accepts:    accepts,
	}

	if !dfa.start.Equal(mapset.NewSetFromSlice([]interface{}{0})) {
		t.Fatal("nfa.start")
	}

	if !dfa.accepts.Equal(accepts) {
		t.Fatal("nfa.accepts")
	}

	if !dfa.transition(mapset.NewSetFromSlice([]interface{}{0}), "a").Equal(mapset.NewSetFromSlice([]interface{}{0, 1, 2})) {
		t.Fatal("nfa.transition")
	}

	if !dfa.transition(mapset.NewSetFromSlice([]interface{}{0, 1, 2}), "a").Equal(mapset.NewSetFromSlice([]interface{}{0, 1, 2})) {
		t.Fatal("nfa.transition")
	}

	if !dfa.transition(mapset.NewSetFromSlice([]interface{}{0, 1, 2}), "b").Equal(mapset.NewSetFromSlice([]interface{}{0, 2})) {
		t.Fatal("nfa.transition")
	}

	if !dfa.transition(mapset.NewSetFromSlice([]interface{}{0, 2}), "a").Equal(mapset.NewSetFromSlice([]interface{}{0, 1, 2})) {
		t.Fatal("nfa.transition")
	}
}

func TestNFA2dfa(t *testing.T) {
	accepts := mapset.NewSetFromSlice([]interface{}{2})

	transition := func(state int, char string) mapset.Set {
		if state == 0 && char == "a" {
			return mapset.NewSetFromSlice([]interface{}{1, 2})
		}
		if state == 1 && char == "b" {
			return mapset.NewSetFromSlice([]interface{}{2})
		}
		if state == 2 && char == "" {
			return mapset.NewSetFromSlice([]interface{}{0})
		}
		return mapset.NewSet()
	}

	nfa := nondeterministicFiniteAutomaton{
		transition: transition,
		start:      0,
		accepts:    accepts,
	}

	dfa := nfa2dfa(nfa)

	newAccepts := mapset.NewSet()
	newAccepts.Add(mapset.NewSetFromSlice([]interface{}{0, 1, 2}))
	newAccepts.Add(mapset.NewSetFromSlice([]interface{}{0, 2}))

	if !dfa.start.Equal(mapset.NewSetFromSlice([]interface{}{0})) {
		t.Fatal("nfa.start")
	}

	if !dfa.transition(mapset.NewSetFromSlice([]interface{}{0}), "a").Equal(mapset.NewSetFromSlice([]interface{}{0, 1, 2})) {
		t.Fatal("nfa.transition")
	}

	if !dfa.transition(mapset.NewSetFromSlice([]interface{}{0, 1, 2}), "a").Equal(mapset.NewSetFromSlice([]interface{}{0, 1, 2})) {
		t.Fatal("nfa.transition")
	}

	if !dfa.transition(mapset.NewSetFromSlice([]interface{}{0, 1, 2}), "b").Equal(mapset.NewSetFromSlice([]interface{}{0, 2})) {
		t.Fatal("nfa.transition")
	}

	if !dfa.transition(mapset.NewSetFromSlice([]interface{}{0, 2}), "a").Equal(mapset.NewSetFromSlice([]interface{}{0, 1, 2})) {
		t.Fatal("nfa.transition")
	}

}

func TestDFARuntime(t *testing.T) {
	accepts := mapset.NewSetFromSlice([]interface{}{2})

	transition := func(state int, char string) mapset.Set {
		if state == 0 && char == "a" {
			return mapset.NewSetFromSlice([]interface{}{1, 2})
		}
		if state == 1 && char == "b" {
			return mapset.NewSetFromSlice([]interface{}{2})
		}
		if state == 2 && char == "" {
			return mapset.NewSetFromSlice([]interface{}{0})
		}
		return mapset.NewSet()
	}

	nfa := nondeterministicFiniteAutomaton{
		transition: transition,
		start:      0,
		accepts:    accepts,
	}

	dfa := nfa2dfa(nfa)

	runtime := dfa.getRuntime()

	if !runtime.doesAccept("ab") {
		t.Fatal("ab should accept")
	}
	if !runtime.doesAccept("aaaaaaaab") {
		t.Fatal("ab should accept")
	}
	if !runtime.doesAccept("aaaaaaabab") {
		t.Fatal("ab should accept")
	}
	if runtime.doesAccept("baaaaaaaaaaaaaaaaaaaaaaabb") {
		t.Fatal("ba should  not accept")
	}

}

func TestLexer(t *testing.T) {
	lexer := lexer{
		stringArray: strings.Split(`a|\(*`, ""),
		index:       0,
	}

	var token token

	token = lexer.scan()
	if token.kind != tokenCharactor || token.value != "a" {
		t.Fatal("a")
	}

	token = lexer.scan()
	if token.kind != tokenOpeUnion || token.value != "|" {
		t.Fatal("|")
	}

	token = lexer.scan()
	if token.kind != tokenCharactor || token.value != "(" {
		t.Fatal(`\(`)
	}

	token = lexer.scan()
	if token.kind != tokenOpeStar || token.value != "*" {
		t.Fatal(`*`)
	}

	token = lexer.scan()
	if token.kind != tokenEOF || token.value != "" {
		t.Fatal(`tokenEOF`)
	}

	token = lexer.scan()
	if token.kind != tokenEOF || token.value != "" {
		t.Fatal(`tokenEOF`)
	}
}

func TestCharacter(t *testing.T) {

	c := character{
		char: "a",
	}
	context := context{}
	f := c.assemble(&context)

	sc := stateChar{
		char:  "a",
		state: 1,
	}
	if !mapset.NewSet(2).Equal(f.stateCharMap[sc]) {
		t.Fatal(`Charactor assemble`)
	}
}

func TestConcat(t *testing.T) {

	c1 := character{
		char: "a",
	}
	c2 := character{
		char: "b",
	}

	concat := concat{
		operand1: c1,
		operand2: c2,
	}

	context := context{}
	f := concat.assemble(&context)

	sc1 := stateChar{
		char:  "a",
		state: 1,
	}

	sc2 := stateChar{
		char:  "b",
		state: 3,
	}

	sc3 := stateChar{
		char:  "",
		state: 2,
	}

	if !mapset.NewSet(2).Equal(f.stateCharMap[sc1]) {
		t.Fatal(`concat assemble`)
	}

	if !mapset.NewSet(4).Equal(f.stateCharMap[sc2]) {
		t.Fatal(`concat assemble`)
	}

	if !mapset.NewSet(3).Equal(f.stateCharMap[sc3]) {
		t.Fatal(`concat assemble`)
	}

	if !mapset.NewSet(4).Equal(f.accepts) {
		t.Fatal(`concat assemble`)
	}
}

func TestUnion(t *testing.T) {

	c1 := character{
		char: "a",
	}
	c2 := character{
		char: "b",
	}

	u := union{
		operand1: c1,
		operand2: c2,
	}

	context := context{}
	f := u.assemble(&context)

	sc1 := stateChar{
		char:  "a",
		state: 1,
	}

	sc2 := stateChar{
		char:  "b",
		state: 3,
	}

	sc3 := stateChar{
		char:  "",
		state: 5,
	}

	if !mapset.NewSet(2).Equal(f.stateCharMap[sc1]) {
		t.Fatal(`union assemble`)
	}

	if !mapset.NewSet(4).Equal(f.stateCharMap[sc2]) {
		t.Fatal(`union assemble`)
	}

	if !mapset.NewSetFromSlice([]interface{}{1, 3}).Equal(f.stateCharMap[sc3]) {
		t.Fatal(`union assemble`)
	}

	if !mapset.NewSetFromSlice([]interface{}{2, 4}).Equal(f.accepts) {
		t.Fatal(`union assemble`)
	}
}

func TestRegexp(t *testing.T) {
	tester := func(regexpt string, successCase, failCase []string) {
		r := NewRegexp(regexpt)
		for _, c := range successCase {
			if !r.Match(c) {
				t.Errorf("%v should match %v", regexpt, c)
			}
		}
		for _, c := range failCase {
			if r.Match(c) {
				t.Errorf("%v should not match %v", regexpt, c)
			}
		}
	}

	type suite struct {
		regexp      string
		successCase []string
		failCase    []string
	}

	suites := []suite{
		suite{
			regexp:      `a`,
			successCase: []string{"a"},
			failCase:    []string{"b"},
		},
		suite{
			regexp:      `ab`,
			successCase: []string{"ab"},
			failCase:    []string{"a", "b", "c", "aba"},
		},
		suite{
			regexp:      `a|b`,
			successCase: []string{"a", "b"},
			failCase:    []string{"ab", "c", "aa", "bb"},
		},
		suite{
			regexp:      `a*`,
			successCase: []string{"", "a", "aa", "aaa"},
			failCase:    []string{"b", "ab"},
		},
		suite{
			regexp:      `a\|b`,
			successCase: []string{"a|b"},
			failCase:    []string{"a", "b"},
		},
		suite{
			regexp:      `p(erl|hp)|ruby`,
			successCase: []string{"perl", "php", "ruby"},
			failCase:    []string{},
		},
		suite{
			regexp:      `w(ww)*|\(笑\)`,
			successCase: []string{"w", "www", "wwwww", "(笑)"},
			failCase:    []string{"ww", "wwww"},
		},
	}

	for _, val := range suites {
		tester(val.regexp, val.successCase, val.failCase)
	}
}
