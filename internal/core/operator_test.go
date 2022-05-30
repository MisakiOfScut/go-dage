package core

import (
	"fmt"
	"github.com/MisakiOfScut/go-dage/internal/utils/log"
	"testing"
	"time"
)

var tOprMgr = NewDefaultOperatorManager()

type timeOpr struct {
	name   string
	number int
}

func (t *timeOpr) Name() string {
	return t.name
}
func (t *timeOpr) OnExecute(ctx *DAGContext) (map[string]interface{}, error) {
	time.Sleep(time.Duration(t.number) * time.Millisecond)
	return nil, nil
}
func (t *timeOpr) InjectDepsData(key string, value interface{}) error {
	return nil
}
func (t *timeOpr) GetInputsID() []string {
	return nil
}
func (t *timeOpr) GetOutputsID() []string {
	return nil
}
func (t *timeOpr) Reset() Operator {
	return t
}

// nonOp opr, used for benchmark
type nonOp struct {
	name string
}

func (t *nonOp) Name() string {
	return t.name
}
func (t *nonOp) OnExecute(ctx *DAGContext) (map[string]interface{}, error) {
	return nil, nil
}
func (t *nonOp) InjectDepsData(key string, value interface{}) error {
	return nil
}
func (t *nonOp) GetInputsID() []string {
	return nil
}
func (t *nonOp) GetOutputsID() []string {
	return nil
}
func (t *nonOp) Reset() Operator {
	return t
}

type DataOperator1 struct {
}

func (p *DataOperator1) Name() string {
	return "DataOperator1"
}
func (p *DataOperator1) OnExecute(ctx *DAGContext) (map[string]interface{}, error) {
	return map[string]interface{}{
		"d1": 1,
		"d2": "Hello from DataOperator1",
	}, nil
}
func (p *DataOperator1) InjectDepsData(key string, value interface{}) error {
	return nil
}
func (p *DataOperator1) GetInputsID() []string {
	return make([]string, 0)
}
func (p *DataOperator1) GetOutputsID() []string {
	return []string{"d1", "d2"}
}
func (p *DataOperator1) Reset() Operator {
	return p
}

type DataOperator2 struct {
	d1 int
	d2 string
}

func (p *DataOperator2) Name() string {
	return "DataOperator2"
}
func (p *DataOperator2) OnExecute(ctx *DAGContext) (map[string]interface{}, error) {
	if p.d1 != 1 || p.d2 != "Hello from DataOperator1" {
		return nil, fmt.Errorf("DataOperator2's input data hasn't been set correctly")
	}
	log.Debugf("d1:%v, d2:%v", p.d1, p.d2)
	return map[string]interface{}{}, nil
}
func (p *DataOperator2) InjectDepsData(key string, value interface{}) error {
	ok := true
	switch key {
	case "d1":
		p.d1, ok = value.(int)
	case "d2":
		p.d2, ok = value.(string)
	default:
		return fmt.Errorf("%s is not the input of %s", key, p.Name())
	}
	if !ok {
		return fmt.Errorf("casting value of key:%s failed in %s Inject function", key, p.Name())
	}
	return nil
}
func (p *DataOperator2) GetInputsID() []string {
	return []string{"d1", "d2"}
}
func (p *DataOperator2) GetOutputsID() []string {
	return make([]string, 0)
}
func (p *DataOperator2) Reset() Operator {
	return p
}

func TestDefaultOperatorManager_RegisterOperator(t *testing.T) {
	for i := 1; i < 15; i++ {
		name := fmt.Sprintf("opr%d", i)
		j := i
		tOprMgr.RegisterOperator(name, func() Operator {
			return &timeOpr{name: name, number: j}
		})
		if tOprMgr.GetOperator(name) == nil {
			t.FailNow()
		}
	}
	for i := 1; i < 15; i++ {
		name := fmt.Sprintf("nonOp%d", i)
		tOprMgr.RegisterOperator(name, func() Operator {
			return &nonOp{name: name}
		})
		if tOprMgr.GetOperator(name) == nil {
			t.FailNow()
		}
	}
	tOprMgr.RegisterOperator("DataOperator1", func() Operator {
		return &DataOperator1{}
	})
	tOprMgr.RegisterOperator("DataOperator2", func() Operator {
		return &DataOperator2{}
	})
}

func TestNewDefaultOperatorManager(t *testing.T) {
	TestDefaultOperatorManager_RegisterOperator(t)
}
