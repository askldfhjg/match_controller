package main

import (
	"match_controller/center"
	"match_controller/handler"
	"match_controller/internal/manager"
	pb "match_controller/proto"

	"github.com/micro/micro/v3/service"
	"github.com/micro/micro/v3/service/logger"
)

func main() {
	// Create service
	center.DefaultManager = manager.NewManager()
	srv := service.New(
		service.Name("match_controller"),
		service.Version("latest"),
		service.AfterStart(center.DefaultManager.Start),
		service.BeforeStop(center.DefaultManager.Stop),
	)

	// Register handler
	pb.RegisterMatchControllerHandler(srv.Server(), new(handler.Match_controller))

	// Run service
	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}
