package eval

import "gopkg.in/Knetic/govaluate.v2"

type EvaluableExpression interface {
	Evaluate(parameters map[string]interface{}) (interface{}, error)
	String() string // Returns the original expression used to create this StrExprEvaluator.
	Vars() []string // Returns an array representing the variables contained in this StrExprEvaluator.
}

func NewEvaluableExpression(expr string) (EvaluableExpression, error) {
	return govaluate.NewEvaluableExpression(expr)
}