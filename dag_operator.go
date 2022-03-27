package dage

const DAGE_EXPR_OPERATOR string = "__DAGE_EXPR_OPERATOR__"

type Context interface {
	getParams() map[string]interface{}
	getParamByName(name string) (interface{}, error)
	setParams(name string, value interface{}) error
}

type Processor interface {
	OnExecute(ctx *Context) error
}

type Operator struct {
	Name      string
	Processor Processor
}
