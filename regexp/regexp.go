package regexp

import (
	"errors"
	"strings"

	"github.com/deckarep/golang-set"
)

type NondeterministicFiniteAutomaton struct {
	transition func(state int, char string) mapset.Set
	start      int
	accepts    mapset.Set
}

func (nfa NondeterministicFiniteAutomaton) epsilonExpand(set mapset.Set) mapset.Set {

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

func nfa2dfa(nfa NondeterministicFiniteAutomaton) DeterministicFiniteAutomaton {
	transition := func(set mapset.Set, alpha string) mapset.Set {
		ret := mapset.NewSetFromSlice([]interface{}{})
		it := set.Iterator()

		for elem := range it.C {
			ret = ret.Union(nfa.transition(elem.(int), alpha))
		}

		return nfa.epsilonExpand(ret)
	}

	return DeterministicFiniteAutomaton{
		transition: transition,
		start:      nfa.epsilonExpand(mapset.NewSetFromSlice([]interface{}{nfa.start})),
		accepts:    nfa.accepts,
	}
}

type DeterministicFiniteAutomaton struct {
	transition func(state mapset.Set, char string) mapset.Set
	start      mapset.Set
	accepts    mapset.Set
}

type DFARuntime struct {
	DFA          DeterministicFiniteAutomaton
	currentState mapset.Set
}

func (r *DFARuntime) doTransition(char string) {
	r.currentState = r.DFA.transition(r.currentState, char)
}

func (r DFARuntime) isAcceptState() bool {
	return r.currentState.Intersect(r.DFA.accepts).Cardinality() > 0
}

func (r DFARuntime) doesAccept(input string) bool {
	for _, rune := range input {
		r.doTransition(string(rune))
	}
	return r.isAcceptState()
}

func (d DeterministicFiniteAutomaton) getRuntime() DFARuntime {
	return DFARuntime{
		currentState: d.start,
		DFA:          d,
	}
}

type Token struct {
	value string
	kind  int
}

type Lexer struct {
	stringArray []string
	index       int
}

const (
	CHARACTER = 0
	OPE_UNION = 1
	OPE_STAR  = 2
	LPAREN    = 3
	RPAREN    = 4
	EOF       = 5
)

type Parser struct {
	lexer Lexer
	look  Token
}

func (l *Lexer) scan() Token {

	if l.index == len(l.stringArray) {
		return Token{
			value: "",
			kind:  EOF,
		}
	}

	ch := l.stringArray[l.index]
	l.index++

	if ch == `\` {
		l.index++
		return Token{
			value: l.stringArray[l.index-1],
			kind:  CHARACTER,
		}
	}
	if ch == "|" {
		return Token{
			value: ch,
			kind:  OPE_UNION,
		}
	}
	if ch == "(" {
		return Token{
			value: "(",
			kind:  LPAREN,
		}
	}
	if ch == ")" {
		return Token{
			value: ")",
			kind:  RPAREN,
		}
	}
	if ch == "*" {
		return Token{
			value: ch,
			kind:  OPE_STAR,
		}
	}
	return Token{
		value: ch,
		kind:  CHARACTER,
	}
}

func (p *Parser) match(tag int) error {
	if p.look.kind != tag {
		return errors.New("syntax error")
	}
	p.move()

	return nil
}

func (p *Parser) move() {
	p.look = p.lexer.scan()
}

func (p *Parser) factor() Node {
	if p.look.kind == LPAREN {
		// factor -> '(' subexpr ')'
		p.match(LPAREN)
		node := p.subexpr()
		p.match(RPAREN)
		return node
	} else {
		// factor -> CHARACTER
		node := Character{
			char: p.look.value,
		}
		p.match(CHARACTER)
		return node
	}
}

func (p *Parser) star() Node {
	// tar -> factor '*' | factor
	node := p.factor()
	if p.look.kind == OPE_STAR {
		p.match(OPE_STAR)
		node = Star{
			operand: node,
		}
	}

	return node
}

func (p *Parser) seq() Node {
	if p.look.kind == LPAREN || p.look.kind == CHARACTER {
		// seq -> subseq
		return p.subseq()
	} else {
		// seq -> ''
		return Character{
			char: "",
		}
	}
}

func (p *Parser) subseq() Node {
	node1 := p.star()
	if p.look.kind == LPAREN || p.look.kind == CHARACTER {
		// subseq -> star subseq
		node2 := p.subseq()
		node := Concat{
			operand1: node1,
			operand2: node2,
		}
		return node
	} else {
		// subseq -> star
		return node1
	}
}

func (p *Parser) subexpr() Node {
	// subexpr    -> seq '|' subexpr | seq
	node := p.seq()
	if p.look.kind == OPE_UNION {
		p.match(OPE_UNION)
		node2 := p.subexpr()
		node = Union{
			operand1: node,
			operand2: node2,
		}
	}
	return node
}

func (p *Parser) expression() NondeterministicFiniteAutomaton {
	// expression -> subexpr EOF
	node := p.subexpr()
	p.match(EOF)

	context := Context{}
	fragment := node.assemble(&context)
	return fragment.build()
}

type Context struct {
	stateCount int
}

func (c *Context) newState() int {
	c.stateCount = c.stateCount + 1
	return c.stateCount
}

type stateChar struct {
	state int
	char  string
}

type NFAFragment struct {
	start        int
	accepts      mapset.Set
	stateCharMap map[stateChar]mapset.Set // (状態, 入力文字) => 次の状態
}

func (f *NFAFragment) connect(from int, char string, to int) {
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

func (f NFAFragment) newSkelton() NFAFragment {
	newFragment := NFAFragment{
		start:        0,
		accepts:      mapset.NewSet(),
		stateCharMap: map[stateChar]mapset.Set{},
	}

	newFragment.stateCharMap = f.stateCharMap

	return newFragment
}

func (f NFAFragment) or(frag NFAFragment) NFAFragment {
	newFrag := f.newSkelton()

	for k, v := range frag.stateCharMap {
		newFrag.stateCharMap[k] = v
	}

	return newFrag
}

func (f NFAFragment) build() NondeterministicFiniteAutomaton {
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

	return NondeterministicFiniteAutomaton{
		transition: transition,
		start:      f.start,
		accepts:    f.accepts,
	}
}

type Node interface {
	assemble(*Context) NFAFragment
}

type Character struct {
	char string
}

func (c Character) assemble(context *Context) NFAFragment {
	frag := NFAFragment{
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

type Union struct {
	operand1 Node
	operand2 Node
}

func (u Union) assemble(context *Context) NFAFragment {
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

type Star struct {
	operand Node
}

func (s Star) assemble(context *Context) NFAFragment {
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

type Concat struct {
	operand1 Node
	operand2 Node
}

func (c Concat) assemble(context *Context) NFAFragment {

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
	dfa    DeterministicFiniteAutomaton
}

func (r *regexp) compile() {
	lexer := Lexer{
		stringArray: strings.Split(r.regexp, ""),
		index:       0,
	}
	parser := Parser{
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
