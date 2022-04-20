package dage

import (
	"github.com/BurntSushi/toml"
	"github.com/MisakiOfScut/go-dage/internal/core"
	"github.com/MisakiOfScut/go-dage/internal/script"
	"github.com/MisakiOfScut/go-dage/internal/utils/executor"
	"github.com/MisakiOfScut/go-dage/internal/utils/log"
)

var (
	_globalOprMgr = core.NewDefaultOperatorManager()
	_globalE      = core.NewGraphManager(executor.NewDefaultExecutor(32, 8), _globalOprMgr)
)

// SetLogger sets an logger instance for dag engine, or it won't print any internal logs
// You can use zap global sugarLogger as default logger:
// 	e.x.
// 		logger := zap.NewExample()  // DEBUG level, log to stdout
// 		defer logger.Sync()
// 		dage.SetLogger(logger.Sugar())
//
func SetLogger(logger log.Logger) {
	log.SetLogger(logger)
}

// Execute a specific graph in a specific graph cluster. You can specify a timeout for the execution,
// non-positive value will be treated as zero while zero means no timeout
func Execute(userData interface{}, graphClusterName string, graphName string,
	timeoutMillisecond int64) error {
	return _globalE.Execute(userData, graphClusterName, graphName, timeoutMillisecond)
}

// BuildAndSetDAG parse the input script and build an executable dag from it,
// and this function only returns build error.
// If you set a dag with a duplicated name, the previous one will be replaced.
func BuildAndSetDAG(dagName string, tomlScript *string) error {
	return _globalE.Build(dagName, tomlScript)
}

// Stop the engine when your app ends. The engine will be stopped after execute the remaining tasks in the executor's
// queue. After calling this function, you shouldn't call any other functions, which may cause undefined behaviors.
func Stop() {
	_globalE.Stop()
}

func DumpDAGDot(graphClusterName string) string {
	return _globalE.DumpDAGDot(graphClusterName)
}

func TestBuildDAG(tomlScript *string, mockGraphMgr script.IGraphManager) error {
	graphCluster := script.NewGraphCluster(mockGraphMgr)
	if _, err := toml.Decode(*tomlScript, graphCluster); err != nil {
		return err
	}
	if err := graphCluster.Build(); err != nil {
		return err
	}
	return nil
}

// ReplaceExecutor replace the executor of the engine.
// The default executor is created with 32 queueLength and 8 concurrentLevel.
// You can call this function before executing graphs.
func ReplaceExecutor(executor executor.Executor) {
	_globalE.ReplaceExecutor(executor)
}

// RegisterOperator add an operator object to engine.
// Attention: 1. add an opr with duplicated name will replace the previous one; 2.
// the operator object will be shared with every dag execution;
func RegisterOperator(oprName string, operator *core.Operator) {
	_globalOprMgr.RegisterOperator(oprName, operator)
}
