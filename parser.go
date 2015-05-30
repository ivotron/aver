//go:generate peg -switch -inline parser.peg

package aver

import "strings"

type Op int

const (
	NONE Op = iota

	EQL // =
	LSS // <
	GTR // >
	NOT // <>
	LEQ // <=
	GEQ // >=
)

type Value struct {
	funcName   string
	predicates string
}

type Validation struct {
	global   string
	left     Value
	op       Op
	right    Value
	relative string
}

type state struct {
	currentPredicates string
	currentString     string
	currentValue      Value
	currentOperator   Op
	validation        Validation
}

func ParseValidation(input string) (v Validation, e error) {
	p := validationParser{Buffer: input}

	p.Init()

	if e = p.Parse(); e != nil {
		return
	}

	p.Execute()

	return p.validation, nil
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
	case "<>":
		s.currentOperator = NOT
	case "<=":
		s.currentOperator = LEQ
	case ">=":
		s.currentOperator = GEQ
	}
}

func (s *state) EndGlobalPredicates() {
	s.validation.global = s.currentPredicates
}

func (s *state) BeginFunctionValue() {
	s.currentValue = Value{funcName: s.currentString}
}

func (s *state) EndFunctionValue() {
	s.currentValue.predicates = s.currentPredicates
}

func (s *state) SetPredicates(predicates string) {
	s.currentPredicates = strings.TrimSpace(predicates)
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

func (s *state) StringValue(value string) {
	s.currentString = value
}

func (s *state) SetRelative() {
	s.validation.relative = s.currentString
}
