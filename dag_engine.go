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

// Execute a specific graph in a specific graph cluster.
// 1. You can specify a timeout for the execution,
// non-positive value will be treated as zero while zero means no timeout.
// 2. You can pass a done function(nil is allowed) which will be executed after executing dag
func Execute(userData interface{}, graphClusterName string, graphName string,
	timeoutMillisecond int64, doneClosure func()) error {
	return _globalE.Execute(userData, graphClusterName, graphName, timeoutMillisecond, doneClosure)
}

// BuildAndSetDAG parse the input script and build an executable dag from it,
// and this function only returns build error.
// If you set a dag with a duplicated name, the previous one will be replaced.
func BuildAndSetDAG(clusterName string, tomlScript *string) error {
	return _globalE.Build(clusterName, tomlScript)
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
	_globalE.ReplaceTaskExecutor(executor)
}

// RegisterOperator add an operator object new function to engine.
// Attention: add a function with duplicated name will replace the previous one;
func RegisterOperator(oprName string, fun core.NewOperatorFunction) {
	_globalOprMgr.RegisterOperator(oprName, fun)
}
