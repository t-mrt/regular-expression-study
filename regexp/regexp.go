package regexp

import (
	"errors"
	"strings"

	"github.com/deckarep/golang-set"
)

type nondeterministicFiniteAutomaton struct {
	transition func(state int, char string) mapset.Set
	start      int
	accepts    mapset.Set
}

func (nfa nondeterministicFiniteAutomaton) epsilonExpand(set mapset.Set) mapset.Set {

	que := set.ToSlice()
	done := mapset.NewSet()

	for {

		if len(que) == 0 {
			break
		}

		stat := que[len(que)-1]
		que = que[:len(que)-1]
		nexts := nfa.transition(stat.(int), "")
		done.Add(stat)
		it := nexts.Iterator()
		for nextStat := range it.C {
			if !done.Contains(nextStat) {
				que = append(que, nextStat)
			}
		}
	}
	return done
}

func nfa2dfa(nfa nondeterministicFiniteAutomaton) deterministicFiniteAutomaton {
	transition := func(set mapset.Set, alpha string) mapset.Set {
		ret := mapset.NewSetFromSlice([]interface{}{})
		it := set.Iterator()

		for elem := range it.C {
			ret = ret.Union(nfa.transition(elem.(int), alpha))
		}

		return nfa.epsilonExpand(ret)
	}

	return deterministicFiniteAutomaton{
		transition: transition,
		start:      nfa.epsilonExpand(mapset.NewSetFromSlice([]interface{}{nfa.start})),
		accepts:    nfa.accepts,
	}
}

type deterministicFiniteAutomaton struct {
	transition func(state mapset.Set, char string) mapset.Set
	start      mapset.Set
	accepts    mapset.Set
}

type dfaRuntime struct {
	DFA          deterministicFiniteAutomaton
	currentState mapset.Set
}

func (r *dfaRuntime) doTransition(char string) {
	r.currentState = r.DFA.transition(r.currentState, char)
}

func (r dfaRuntime) isAcceptState() bool {
	return r.currentState.Intersect(r.DFA.accepts).Cardinality() > 0
}

func (r dfaRuntime) doesAccept(input string) bool {
	for _, rune := range input {
		r.doTransition(string(rune))
	}
	return r.isAcceptState()
}

func (d deterministicFiniteAutomaton) getRuntime() dfaRuntime {
	return dfaRuntime{
		currentState: d.start,
		DFA:          d,
	}
}

type token struct {
	value string
	kind  int
}

type lexer struct {
	stringArray []string
	index       int
}

const (
	tokenCharactor  = 0
	tokenOpeUnion   = 1
	tokenOpeStar    = 2
	tokenLeftparen  = 3
	tokenRightParen = 4
	tokenEOF        = 5
)

type parser struct {
	lexer lexer
	look  token
}

