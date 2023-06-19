package main

import (
	"match_controller/handler"
	pb "match_controller/proto"

	"github.com/micro/micro/v3/service"
	"github.com/micro/micro/v3/service/logger"
)

func main() {
	// Create service
	srv := service.New(
		service.Name("match_controller"),
		service.Version("latest"),
	)

	// Register handler
	pb.RegisterMatch_controllerHandler(srv.Server(), new(handler.Match_controller))

	// Run service
	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}
