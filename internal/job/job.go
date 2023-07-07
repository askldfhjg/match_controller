package job

import (
	"context"
	"encoding/json"
	"fmt"
	"match_controller/controller"
	"os"

	"github.com/micro/micro/v3/service/logger"
	"github.com/xxl-job/xxl-job-executor-go"
)

func Start() {
	addr := os.Getenv("XXLJOB_ADDRESS")
	exec := xxl.NewExecutor(
		xxl.ServerAddr(fmt.Sprintf("http://%s/xxl-job-admin", addr)),
		xxl.AccessToken(""),                  //请求令牌(默认为空)
		xxl.RegistryKey("mcbeam-match-jobs"), //执行器名称
	)
	exec.Init()
	//注册任务
	exec.RegTask("match_controller.task", task)
	err := exec.Run()
	if err != nil {
		logger.Errorf("xxjob init error %s", err.Error())
	}

}

type executorParams struct {
	GameId string
}

func task(cxt context.Context, param *xxl.RunReq) (msg string) {
	logger.Info("xxl-job===" + param.ExecutorHandler + " param：" + param.ExecutorParams + " log_id:" + xxl.Int64ToStr(param.LogID))
	params := &executorParams{}
	err := json.Unmarshal([]byte(param.ExecutorParams), params)
	if err != nil {
		logger.Info("xxl-job===ExecutorParams error", err)
	}
	err = controller.DefaultManager.AddTask(params.GameId)
	if err != nil {
		logger.Info("xxl-job===invoke addTask error", err)
		return "error"
	}
	return "done"
}
