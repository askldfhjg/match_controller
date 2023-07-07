package job

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"match_controller/controller"
	"os"

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
	log.Fatal(exec.Run())
}

type executorParams struct {
	GameId string
}

func task(cxt context.Context, param *xxl.RunReq) (msg string) {
	fmt.Println("xxl-job===" + param.ExecutorHandler + " param：" + param.ExecutorParams + " log_id:" + xxl.Int64ToStr(param.LogID))
	params := &executorParams{}
	err := json.Unmarshal([]byte(param.ExecutorParams), params)
	if err != nil {
		fmt.Println("xxl-job===ExecutorParams error", err)
	}
	err = controller.DefaultManager.AddTask(params.GameId)
	if err != nil {
		fmt.Println("xxl-job===invoke addTask error", err)
		return "error"
	}
	return "done"
}
