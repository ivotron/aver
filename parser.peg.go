package aver

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const end_symbol rune = 4

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleexpression
	ruleglobal_predicates
	rulepredicates
	rulevalidation
	ruleresult
	rulevalue
	ruleop
	rulepredicate
	ruleliteral
	rulerelative
	rulestr
	rulenumber
	rulews
	ruleAction0
	rulePegText
	ruleAction1
	ruleAction2
	ruleAction3
	ruleAction4
	ruleAction5
	ruleAction6
	ruleAction7
	ruleAction8
	ruleAction9

	rulePre_
	rule_In_
	rule_Suf
)

var rul3s = [...]string{
	"Unknown",
	"expression",
	"global_predicates",
	"predicates",
	"validation",
	"result",
	"value",
	"op",
	"predicate",
	"literal",
	"relative",
	"str",
	"number",
	"ws",
	"Action0",
	"PegText",
	"Action1",
	"Action2",
	"Action3",
	"Action4",
	"Action5",
	"Action6",
	"Action7",
	"Action8",
	"Action9",

	"Pre_",
	"_In_",
	"_Suf",
}

type tokenTree interface {
	Print()
	PrintSyntax()
	PrintSyntaxTree(buffer string)
	Add(rule pegRule, begin, end, next uint32, depth int)
	Expand(index int) tokenTree
	Tokens() <-chan token32
	AST() *node32
	Error() []token32
	trim(length int)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(depth int, buffer string) {
	for node != nil {
		for c := 0; c < depth; c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[node.pegRule], strconv.Quote(string(([]rune(buffer)[node.begin:node.end]))))
		if node.up != nil {
			node.up.print(depth+1, buffer)
		}
		node = node.next
	}
}

func (ast *node32) Print(buffer string) {
	ast.print(0, buffer)
}

type element struct {
	node *node32
	down *element
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	pegRule
	begin, end, next uint32
}

func (t *token32) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: uint32(t.begin), end: uint32(t.end), next: uint32(t.next)}
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
}

type tokens32 struct {
	tree    []token32
	ordered [][]token32
}

