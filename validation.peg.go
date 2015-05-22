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
	ruleglobal_assignment
	rulevalidations
	ruleresult
	rulerange
	rulecomparison
	rulevalue
	rulecomparison_op
	rulefrom_function
	ruleassignment
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
	ruleAction7
	ruleAction8
	ruleAction9
	rulePegText
	ruleAction10
	ruleAction11
	ruleAction12
	ruleAction13
	ruleAction14
	ruleAction15
	ruleAction16
	ruleAction17
	ruleAction18
	ruleAction19

	rulePre_
	rule_In_
	rule_Suf
)

var rul3s = [...]string{
	"Unknown",
	"expression",
	"global_assignment",
	"validations",
	"result",
	"range",
	"comparison",
	"value",
	"comparison_op",
	"from_function",
	"assignment",
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
	"Action7",
	"Action8",
	"Action9",
	"PegText",
	"Action10",
	"Action11",
	"Action12",
	"Action13",
	"Action14",
	"Action15",
	"Action16",
	"Action17",
	"Action18",
	"Action19",

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
	rules  [38]func() bool
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
			p.EndGlobalAssignment()
		case ruleAction1:
			p.BeginValidation()
		case ruleAction2:
			p.BeginValidation()
		case ruleAction3:
			p.EndLeft()
		case ruleAction4:
			p.EndValidation()
		case ruleAction5:
			p.EndLowest()
		case ruleAction6:
			p.EndHighest()
		case ruleAction7:
			p.EndComparison()
		case ruleAction8:
			p.EndFunctionValue()
		case ruleAction9:
			p.EndLiteralValue()
		case ruleAction10:
			p.SetComparisonOp(buffer[begin:end])
		case ruleAction11:
			p.BeginFunctionValue()
		case ruleAction12:
			p.BeginAssignment()
		case ruleAction13:
			p.EndAssignment()
		case ruleAction14:
			p.SetNumeric()
		case ruleAction15:
			p.SetString()
		case ruleAction16:
			p.SetRelative()
		case ruleAction17:
			p.SetAny()
		case ruleAction18:
			p.StringValue(buffer[begin:end])
		case ruleAction19:
			p.NumericValue(buffer[begin:end])

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
		/* 0 expression <- <(global_assignment? validations !.)> */
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
						if !_rules[ruleassignment]() {
							goto l2
						}
					l5:
						{
							position6, tokenIndex6, depth6 := position, tokenIndex, depth
							if buffer[position] != rune(',') {
								goto l6
							}
							position++
							if !_rules[ruleassignment]() {
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
						add(ruleglobal_assignment, position4)
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
					if !_rules[ruleresult]() {
						goto l0
					}
				l10:
					{
						position11, tokenIndex11, depth11 := position, tokenIndex, depth
						if buffer[position] != rune('a') {
							goto l11
						}
						position++
						if buffer[position] != rune('n') {
							goto l11
						}
						position++
						if buffer[position] != rune('d') {
							goto l11
						}
						position++
						{
							add(ruleAction2, position)
						}
						if !_rules[ruleresult]() {
							goto l11
						}
						goto l10
					l11:
						position, tokenIndex, depth = position11, tokenIndex11, depth11
					}
					depth--
					add(rulevalidations, position8)
				}
				{
					position13, tokenIndex13, depth13 := position, tokenIndex, depth
					if !matchDot() {
						goto l13
					}
					goto l0
				l13:
					position, tokenIndex, depth = position13, tokenIndex13, depth13
				}
				depth--
				add(ruleexpression, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 global_assignment <- <('f' 'o' 'r' assignment (',' assignment)* Action0)> */
		nil,
		/* 2 validations <- <('e' 'x' 'p' 'e' 'c' 't' Action1 result ('a' 'n' 'd' Action2 result)*)> */
		nil,
		/* 3 result <- <(value Action3 (range / comparison) Action4)> */
		func() bool {
			position16, tokenIndex16, depth16 := position, tokenIndex, depth
			{
				position17 := position
				depth++
				if !_rules[rulevalue]() {
					goto l16
				}
				{
					add(ruleAction3, position)
				}
				{
					position19, tokenIndex19, depth19 := position, tokenIndex, depth
					{
						position21 := position
						depth++
						if buffer[position] != rune('i') {
							goto l20
						}
						position++
						if buffer[position] != rune('n') {
							goto l20
						}
						position++
						if !_rules[rulews]() {
							goto l20
						}
						if buffer[position] != rune('[') {
							goto l20
						}
						position++
						if !_rules[rulevalue]() {
							goto l20
						}
						{
							add(ruleAction5, position)
						}
						if buffer[position] != rune(',') {
							goto l20
						}
						position++
						if !_rules[rulevalue]() {
							goto l20
						}
						{
							add(ruleAction6, position)
						}
						if buffer[position] != rune(']') {
							goto l20
						}
						position++
						depth--
						add(rulerange, position21)
					}
					goto l19
				l20:
					position, tokenIndex, depth = position19, tokenIndex19, depth19
					{
						position24 := position
						depth++
						{
							position25 := position
							depth++
							if !_rules[rulews]() {
								goto l16
							}
							{
								position26 := position
								depth++
								{
									position27, tokenIndex27, depth27 := position, tokenIndex, depth
									if buffer[position] != rune('>') {
										goto l28
									}
									position++
									goto l27
								l28:
									position, tokenIndex, depth = position27, tokenIndex27, depth27
									if buffer[position] != rune('<') {
										goto l29
									}
									position++
									goto l27
								l29:
									position, tokenIndex, depth = position27, tokenIndex27, depth27
									{
										switch buffer[position] {
										case '!':
											if buffer[position] != rune('!') {
												goto l16
											}
											position++
											if buffer[position] != rune('=') {
												goto l16
											}
											position++
											break
										case '<':
											if buffer[position] != rune('<') {
												goto l16
											}
											position++
											if buffer[position] != rune('=') {
												goto l16
											}
											position++
											break
										case '>':
											if buffer[position] != rune('>') {
												goto l16
											}
											position++
											if buffer[position] != rune('=') {
												goto l16
											}
											position++
											break
										default:
											if buffer[position] != rune('=') {
												goto l16
											}
											position++
											break
										}
									}

								}
							l27:
								depth--
								add(rulePegText, position26)
							}
							{
								add(ruleAction10, position)
							}
							depth--
							add(rulecomparison_op, position25)
						}
						if !_rules[rulevalue]() {
							goto l16
						}
						{
							add(ruleAction7, position)
						}
						depth--
						add(rulecomparison, position24)
					}
				}
			l19:
				{
					add(ruleAction4, position)
				}
				depth--
				add(ruleresult, position17)
			}
			return true
		l16:
			position, tokenIndex, depth = position16, tokenIndex16, depth16
			return false
		},
		/* 4 range <- <('i' 'n' ws '[' value Action5 ',' value Action6 ']')> */
		nil,
		/* 5 comparison <- <(comparison_op value Action7)> */
		nil,
		/* 6 value <- <((from_function Action8) / (literal Action9))> */
		func() bool {
			position36, tokenIndex36, depth36 := position, tokenIndex, depth
			{
				position37 := position
				depth++
				{
					position38, tokenIndex38, depth38 := position, tokenIndex, depth
					{
						position40 := position
						depth++
						if !_rules[rulestr]() {
							goto l39
						}
						{
							add(ruleAction11, position)
						}
						if buffer[position] != rune('(') {
							goto l39
						}
						position++
						if !_rules[ruleassignment]() {
							goto l39
						}
					l42:
						{
							position43, tokenIndex43, depth43 := position, tokenIndex, depth
							if buffer[position] != rune(',') {
								goto l43
							}
							position++
							if !_rules[ruleassignment]() {
								goto l43
							}
							goto l42
						l43:
							position, tokenIndex, depth = position43, tokenIndex43, depth43
						}
						if buffer[position] != rune(')') {
							goto l39
						}
						position++
						depth--
						add(rulefrom_function, position40)
					}
					{
						add(ruleAction8, position)
					}
					goto l38
				l39:
					position, tokenIndex, depth = position38, tokenIndex38, depth38
					if !_rules[ruleliteral]() {
						goto l36
					}
					{
						add(ruleAction9, position)
					}
				}
			l38:
				depth--
				add(rulevalue, position37)
			}
			return true
		l36:
			position, tokenIndex, depth = position36, tokenIndex36, depth36
			return false
		},
		/* 7 comparison_op <- <(ws <('>' / '<' / ((&('!') ('!' '=')) | (&('<') ('<' '=')) | (&('>') ('>' '=')) | (&('=') '=')))> Action10)> */
		nil,
		/* 8 from_function <- <(str Action11 '(' assignment (',' assignment)* ')')> */
		nil,
		/* 9 assignment <- <(str Action12 '=' literal Action13)> */
		func() bool {
			position48, tokenIndex48, depth48 := position, tokenIndex, depth
			{
				position49 := position
				depth++
				if !_rules[rulestr]() {
					goto l48
				}
				{
					add(ruleAction12, position)
				}
				if buffer[position] != rune('=') {
					goto l48
				}
				position++
				if !_rules[ruleliteral]() {
					goto l48
				}
				{
					add(ruleAction13, position)
				}
				depth--
				add(ruleassignment, position49)
			}
			return true
		l48:
			position, tokenIndex, depth = position48, tokenIndex48, depth48
			return false
		},
		/* 10 literal <- <((number Action14) / (str Action15) / (relative Action16) / (any Action17))> */
		func() bool {
			position52, tokenIndex52, depth52 := position, tokenIndex, depth
			{
				position53 := position
				depth++
				{
					position54, tokenIndex54, depth54 := position, tokenIndex, depth
					if !_rules[rulenumber]() {
						goto l55
					}
					{
						add(ruleAction14, position)
					}
					goto l54
				l55:
					position, tokenIndex, depth = position54, tokenIndex54, depth54
					if !_rules[rulestr]() {
						goto l57
					}
					{
						add(ruleAction15, position)
					}
					goto l54
				l57:
					position, tokenIndex, depth = position54, tokenIndex54, depth54
					{
						position60 := position
						depth++
						if !_rules[rulestr]() {
							goto l59
						}
						if buffer[position] != rune('*') {
							goto l59
						}
						position++
						if !_rules[rulenumber]() {
							goto l59
						}
						depth--
						add(rulerelative, position60)
					}
					{
						add(ruleAction16, position)
					}
					goto l54
				l59:
					position, tokenIndex, depth = position54, tokenIndex54, depth54
					{
						position62 := position
						depth++
						if !_rules[rulews]() {
							goto l52
						}
						{
							position63, tokenIndex63, depth63 := position, tokenIndex, depth
							if buffer[position] != rune('*') {
								goto l64
							}
							position++
							goto l63
						l64:
							position, tokenIndex, depth = position63, tokenIndex63, depth63
							if buffer[position] != rune('a') {
								goto l52
							}
							position++
							if buffer[position] != rune('n') {
								goto l52
							}
							position++
							if buffer[position] != rune('y') {
								goto l52
							}
							position++
						}
					l63:
						if !_rules[rulews]() {
							goto l52
						}
						depth--
						add(ruleany, position62)
					}
					{
						add(ruleAction17, position)
					}
				}
			l54:
				depth--
				add(ruleliteral, position53)
			}
			return true
		l52:
			position, tokenIndex, depth = position52, tokenIndex52, depth52
			return false
		},
		/* 11 any <- <(ws ('*' / ('a' 'n' 'y')) ws)> */
		nil,
		/* 12 relative <- <(str '*' number)> */
		nil,
		/* 13 str <- <(ws <(([a-z] / [A-Z]) ((&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))*)> ws Action18)> */
		func() bool {
			position68, tokenIndex68, depth68 := position, tokenIndex, depth
			{
				position69 := position
				depth++
				if !_rules[rulews]() {
					goto l68
				}
				{
					position70 := position
					depth++
					{
						position71, tokenIndex71, depth71 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l72
						}
						position++
						goto l71
					l72:
						position, tokenIndex, depth = position71, tokenIndex71, depth71
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l68
						}
						position++
					}
				l71:
				l73:
					{
						position74, tokenIndex74, depth74 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l74
								}
								position++
								break
							case '_':
								if buffer[position] != rune('_') {
									goto l74
								}
								position++
								break
							case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l74
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l74
								}
								position++
								break
							}
						}

						goto l73
					l74:
						position, tokenIndex, depth = position74, tokenIndex74, depth74
					}
					depth--
					add(rulePegText, position70)
				}
				if !_rules[rulews]() {
					goto l68
				}
				{
					add(ruleAction18, position)
				}
				depth--
				add(rulestr, position69)
			}
			return true
		l68:
			position, tokenIndex, depth = position68, tokenIndex68, depth68
			return false
		},
		/* 14 number <- <(ws <([0-9]+ ('.' [0-9]+)?)> ws Action19)> */
		func() bool {
			position77, tokenIndex77, depth77 := position, tokenIndex, depth
			{
				position78 := position
				depth++
				if !_rules[rulews]() {
					goto l77
				}
				{
					position79 := position
					depth++
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l77
					}
					position++
				l80:
					{
						position81, tokenIndex81, depth81 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l81
						}
						position++
						goto l80
					l81:
						position, tokenIndex, depth = position81, tokenIndex81, depth81
					}
					{
						position82, tokenIndex82, depth82 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l82
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l82
						}
						position++
					l84:
						{
							position85, tokenIndex85, depth85 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l85
							}
							position++
							goto l84
						l85:
							position, tokenIndex, depth = position85, tokenIndex85, depth85
						}
						goto l83
					l82:
						position, tokenIndex, depth = position82, tokenIndex82, depth82
					}
				l83:
					depth--
					add(rulePegText, position79)
				}
				if !_rules[rulews]() {
					goto l77
				}
				{
					add(ruleAction19, position)
				}
				depth--
				add(rulenumber, position78)
			}
			return true
		l77:
			position, tokenIndex, depth = position77, tokenIndex77, depth77
			return false
		},
		/* 15 ws <- <((&('\r') '\r') | (&('\n') '\n') | (&('\t') '\t') | (&(' ') ' '))*> */
		func() bool {
			{
				position88 := position
				depth++
			l89:
				{
					position90, tokenIndex90, depth90 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '\r':
							if buffer[position] != rune('\r') {
								goto l90
							}
							position++
							break
						case '\n':
							if buffer[position] != rune('\n') {
								goto l90
							}
							position++
							break
						case '\t':
							if buffer[position] != rune('\t') {
								goto l90
							}
							position++
							break
						default:
							if buffer[position] != rune(' ') {
								goto l90
							}
							position++
							break
						}
					}

					goto l89
				l90:
					position, tokenIndex, depth = position90, tokenIndex90, depth90
				}
				depth--
				add(rulews, position88)
			}
			return true
		},
		/* 17 Action0 <- <{ p.EndGlobalAssignment() }> */
		nil,
		/* 18 Action1 <- <{ p.BeginValidation() }> */
		nil,
		/* 19 Action2 <- <{ p.BeginValidation() }> */
		nil,
		/* 20 Action3 <- <{ p.EndLeft() }> */
		nil,
		/* 21 Action4 <- <{ p.EndValidation() }> */
		nil,
		/* 22 Action5 <- <{ p.EndLowest() }> */
		nil,
		/* 23 Action6 <- <{ p.EndHighest() }> */
		nil,
		/* 24 Action7 <- <{ p.EndComparison() }> */
		nil,
		/* 25 Action8 <- <{ p.EndFunctionValue() }> */
		nil,
		/* 26 Action9 <- <{ p.EndLiteralValue() }> */
		nil,
		nil,
		/* 28 Action10 <- <{ p.SetComparisonOp(buffer[begin:end]) }> */
		nil,
		/* 29 Action11 <- <{ p.BeginFunctionValue() }> */
		nil,
		/* 30 Action12 <- <{ p.BeginAssignment() }> */
		nil,
		/* 31 Action13 <- <{ p.EndAssignment() }> */
		nil,
		/* 32 Action14 <- <{ p.SetNumeric() }> */
		nil,
		/* 33 Action15 <- <{ p.SetString() }> */
		nil,
		/* 34 Action16 <- <{ p.SetRelative() }> */
		nil,
		/* 35 Action17 <- <{ p.SetAny() }> */
		nil,
		/* 36 Action18 <- <{ p.StringValue(buffer[begin:end]) }> */
		nil,
		/* 37 Action19 <- <{ p.NumericValue(buffer[begin:end]) }> */
		nil,
	}
	p.rules = _rules
}
