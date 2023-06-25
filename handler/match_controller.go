package handler

import (
	"context"

	"match_controller/controller"
	match_controller "match_controller/proto"
)

type Match_controller struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *Match_controller) AddTask(ctx context.Context, req *match_controller.AddTaskReq, rsp *match_controller.AddTaskRsp) error {
	controller.DefaultManager.AddTask(req.GameId)
	return nil
}
