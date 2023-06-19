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

// Stream is a server side stream handler called via client.Stream or the generated client code
func (e *Match_controller) Stream(ctx context.Context, req *match_controller.StreamingRequest, stream match_controller.Match_controller_StreamStream) error {
	log.Infof("Received Match_controller.Stream request with count: %d", req.Count)

	for i := 0; i < int(req.Count); i++ {
		log.Infof("Responding: %d", i)
		if err := stream.Send(&match_controller.StreamingResponse{
			Count: int64(i),
		}); err != nil {
			return err
		}
	}

	return nil
}

// PingPong is a bidirectional stream handler called via client.Stream or the generated client code
func (e *Match_controller) PingPong(ctx context.Context, stream match_controller.Match_controller_PingPongStream) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		log.Infof("Got ping %v", req.Stroke)
		if err := stream.Send(&match_controller.Pong{Stroke: req.Stroke}); err != nil {
			return err
		}
	}
}
