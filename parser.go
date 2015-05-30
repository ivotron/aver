//go:generate peg -switch -inline parser.peg

package aver

import "strconv"

type Op int

const (
	NONE Op = iota

	EQL // =
	LSS // <
	GTR // >
	NOT // !
	NEQ // !=
	LEQ // <=
	GEQ // >=
	ANY // *
)

type Predicate struct {
	variable string
	op       Op
	value    interface{}
}

type Value struct {
	funcName   string
	predicates []Predicate
}

type Validation struct {
	global   []Predicate
	left     Value
	op       Op
	right    Value
	relative float64
}

type state struct {
	currentPredicates []Predicate
	currentPredicate  Predicate
	currentString     string
	currentValue      Value
	currentOperator   Op
	validation        Validation
}

func ParseValidation(input string) (v Validation, e error) {
	p := validationParser{Buffer: input}

	p.Init()
	p.state.Init()

	if e = p.Parse(); e != nil {
		return
	}

	p.Execute()

	return p.state.validation, nil
}

func (s *state) Init() {
	s.currentPredicates = make([]Predicate, 0)
}

func (s *state) BeginValidation() {
	s.validation = Validation{}
}

func (s *state) SetComparisonOp(op string) {
	switch op {
	default:
		s.currentOperator = NONE
	case "=":
		s.currentOperator = EQL
	case "<":
		s.currentOperator = LSS
	case ">":
		s.currentOperator = GTR
	case "!=":
		s.currentOperator = NOT
	case "<=":
		s.currentOperator = LEQ
	case ">=":
		s.currentOperator = GEQ
	}
}

func (s *state) EndGlobalPredicate() {
	s.validation.global = s.currentPredicates
}

func (s *state) BeginFunctionValue() {
	s.currentValue = Value{funcName: s.currentString}
}

func (s *state) EndFunctionValue() {
	s.currentValue.predicates = s.currentPredicates
	s.currentPredicates = make([]Predicate, 0)
}

func (s *state) EndLiteralValue() {
	s.currentValue = Value{funcName: "", predicates: s.currentPredicates}
}

func (s *state) BeginPredicate() {
	s.currentPredicate = Predicate{variable: s.currentString}
}

func (s *state) EndPredicate() {
	s.currentPredicate.op = s.currentOperator
	s.currentPredicates = append(s.currentPredicates, s.currentPredicate)
}

func (s *state) EndLeft() {
	s.validation.left = s.currentValue
}

func (s *state) EndRight() {
	s.validation.right = s.currentValue
}

func (s *state) SetResultOp() {
	s.validation.op = s.currentOperator
}

func (s *state) SetNumeric() {
	s.currentPredicate.value = toNumber(s.currentString)
}

func toNumber(value string) float64 {
	num, err := strconv.ParseFloat(value, 64)

	if err != nil {
		panic("Conversion error")
	}

	return num
}

func (s *state) SetString() {
	s.currentPredicate.value = s.currentString
}

func (s *state) StringValue(value string) {
	s.currentString = value
}

func (s *state) SetRelative() {
	s.validation.relative = toNumber(s.currentString)
}

func (s *state) SetAny() {
	s.currentPredicate.op = ANY
}
