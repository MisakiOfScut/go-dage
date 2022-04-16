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

type Processor interface {
	OnExecute(ctx *DAGContext) error
}

type Operator struct {
	Name      string
	Processor Processor
}

type OperatorManager interface {
	RegisterOperator(id string, operator *Operator)
	GetOperator(id string) *Operator
}

type defaultOperatorManager struct {
	operators map[string]*Operator
}

func NewDefaultOperatorManager() *defaultOperatorManager {
	oprMgr := &defaultOperatorManager{operators: make(map[string]*Operator)}
	oprMgr.addPredefinedOpr()
	return oprMgr
}

func (m *defaultOperatorManager) addPredefinedOpr() {
	m.RegisterOperator(script.DAGE_EXPR_OPERATOR, &Operator{
		Name:      script.DAGE_EXPR_OPERATOR,
		Processor: &DAGEExpressionProcessor{},
	})
}

// RegisterOperator add an operator object to opr manager.
// Attention: add an opr with duplicated name will replace the previous one.
func (m *defaultOperatorManager) RegisterOperator(oprName string, operator *Operator) {
	m.operators[oprName] = operator
}

func (m *defaultOperatorManager) GetOperator(oprName string) *Operator {
	if v, ok := m.operators[oprName]; ok {
		return v
	}
	return nil
}

type DAGEExpressionProcessor struct {
}

func (p *DAGEExpressionProcessor) OnExecute(ctx *DAGContext) error {
	return nil
}
