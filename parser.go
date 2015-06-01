//go:generate peg -switch -inline parser.peg

package aver

import "strings"

type Value struct {
	funcName   string
	predicates string
}

type Validation struct {
	global   string
	left     Value
	op       string
	right    Value
	relative string
}

type state struct {
	currentPredicates string
	currentString     string
	currentValue      Value
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

func (s *state) SetResultOp(op string) {
	s.validation.op = op
}

func (s *state) StringValue(value string) {
	s.currentString = value
}

func (s *state) SetRelative() {
	s.validation.relative = s.currentString
}
