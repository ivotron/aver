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
	ruleglobal_predicate
	rulevalidation
	ruleresult
	rulevalue
	ruleop
	rulefrom_function
	rulepredicate
	ruleliteral
	ruleany
	rulerelative
	rulestr
	rulenumber
	rulews
	ruleAction0
	ruleAction1
	ruleAction2
	ruleAction3
	ruleAction4
	ruleAction5
	ruleAction6
	rulePegText
	ruleAction7
	ruleAction8
	ruleAction9
	ruleAction10
	ruleAction11
	ruleAction12
	ruleAction13
	ruleAction14
	ruleAction15
	ruleAction16

	rulePre_
	rule_In_
	rule_Suf
)

var rul3s = [...]string{
	"Unknown",
	"expression",
	"global_predicate",
	"validation",
	"result",
	"value",
	"op",
	"from_function",
	"predicate",
	"literal",
	"any",
	"relative",
	"str",
	"number",
	"ws",
	"Action0",
	"Action1",
	"Action2",
	"Action3",
	"Action4",
	"Action5",
	"Action6",
	"PegText",
	"Action7",
	"Action8",
	"Action9",
	"Action10",
	"Action11",
	"Action12",
	"Action13",
	"Action14",
	"Action15",
	"Action16",

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
	rules  [33]func() bool
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
			p.EndGlobalPredicate()
		case ruleAction1:
			p.BeginValidation()
		case ruleAction2:
			p.EndLeft()
		case ruleAction3:
			p.SetResultOp()
		case ruleAction4:
			p.EndRight()
		case ruleAction5:
			p.EndFunctionValue()
		case ruleAction6:
			p.EndLiteralValue()
		case ruleAction7:
			p.SetComparisonOp(buffer[begin:end])
		case ruleAction8:
			p.BeginFunctionValue()
		case ruleAction9:
			p.BeginPredicate()
		case ruleAction10:
			p.EndPredicate()
		case ruleAction11:
			p.SetNumeric()
		case ruleAction12:
			p.SetString()
		case ruleAction13:
			p.SetAny()
		case ruleAction14:
			p.SetRelative()
		case ruleAction15:
			p.StringValue(buffer[begin:end])
		case ruleAction16:
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
		/* 0 expression <- <(global_predicate? validation !.)> */
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
						if !_rules[rulepredicate]() {
							goto l2
						}
					l5:
						{
							position6, tokenIndex6, depth6 := position, tokenIndex, depth
							if buffer[position] != rune(',') {
								goto l6
							}
							position++
							if !_rules[rulepredicate]() {
								goto l6
							}
							goto l5
						l6:
							position, tokenIndex, depth = position6, tokenIndex6, depth6
						}
						{
							add(ruleAction0, position)
						}
						depth--
						add(ruleglobal_predicate, position4)
					}
					goto l3
				l2:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
				}
			l3:
				{
					position8 := position
					depth++
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
						add(ruleAction1, position)
					}
					{
						position10 := position
						depth++
						if !_rules[rulevalue]() {
							goto l0
						}
						{
							add(ruleAction2, position)
						}
						if !_rules[ruleop]() {
							goto l0
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
							position14, tokenIndex14, depth14 := position, tokenIndex, depth
							{
								position16 := position
								depth++
								if !_rules[rulews]() {
									goto l14
								}
								if buffer[position] != rune('*') {
									goto l14
								}
								position++
								if !_rules[rulenumber]() {
									goto l14
								}
								{
									add(ruleAction14, position)
								}
								depth--
								add(rulerelative, position16)
							}
							goto l15
						l14:
							position, tokenIndex, depth = position14, tokenIndex14, depth14
						}
					l15:
						depth--
						add(ruleresult, position10)
					}
					depth--
					add(rulevalidation, position8)
				}
				{
					position18, tokenIndex18, depth18 := position, tokenIndex, depth
					if !matchDot() {
						goto l18
					}
					goto l0
				l18:
					position, tokenIndex, depth = position18, tokenIndex18, depth18
				}
				depth--
				add(ruleexpression, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 global_predicate <- <('f' 'o' 'r' predicate (',' predicate)* Action0)> */
		nil,
		/* 2 validation <- <('e' 'x' 'p' 'e' 'c' 't' Action1 result)> */
		nil,
		/* 3 result <- <(value Action2 op Action3 value Action4 relative?)> */
		nil,
		/* 4 value <- <((from_function Action5) / (literal Action6))> */
		func() bool {
			position22, tokenIndex22, depth22 := position, tokenIndex, depth
			{
				position23 := position
				depth++
				{
					position24, tokenIndex24, depth24 := position, tokenIndex, depth
					{
						position26 := position
						depth++
						if !_rules[rulestr]() {
							goto l25
						}
						if !_rules[rulews]() {
							goto l25
						}
						{
							add(ruleAction8, position)
						}
						if buffer[position] != rune('(') {
							goto l25
						}
						position++
						if !_rules[rulepredicate]() {
							goto l25
						}
					l28:
						{
							position29, tokenIndex29, depth29 := position, tokenIndex, depth
							if buffer[position] != rune(',') {
								goto l29
							}
							position++
							if !_rules[rulepredicate]() {
								goto l29
							}
							goto l28
						l29:
							position, tokenIndex, depth = position29, tokenIndex29, depth29
						}
						if buffer[position] != rune(')') {
							goto l25
						}
						position++
						depth--
						add(rulefrom_function, position26)
					}
					{
						add(ruleAction5, position)
					}
					goto l24
				l25:
					position, tokenIndex, depth = position24, tokenIndex24, depth24
					if !_rules[ruleliteral]() {
						goto l22
					}
					{
						add(ruleAction6, position)
					}
				}
			l24:
				depth--
				add(rulevalue, position23)
			}
			return true
		l22:
			position, tokenIndex, depth = position22, tokenIndex22, depth22
			return false
		},
		/* 5 op <- <(ws <('>' / '<' / ((&('!') ('!' '=')) | (&('<') ('<' '=')) | (&('>') ('>' '=')) | (&('=') '=')))> Action7)> */
		func() bool {
			position32, tokenIndex32, depth32 := position, tokenIndex, depth
			{
				position33 := position
				depth++
				if !_rules[rulews]() {
					goto l32
				}
				{
					position34 := position
					depth++
					{
						position35, tokenIndex35, depth35 := position, tokenIndex, depth
						if buffer[position] != rune('>') {
							goto l36
						}
						position++
						goto l35
					l36:
						position, tokenIndex, depth = position35, tokenIndex35, depth35
						if buffer[position] != rune('<') {
							goto l37
						}
						position++
						goto l35
					l37:
						position, tokenIndex, depth = position35, tokenIndex35, depth35
						{
							switch buffer[position] {
							case '!':
								if buffer[position] != rune('!') {
									goto l32
								}
								position++
								if buffer[position] != rune('=') {
									goto l32
								}
								position++
								break
							case '<':
								if buffer[position] != rune('<') {
									goto l32
								}
								position++
								if buffer[position] != rune('=') {
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
				l35:
					depth--
					add(rulePegText, position34)
				}
				{
					add(ruleAction7, position)
				}
				depth--
				add(ruleop, position33)
			}
			return true
		l32:
			position, tokenIndex, depth = position32, tokenIndex32, depth32
			return false
		},
		/* 6 from_function <- <(str ws Action8 '(' predicate (',' predicate)* ')')> */
		nil,
		/* 7 predicate <- <(str Action9 op literal Action10)> */
		func() bool {
			position41, tokenIndex41, depth41 := position, tokenIndex, depth
			{
				position42 := position
				depth++
				if !_rules[rulestr]() {
					goto l41
				}
				{
					add(ruleAction9, position)
				}
				if !_rules[ruleop]() {
					goto l41
				}
				if !_rules[ruleliteral]() {
					goto l41
				}
				{
					add(ruleAction10, position)
				}
				depth--
				add(rulepredicate, position42)
			}
			return true
		l41:
			position, tokenIndex, depth = position41, tokenIndex41, depth41
			return false
		},
		/* 8 literal <- <((number Action11) / (str Action12) / (any Action13))> */
		func() bool {
			position45, tokenIndex45, depth45 := position, tokenIndex, depth
			{
				position46 := position
				depth++
				{
					position47, tokenIndex47, depth47 := position, tokenIndex, depth
					if !_rules[rulenumber]() {
						goto l48
					}
					{
						add(ruleAction11, position)
					}
					goto l47
				l48:
					position, tokenIndex, depth = position47, tokenIndex47, depth47
					if !_rules[rulestr]() {
						goto l50
					}
					{
						add(ruleAction12, position)
					}
					goto l47
				l50:
					position, tokenIndex, depth = position47, tokenIndex47, depth47
					{
						position52 := position
						depth++
						if !_rules[rulews]() {
							goto l45
						}
						{
							position53, tokenIndex53, depth53 := position, tokenIndex, depth
							if buffer[position] != rune('*') {
								goto l54
							}
							position++
							goto l53
						l54:
							position, tokenIndex, depth = position53, tokenIndex53, depth53
							if buffer[position] != rune('a') {
								goto l45
							}
							position++
							if buffer[position] != rune('n') {
								goto l45
							}
							position++
							if buffer[position] != rune('y') {
								goto l45
							}
							position++
						}
					l53:
						if !_rules[rulews]() {
							goto l45
						}
						depth--
						add(ruleany, position52)
					}
					{
						add(ruleAction13, position)
					}
				}
			l47:
				depth--
				add(ruleliteral, position46)
			}
			return true
		l45:
			position, tokenIndex, depth = position45, tokenIndex45, depth45
			return false
		},
		/* 9 any <- <(ws ('*' / ('a' 'n' 'y')) ws)> */
		nil,
		/* 10 relative <- <(ws '*' number Action14)> */
		nil,
		/* 11 str <- <(ws <(([a-z] / [A-Z]) ((&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))*)> ws Action15)> */
		func() bool {
			position58, tokenIndex58, depth58 := position, tokenIndex, depth
			{
				position59 := position
				depth++
				if !_rules[rulews]() {
					goto l58
				}
				{
					position60 := position
					depth++
					{
						position61, tokenIndex61, depth61 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l62
						}
						position++
						goto l61
					l62:
						position, tokenIndex, depth = position61, tokenIndex61, depth61
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l58
						}
						position++
					}
				l61:
				l63:
					{
						position64, tokenIndex64, depth64 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l64
								}
								position++
								break
							case '_':
								if buffer[position] != rune('_') {
									goto l64
								}
								position++
								break
							case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l64
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l64
								}
								position++
								break
							}
						}

						goto l63
					l64:
						position, tokenIndex, depth = position64, tokenIndex64, depth64
					}
					depth--
					add(rulePegText, position60)
				}
				if !_rules[rulews]() {
					goto l58
				}
				{
					add(ruleAction15, position)
				}
				depth--
				add(rulestr, position59)
			}
			return true
		l58:
			position, tokenIndex, depth = position58, tokenIndex58, depth58
			return false
		},
		/* 12 number <- <(ws <([0-9]+ ('.' [0-9]+)?)> ws Action16)> */
		func() bool {
			position67, tokenIndex67, depth67 := position, tokenIndex, depth
			{
				position68 := position
				depth++
				if !_rules[rulews]() {
					goto l67
				}
				{
					position69 := position
					depth++
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l67
					}
					position++
				l70:
					{
						position71, tokenIndex71, depth71 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l71
						}
						position++
						goto l70
					l71:
						position, tokenIndex, depth = position71, tokenIndex71, depth71
					}
					{
						position72, tokenIndex72, depth72 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l72
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l72
						}
						position++
					l74:
						{
							position75, tokenIndex75, depth75 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l75
							}
							position++
							goto l74
						l75:
							position, tokenIndex, depth = position75, tokenIndex75, depth75
						}
						goto l73
					l72:
						position, tokenIndex, depth = position72, tokenIndex72, depth72
					}
				l73:
					depth--
					add(rulePegText, position69)
				}
				if !_rules[rulews]() {
					goto l67
				}
				{
					add(ruleAction16, position)
				}
				depth--
				add(rulenumber, position68)
			}
			return true
		l67:
			position, tokenIndex, depth = position67, tokenIndex67, depth67
			return false
		},
		/* 13 ws <- <((&('\r') '\r') | (&('\n') '\n') | (&('\t') '\t') | (&(' ') ' '))*> */
		func() bool {
			{
				position78 := position
				depth++
			l79:
				{
					position80, tokenIndex80, depth80 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '\r':
							if buffer[position] != rune('\r') {
								goto l80
							}
							position++
							break
						case '\n':
							if buffer[position] != rune('\n') {
								goto l80
							}
							position++
							break
						case '\t':
							if buffer[position] != rune('\t') {
								goto l80
							}
							position++
							break
						default:
							if buffer[position] != rune(' ') {
								goto l80
							}
							position++
							break
						}
					}

					goto l79
				l80:
					position, tokenIndex, depth = position80, tokenIndex80, depth80
				}
				depth--
				add(rulews, position78)
			}
			return true
		},
		/* 15 Action0 <- <{ p.EndGlobalPredicate() }> */
		nil,
		/* 16 Action1 <- <{ p.BeginValidation() }> */
		nil,
		/* 17 Action2 <- <{ p.EndLeft() }> */
		nil,
		/* 18 Action3 <- <{ p.SetResultOp() }> */
		nil,
		/* 19 Action4 <- <{ p.EndRight() }> */
		nil,
		/* 20 Action5 <- <{ p.EndFunctionValue() }> */
		nil,
		/* 21 Action6 <- <{ p.EndLiteralValue() }> */
		nil,
		nil,
		/* 23 Action7 <- <{ p.SetComparisonOp(buffer[begin:end]) }> */
		nil,
		/* 24 Action8 <- <{ p.BeginFunctionValue() }> */
		nil,
		/* 25 Action9 <- <{ p.BeginPredicate() }> */
		nil,
		/* 26 Action10 <- <{ p.EndPredicate() }> */
		nil,
		/* 27 Action11 <- <{ p.SetNumeric() }> */
		nil,
		/* 28 Action12 <- <{ p.SetString() }> */
		nil,
		/* 29 Action13 <- <{ p.SetAny() }> */
		nil,
		/* 30 Action14 <- <{ p.SetRelative() }> */
		nil,
		/* 31 Action15 <- <{ p.StringValue(buffer[begin:end]) }> */
		nil,
		/* 32 Action16 <- <{ p.StringValue(buffer[begin:end]) }> */
		nil,
	}
	p.rules = _rules
}
