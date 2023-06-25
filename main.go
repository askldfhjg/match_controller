package main

import (
	"match_controller/controller"
	"match_controller/handler"
	"match_controller/internal/db"
	"match_controller/internal/db/redis"
	"match_controller/internal/manager"
	pb "match_controller/proto"

	"github.com/micro/micro/v3/service"
	"github.com/micro/micro/v3/service/logger"
)

func main() {
	// Create service
	controller.DefaultManager = manager.NewManager()
	srv := service.New(
		service.Name("match_controller"),
		service.Version("latest"),
		service.BeforeStart(func() error {
			svr, err := redis.New(
				db.WithAddress("127.0.0.1:6379"),
				db.WithPoolMaxActive(5),
				db.WithPoolMaxIdle(100),
				db.WithPoolIdleTimeout(300))
			if err != nil {
				return err
			}
			db.Default = svr
			return nil
		}),
		service.AfterStart(controller.DefaultManager.Start),
		service.BeforeStop(controller.DefaultManager.Stop),
	)

	// Register handler
	pb.RegisterMatchControllerHandler(srv.Server(), new(handler.Match_controller))

	// Run service
	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}