func (t *tokens32) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) Order() [][]token32 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int32, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegRule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token32, len(depths)), make([]token32, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = uint32(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state32 struct {
	token32
	depths []int32
	leaf   bool
}

func (t *tokens32) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
}

func (t *tokens32) PreOrder() (<-chan state32, [][]token32) {
	s, ordered := make(chan state32, 6), t.Order()
	go func() {
		var states [8]state32
		for i, _ := range states {
			states[i].depths = make([]int32, len(ordered))
		}
		depths, state, depth := make([]int32, len(ordered)), 0, 1
		write := func(t token32, leaf bool) {
			S := states[state]
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, uint32(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token32 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token32{pegRule: rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{pegRule: rulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegRule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegRule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{pegRule: rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens32) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(string(([]rune(buffer)[token.begin:token.end]))))
	}
}

func (t *tokens32) Add(rule pegRule, begin, end, depth uint32, index int) {
	t.tree[index] = token32{pegRule: rule, begin: uint32(begin), end: uint32(end), next: uint32(depth)}
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens32) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

/*func (t *tokens16) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2 * len(tree))
		for i, v := range tree {
			expanded[i] = v.getToken32()
		}
		return &tokens32{tree: expanded}
	}
	return nil
}*/

func (t *tokens32) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	return nil
}

type validationParser struct {
	state

	Buffer string
	buffer []rune
	rules  [25]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	tokenTree
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer string, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer[0:] {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p *validationParser
}

func (e *parseError) Error() string {
	tokens, error := e.p.tokenTree.Error(), "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.Buffer, positions)
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf("parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n",
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			/*strconv.Quote(*/ e.p.Buffer[begin:end] /*)*/)
	}

	return error
}

func (p *validationParser) PrintSyntaxTree() {
	p.tokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *validationParser) Highlighter() {
	p.tokenTree.PrintSyntax()
}

func (p *validationParser) Execute() {
	buffer, begin, end := p.Buffer, 0, 0
	for token := range p.tokenTree.Tokens() {
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)

		case ruleAction0:
			p.EndGlobalPredicates()
		case ruleAction1:
			p.SetPredicates(buffer[begin:end])
		case ruleAction2:
			p.EndLeft()
		case ruleAction3:
			p.SetResultOp(buffer[begin:end])
		case ruleAction4:
			p.EndRight()
		case ruleAction5:
			p.BeginFunctionValue()
		case ruleAction6:
			p.EndFunctionValue()
		case ruleAction7:
			p.SetRelative()
		case ruleAction8:
			p.StringValue(buffer[begin:end])
		case ruleAction9:
			p.StringValue(buffer[begin:end])

		}
	}
	_, _, _ = buffer, begin, end
}

func (p *validationParser) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != end_symbol {
		p.buffer = append(p.buffer, end_symbol)
	}

	var tree tokenTree = &tokens32{tree: make([]token32, math.MaxInt16)}
	position, depth, tokenIndex, buffer, _rules := uint32(0), uint32(0), 0, p.buffer, p.rules

	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokenTree = tree
		if matches {
			p.tokenTree.trim(tokenIndex)
			return nil
		}
		return &parseError{p}
	}

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
	}

	add := func(rule pegRule, begin uint32) {
		if t := tree.Expand(tokenIndex); t != nil {
			tree = t
		}
		tree.Add(rule, begin, position, depth, tokenIndex)
		tokenIndex++
	}

	matchDot := func() bool {
		if buffer[position] != end_symbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 expression <- <(global_predicates? validation !.)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				{
					position2, tokenIndex2, depth2 := position, tokenIndex, depth
					{
						position4 := position
						depth++
						if !_rules[rulews]() {
							goto l2
						}
						if buffer[position] != rune('f') {
							goto l2
						}
						position++
						if buffer[position] != rune('o') {
							goto l2
						}
						position++
						if buffer[position] != rune('r') {
							goto l2
						}
						position++
						if !_rules[rulepredicates]() {
							goto l2
						}
						{
							add(ruleAction0, position)
						}
						depth--
						add(ruleglobal_predicates, position4)
					}
					goto l3
				l2:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
				}
			l3:
				{
					position6 := position
					depth++
					if !_rules[rulews]() {
						goto l0
					}
					if buffer[position] != rune('e') {
						goto l0
					}
					position++
					if buffer[position] != rune('x') {
						goto l0
					}
					position++
					if buffer[position] != rune('p') {
						goto l0
					}
					position++
					if buffer[position] != rune('e') {
						goto l0
					}
					position++
					if buffer[position] != rune('c') {
						goto l0
					}
					position++
					if buffer[position] != rune('t') {
						goto l0
					}
					position++
					{
						position7 := position
						depth++
						if !_rules[rulevalue]() {
							goto l0
						}
						{
							add(ruleAction2, position)
						}
						{
							position9 := position
							depth++
							if !_rules[ruleop]() {
								goto l0
							}
							depth--
							add(rulePegText, position9)
						}
						{
							add(ruleAction3, position)
						}
						if !_rules[rulevalue]() {
							goto l0
						}
						{
							add(ruleAction4, position)
						}
						{
							position12, tokenIndex12, depth12 := position, tokenIndex, depth
							{
								position14 := position
								depth++
								if !_rules[rulews]() {
									goto l12
								}
								if buffer[position] != rune('*') {
									goto l12
								}
								position++
								if !_rules[rulenumber]() {
									goto l12
								}
								{
									add(ruleAction7, position)
								}
								depth--
								add(rulerelative, position14)
							}
							goto l13
						l12:
							position, tokenIndex, depth = position12, tokenIndex12, depth12
						}
					l13:
						depth--
						add(ruleresult, position7)
					}
					depth--
					add(rulevalidation, position6)
				}
				{
					position16, tokenIndex16, depth16 := position, tokenIndex, depth
					if !matchDot() {
						goto l16
					}
					goto l0
				l16:
					position, tokenIndex, depth = position16, tokenIndex16, depth16
				}
				depth--
				add(ruleexpression, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 global_predicates <- <(ws ('f' 'o' 'r') predicates Action0)> */
		nil,
		/* 2 predicates <- <(<(predicate ('a' 'n' 'd' predicate)*)> Action1)> */
		func() bool {
			position18, tokenIndex18, depth18 := position, tokenIndex, depth
			{
				position19 := position
				depth++
				{
					position20 := position
					depth++
					if !_rules[rulepredicate]() {
						goto l18
					}
				l21:
					{
						position22, tokenIndex22, depth22 := position, tokenIndex, depth
						if buffer[position] != rune('a') {
							goto l22
						}
						position++
						if buffer[position] != rune('n') {
							goto l22
						}
						position++
						if buffer[position] != rune('d') {
							goto l22
						}
						position++
						if !_rules[rulepredicate]() {
							goto l22
						}
						goto l21
					l22:
						position, tokenIndex, depth = position22, tokenIndex22, depth22
					}
					depth--
					add(rulePegText, position20)
				}
				{
					add(ruleAction1, position)
				}
				depth--
				add(rulepredicates, position19)
			}
			return true
		l18:
			position, tokenIndex, depth = position18, tokenIndex18, depth18
			return false
		},
		/* 3 validation <- <(ws ('e' 'x' 'p' 'e' 'c' 't') result)> */
		nil,
		/* 4 result <- <(value Action2 <op> Action3 value Action4 relative?)> */
		nil,
		/* 5 value <- <(str ws Action5 ('(' predicates ')' ws)? Action6)> */
		func() bool {
			position26, tokenIndex26, depth26 := position, tokenIndex, depth
			{
				position27 := position
				depth++
				if !_rules[rulestr]() {
					goto l26
				}
				if !_rules[rulews]() {
					goto l26
				}
				{
					add(ruleAction5, position)
				}
				{
					position29, tokenIndex29, depth29 := position, tokenIndex, depth
					if buffer[position] != rune('(') {
						goto l29
					}
					position++
					if !_rules[rulepredicates]() {
						goto l29
					}
					if buffer[position] != rune(')') {
						goto l29
					}
					position++
					if !_rules[rulews]() {
						goto l29
					}
					goto l30
				l29:
					position, tokenIndex, depth = position29, tokenIndex29, depth29
				}
			l30:
				{
					add(ruleAction6, position)
				}
				depth--
				add(rulevalue, position27)
			}
			return true
		l26:
			position, tokenIndex, depth = position26, tokenIndex26, depth26
			return false
		},
		/* 6 op <- <(ws ('>' / '<' / ('<' '=') / ((&('<') ('<' '>')) | (&('>') ('>' '=')) | (&('=') '='))))> */
		func() bool {
			position32, tokenIndex32, depth32 := position, tokenIndex, depth
			{
				position33 := position
				depth++
				if !_rules[rulews]() {
					goto l32
				}
				{
					position34, tokenIndex34, depth34 := position, tokenIndex, depth
					if buffer[position] != rune('>') {
						goto l35
					}
					position++
					goto l34
				l35:
					position, tokenIndex, depth = position34, tokenIndex34, depth34
					if buffer[position] != rune('<') {
						goto l36
					}
					position++
					goto l34
				l36:
					position, tokenIndex, depth = position34, tokenIndex34, depth34
					if buffer[position] != rune('<') {
						goto l37
					}
					position++
					if buffer[position] != rune('=') {
						goto l37
					}
					position++
					goto l34
				l37:
					position, tokenIndex, depth = position34, tokenIndex34, depth34
					{
						switch buffer[position] {
						case '<':
							if buffer[position] != rune('<') {
								goto l32
							}
							position++
							if buffer[position] != rune('>') {
								goto l32
							}
							position++
							break
						case '>':
							if buffer[position] != rune('>') {
								goto l32
							}
							position++
							if buffer[position] != rune('=') {
								goto l32
							}
							position++
							break
						default:
							if buffer[position] != rune('=') {
								goto l32
							}
							position++
							break
						}
					}

				}
			l34:
				depth--
				add(ruleop, position33)
			}
			return true
		l32:
			position, tokenIndex, depth = position32, tokenIndex32, depth32
			return false
		},
		/* 7 predicate <- <(str op literal)> */
		func() bool {
			position39, tokenIndex39, depth39 := position, tokenIndex, depth
			{
				position40 := position
				depth++
				if !_rules[rulestr]() {
					goto l39
				}
				if !_rules[ruleop]() {
					goto l39
				}
				{
					position41 := position
					depth++
					if !_rules[rulews]() {
						goto l39
					}
					{
						position42, tokenIndex42, depth42 := position, tokenIndex, depth
						if !_rules[rulenumber]() {
							goto l43
						}
						goto l42
					l43:
						position, tokenIndex, depth = position42, tokenIndex42, depth42
						if buffer[position] != rune('\'') {
							goto l39
						}
						position++
						if !_rules[rulestr]() {
							goto l39
						}
						if buffer[position] != rune('\'') {
							goto l39
						}
						position++
					}
				l42:
					if !_rules[rulews]() {
						goto l39
					}
					depth--
					add(ruleliteral, position41)
				}
				depth--
				add(rulepredicate, position40)
			}
			return true
		l39:
			position, tokenIndex, depth = position39, tokenIndex39, depth39
			return false
		},
		/* 8 literal <- <(ws (number / ('\'' str '\'')) ws)> */
		nil,
		/* 9 relative <- <(ws '*' number Action7)> */
		nil,
		/* 10 str <- <(ws <(([a-z] / [A-Z]) ((&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))*)> ws Action8)> */
		func() bool {
			position46, tokenIndex46, depth46 := position, tokenIndex, depth
			{
				position47 := position
				depth++
				if !_rules[rulews]() {
					goto l46
				}
				{
					position48 := position
					depth++
					{
						position49, tokenIndex49, depth49 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l50
						}
						position++
						goto l49
					l50:
						position, tokenIndex, depth = position49, tokenIndex49, depth49
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l46
						}
						position++
					}
				l49:
				l51:
					{
						position52, tokenIndex52, depth52 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l52
								}
								position++
								break
							case '_':
								if buffer[position] != rune('_') {
									goto l52
								}
								position++
								break
							case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l52
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l52
								}
								position++
								break
							}
						}

						goto l51
					l52:
						position, tokenIndex, depth = position52, tokenIndex52, depth52
					}
					depth--
					add(rulePegText, position48)
				}
				if !_rules[rulews]() {
					goto l46
				}
				{
					add(ruleAction8, position)
				}
				depth--
				add(rulestr, position47)
			}
			return true
		l46:
			position, tokenIndex, depth = position46, tokenIndex46, depth46
			return false
		},
		/* 11 number <- <(ws <([0-9]+ ('.' [0-9]+)?)> ws Action9)> */
		func() bool {
			position55, tokenIndex55, depth55 := position, tokenIndex, depth
			{
				position56 := position
				depth++
				if !_rules[rulews]() {
					goto l55
				}
				{
					position57 := position
					depth++
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l55
					}
					position++
				l58:
					{
						position59, tokenIndex59, depth59 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l59
						}
						position++
						goto l58
					l59:
						position, tokenIndex, depth = position59, tokenIndex59, depth59
					}
					{
						position60, tokenIndex60, depth60 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l60
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l60
						}
						position++
					l62:
						{
							position63, tokenIndex63, depth63 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l63
							}
							position++
							goto l62
						l63:
							position, tokenIndex, depth = position63, tokenIndex63, depth63
						}
						goto l61
					l60:
						position, tokenIndex, depth = position60, tokenIndex60, depth60
					}
				l61:
					depth--
					add(rulePegText, position57)
				}
				if !_rules[rulews]() {
					goto l55
				}
				{
					add(ruleAction9, position)
				}
				depth--
				add(rulenumber, position56)
			}
			return true
		l55:
			position, tokenIndex, depth = position55, tokenIndex55, depth55
			return false
		},
		/* 12 ws <- <((&('\r') '\r') | (&('\n') '\n') | (&('\t') '\t') | (&(' ') ' '))*> */
		func() bool {
			{
				position66 := position
				depth++
			l67:
				{
					position68, tokenIndex68, depth68 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '\r':
							if buffer[position] != rune('\r') {
								goto l68
							}
							position++
							break
						case '\n':
							if buffer[position] != rune('\n') {
								goto l68
							}
							position++
							break
						case '\t':
							if buffer[position] != rune('\t') {
								goto l68
							}
							position++
							break
						default:
							if buffer[position] != rune(' ') {
								goto l68
							}
							position++
							break
						}
					}

					goto l67
				l68:
					position, tokenIndex, depth = position68, tokenIndex68, depth68
				}
				depth--
				add(rulews, position66)
			}
			return true
		},
		/* 14 Action0 <- <{ p.EndGlobalPredicates() }> */
		nil,
		nil,
		/* 16 Action1 <- <{ p.SetPredicates(buffer[begin:end]) }> */
		nil,
		/* 17 Action2 <- <{ p.EndLeft() }> */
		nil,
		/* 18 Action3 <- <{ p.SetResultOp(buffer[begin:end]) }> */
		nil,
		/* 19 Action4 <- <{ p.EndRight() }> */
		nil,
		/* 20 Action5 <- <{ p.BeginFunctionValue() }> */
		nil,
		/* 21 Action6 <- <{ p.EndFunctionValue() }> */
		nil,
		/* 22 Action7 <- <{ p.SetRelative() }> */
		nil,
		/* 23 Action8 <- <{ p.StringValue(buffer[begin:end]) }> */
		nil,
		/* 24 Action9 <- <{ p.StringValue(buffer[begin:end]) }> */
		nil,
	}
	p.rules = _rules
}