func (l *lexer) scan() token {

	if l.index == len(l.stringArray) {
		return token{
			value: "",
			kind:  tokenEOF,
		}
	}

	ch := l.stringArray[l.index]
	l.index++

	if ch == `\` {
		l.index++
		return token{
			value: l.stringArray[l.index-1],
			kind:  tokenCharactor,
		}
	}
	if ch == "|" {
		return token{
			value: ch,
			kind:  tokenOpeUnion,
		}
	}
	if ch == "(" {
		return token{
			value: "(",
			kind:  tokenLeftparen,
		}
	}
	if ch == ")" {
		return token{
			value: ")",
			kind:  tokenRightParen,
		}
	}
	if ch == "*" {
		return token{
			value: ch,
			kind:  tokenOpeStar,
		}
	}
	return token{
		value: ch,
		kind:  tokenCharactor,
	}
}

func (p *parser) match(tag int) error {
	if p.look.kind != tag {
		return errors.New("syntax error")
	}
	p.move()

	return nil
}

func (p *parser) move() {
	p.look = p.lexer.scan()
}

func (p *parser) factor() node {
	if p.look.kind == tokenLeftparen {
		// factor -> '(' subexpr ')'
		p.match(tokenLeftparen)
		node := p.subexpr()
		p.match(tokenRightParen)
		return node
	} else {
		// factor -> character
		node := character{
			char: p.look.value,
		}
		p.match(tokenCharactor)
		return node
	}
}

func (p *parser) star() node {
	// tar -> factor '*' | factor
	node := p.factor()
	if p.look.kind == tokenOpeStar {
		p.match(tokenOpeStar)
		node = star{
			operand: node,
		}
	}

	return node
}

func (p *parser) seq() node {
	if p.look.kind == tokenLeftparen || p.look.kind == tokenCharactor {
		// seq -> subseq
		return p.subseq()
	} else {
		// seq -> ''
		return character{
			char: "",
		}
	}
}

func (p *parser) subseq() node {
	node1 := p.star()
	if p.look.kind == tokenLeftparen || p.look.kind == tokenCharactor {
		// subseq -> star subseq
		node2 := p.subseq()
		node := concat{
			operand1: node1,
			operand2: node2,
		}
		return node
	} else {
		// subseq -> star
		return node1
	}
}

func (p *parser) subexpr() node {
	// subexpr    -> seq '|' subexpr | seq
	node := p.seq()
	if p.look.kind == tokenOpeUnion {
		p.match(tokenOpeUnion)
		node2 := p.subexpr()
		node = union{
			operand1: node,
			operand2: node2,
		}
	}
	return node
}

func (p *parser) expression() nondeterministicFiniteAutomaton {
	// expression -> subexpr tokenEOF
	node := p.subexpr()
	p.match(tokenEOF)

	context := context{}
	fragment := node.assemble(&context)
	return fragment.build()
}

type context struct {
	stateCount int
}

func (c *context) newState() int {
	c.stateCount = c.stateCount + 1
	return c.stateCount
}

type stateChar struct {
	state int
	char  string
}

type nfaFragment struct {
	start        int
	accepts      mapset.Set
	stateCharMap map[stateChar]mapset.Set // (状態, 入力文字) => 次の状態
}

func (f *nfaFragment) connect(from int, char string, to int) {
	sc := stateChar{
		state: from,
		char:  char,
	}
	s := mapset.NewSet()
	if val, ok := f.stateCharMap[sc]; ok {
		val.Add(to)
		f.stateCharMap[sc] = val
	} else {
		s.Add(to)
		f.stateCharMap[sc] = s
	}
}

func (f nfaFragment) newSkelton() nfaFragment {
	newFragment := nfaFragment{
		start:        0,
		accepts:      mapset.NewSet(),
		stateCharMap: map[stateChar]mapset.Set{},
	}

	newFragment.stateCharMap = f.stateCharMap

	return newFragment
}

func (f nfaFragment) or(frag nfaFragment) nfaFragment {
	newFrag := f.newSkelton()

	for k, v := range frag.stateCharMap {
		newFrag.stateCharMap[k] = v
	}

	return newFrag
}

func (f nfaFragment) build() nondeterministicFiniteAutomaton {
	scmap := f.stateCharMap

	transition := func(state int, char string) mapset.Set {
		sc := stateChar{
			state: state,
			char:  char,
		}

		if val, ok := scmap[sc]; ok {
			return val
		} else {
			return mapset.NewSet()
		}
	}

	return nondeterministicFiniteAutomaton{
		transition: transition,
		start:      f.start,
		accepts:    f.accepts,
	}
}

type node interface {
	assemble(*context) nfaFragment
}

type character struct {
	char string
}

func (c character) assemble(context *context) nfaFragment {
	frag := nfaFragment{
		stateCharMap: map[stateChar]mapset.Set{},
	}

	s1 := context.newState()
	s2 := context.newState()

	frag.connect(s1, c.char, s2)

	frag.start = s1
	s := mapset.NewSet()
	s.Add(s2)
	frag.accepts = s

	return frag
}

type union struct {
	operand1 node
	operand2 node
}

func (u union) assemble(context *context) nfaFragment {
	frag1 := u.operand1.assemble(context)
	frag2 := u.operand2.assemble(context)
	frag := frag1.or(frag2)

	s := context.newState()
	frag.connect(s, "", frag1.start)
	frag.connect(s, "", frag2.start)

	frag.start = s
	frag.accepts = frag1.accepts.Union(frag2.accepts)

	return frag
}

type star struct {
	operand node
}

func (s star) assemble(context *context) nfaFragment {
	fragOrig := s.operand.assemble(context)
	frag := fragOrig.newSkelton()

	it := fragOrig.accepts.Iterator()

	for state := range it.C {
		frag.connect(state.(int), "", fragOrig.start)
	}
	state := context.newState()
	frag.connect(state, "", fragOrig.start)

	frag.start = state

	set := mapset.NewSet()
	set.Add(state)
	frag.accepts = fragOrig.accepts.Union(set)

	return frag
}

type concat struct {
	operand1 node
	operand2 node
}

func (c concat) assemble(context *context) nfaFragment {

	frag1 := c.operand1.assemble(context)
	frag2 := c.operand2.assemble(context)
	frag := frag1.or(frag2)

	it := frag1.accepts.Iterator()

	for state := range it.C {
		frag.connect(state.(int), "", frag2.start)
	}
	frag.start = frag1.start
	frag.accepts = frag2.accepts

	return frag
}

type Regexp interface {
	Match(regexp string) bool
}

type regexp struct {
	regexp string
	dfa    deterministicFiniteAutomaton
}

func (r *regexp) compile() {
	lexer := lexer{
		stringArray: strings.Split(r.regexp, ""),
		index:       0,
	}
	parser := parser{
		lexer: lexer,
	}
	parser.move()
	nfa := parser.expression()

	r.dfa = nfa2dfa(nfa)
}

func (r regexp) Match(input string) bool {
	runtime := r.dfa.getRuntime()
	return runtime.doesAccept(input)
}

func NewRegexp(regexpString string) Regexp {
	r := regexp{
		regexp: regexpString,
	}
	r.compile()

	return r
}
