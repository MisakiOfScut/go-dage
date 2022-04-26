package core

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/MisakiOfScut/go-dage/internal/script"
	"github.com/MisakiOfScut/go-dage/internal/utils/eval"
	"sync"
)

func init() {
	gob.Register(map[string]interface{}{})
}

type dagParams interface {
	DoEval(expression eval.EvaluableExpression) (interface{}, error)
	GetParams() (map[string]interface{}, error)
	GetParamByName(name string) (interface{}, error)
	SetParams(name string, value interface{}) error
	Clear()
}

type defaultDagParams struct {
	params map[string]interface{}
	lock   sync.RWMutex
}

func newDagParams() *defaultDagParams {
	return &defaultDagParams{params: make(map[string]interface{}), lock: sync.RWMutex{}}
}

func (m *defaultDagParams) DoEval(expression eval.EvaluableExpression) (interface{}, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return expression.Evaluate(m.params)
}

func (m *defaultDagParams) GetParams() (map[string]interface{}, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)
	err := enc.Encode(m.params)
	if err != nil {
		return nil, err
	}
	var copyMap map[string]interface{}
	err = dec.Decode(&copyMap)
	if err != nil {
		return nil, err
	}
	return copyMap, nil
}

func (m *defaultDagParams) GetParamByName(name string) (interface{}, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if v, ok := m.params[name]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("can't find %s", name)
}

func (m *defaultDagParams) SetParams(name string, value interface{}) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.params[name] = value
	return nil
}

func (m *defaultDagParams) Clear() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.params = make(map[string]interface{})
}

type DAGContext struct {
	dagParams
	UserData interface{}
}

type Operator interface {
	Name() string  // return operator's name
	OnExecute(ctx *DAGContext) error  // processing
	Reset() Operator  // if it is able to reset then return itself, otherwise return a new Operator object
}

type NewOperatorFunction func() Operator

type OperatorManager interface {
	RegisterOperator(id string, f NewOperatorFunction)
	GetOperator(id string) Operator
}

type defaultOperatorManager struct {
	operators map[string]NewOperatorFunction
}

func NewDefaultOperatorManager() *defaultOperatorManager {
	oprMgr := &defaultOperatorManager{operators: make(map[string]NewOperatorFunction)}
	oprMgr.addPredefinedOpr()
	return oprMgr
}

func (m *defaultOperatorManager) addPredefinedOpr() {
	m.RegisterOperator(script.DAGE_EXPR_OPERATOR, func() Operator {
		o := new(DAGEExpressionOperator)
		return o
	})
}

// RegisterOperator add an operator object create function to opr manager.
// Attention: add a func with duplicated name will replace the previous one.
func (m *defaultOperatorManager) RegisterOperator(oprName string, f NewOperatorFunction) {
	m.operators[oprName] = f
}

func (m *defaultOperatorManager) GetOperator(oprName string) Operator {
	if v, ok := m.operators[oprName]; ok {
		return v()
	}
	return nil
}

type DAGEExpressionOperator struct {
}

func (p *DAGEExpressionOperator) Name() string {
	return script.DAGE_EXPR_OPERATOR
}

func (p *DAGEExpressionOperator) OnExecute(ctx *DAGContext) error {
	return nil
}

func (p *DAGEExpressionOperator) Reset() Operator {
	return p
}