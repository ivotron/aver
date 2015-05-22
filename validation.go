//go:generate peg -switch -inline validation.peg

package aver

import "strconv"

type Assignment struct {
	name        string
	ofType      string
	literal     string
	relativeVar string
	relativeNum float64
	numeric     float64
}

type Value struct {
	funcName    string
	assignments []Assignment
}

type Validation struct {

	// - left is always present
	// - right is used for comparisons
	// - lowest/highest is used for ranges

	left       Value
	right      Value
	lowest     Value
	highest    Value
	comparison string
}

type state struct {
	functionPoints    map[string][]Assignment
	assignments       []Assignment
	currentAssignment Assignment
	currentString     string
	currentValue      Value
	currentValidation Validation
	comparisonOp      string
	validations       []Validation
}

func ParseValidation(input string) ([]Validation, error) {
	p := validationParser{Buffer: input}

	p.Init()
	p.state.Init()

	if err := p.Parse(); err != nil {
		return nil, err
	}

	p.Execute()

	return p.validations, nil
}

func (s *state) Init() {
	s.functionPoints = make(map[string][]Assignment)
	s.assignments = make([]Assignment, 0)
	s.validations = make([]Validation, 0)
}

func (s *state) EndLeft() {
	s.currentValidation.left = s.currentValue
}

func (s *state) BeginValidation() {
	s.currentValidation = Validation{}
}

func (s *state) EndValidation() {
	s.validations = append(s.validations, s.currentValidation)
}

func (s *state) SetComparisonOp(op string) {
	s.comparisonOp = op
}

func (s *state) EndLowest() {
	s.currentValidation.lowest = s.currentValue
}

func (s *state) EndComparison() {
	s.currentValidation.right = s.currentValue
	s.currentValidation.comparison = s.comparisonOp
}

func (s *state) EndHighest() {
	s.currentValidation.highest = s.currentValue
}

func (s *state) EndGlobalAssignment() {
	s.functionPoints["global"] = s.assignments
}

func (s *state) BeginFunctionValue() {
	s.currentValue = Value{funcName: s.currentString}
}

func (s *state) EndFunctionValue() {
	s.currentValue.assignments = s.assignments
}

func (s *state) EndLiteralValue() {
	s.currentValue = Value{funcName: "", assignments: s.assignments}
}

func (s *state) BeginAssignment() {
	s.currentAssignment = Assignment{name: s.currentString}
}

func (s *state) EndAssignment() {
	s.assignments = append(s.assignments, s.currentAssignment)
}

func (s *state) SetNumeric() {
	s.currentAssignment.ofType = "numeric"
}

func (s *state) SetString() {
	s.currentAssignment.ofType = "string"
	s.currentAssignment.literal = s.currentString
}

func (s *state) StringValue(value string) {
	s.currentString = value
}

func (s *state) NumericValue(value string) {
	num, err := strconv.ParseFloat(value, 64)

	if err != nil {
		panic("Conversion error")
	}

	s.currentAssignment.numeric = num
}

func (s *state) SetRelative() {
	s.currentAssignment.ofType = "relative"
	s.currentAssignment.relativeVar = s.currentAssignment.literal
	s.currentAssignment.relativeNum = s.currentAssignment.numeric
	s.currentAssignment.literal = ""
	s.currentAssignment.numeric = 0.0
}

func (s *state) SetAny() {
	s.currentAssignment.ofType = "any"
}
