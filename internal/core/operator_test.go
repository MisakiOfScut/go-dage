package core

import (
	"fmt"
	"testing"
	"time"
)

var tOprMgr = NewDefaultOperatorManager()

type tOpr struct {
	name string
}

func (t *tOpr) OnExecute(ctx *dagContext) error {
	fmt.Println(t.name, time.Now().UnixNano())
	return nil
}

func TestDefaultOperatorManager_RegisterOperator(t *testing.T) {
	for i := 1; i < 5; i++ {
		name := fmt.Sprintf("opr%d", i)
		opr := Operator{Name: name, Processor: &tOpr{name: name}}
		tOprMgr.RegisterOperator(opr.Name, &opr)
		if tOprMgr.GetOperator(name) == nil {
			t.Fail()
		}
	}
}

func TestDefaultOperatorManager_GetOperator(t *testing.T) {
	if tOprMgr.GetOperator("@#$%") != nil {
		t.Fail()
	}
}

func TestNewDefaultOperatorManager(t *testing.T) {
	TestDefaultOperatorManager_RegisterOperator(t)
	TestDefaultOperatorManager_GetOperator(t)
}
