package handler

import (
	"context"
	"time"

	"match_controller/controller"
	match_controller "match_controller/proto"

	"github.com/micro/micro/v3/service/logger"
)

type Match_controller struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *Match_controller) AddTask(ctx context.Context, req *match_controller.AddTaskReq, rsp *match_controller.AddTaskRsp) error {
	logger.Infof("AddTask timer %v", time.Now().UnixNano()/1e6)
	controller.DefaultManager.AddTask(req.GameId)
	return nil
}
