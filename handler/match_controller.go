package handler

import (
	"context"

	log "github.com/micro/micro/v3/service/logger"

	match_controller "match_controller/proto"
)

type Match_controller struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *Match_controller) Call(ctx context.Context, req *match_controller.Request, rsp *match_controller.Response) error {
	log.Info("Received Match_controller.Call request")
	rsp.Msg = "Hello " + req.Name
	return nil
}
