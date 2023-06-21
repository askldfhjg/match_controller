package manager

import (
	"context"
	"fmt"
	"match_controller/center"
	"match_controller/internal/db"
	"match_controller/utils"
	"math"
	"sync"
	"time"

	match_process "github.com/askldfhjg/match_apis/match_process/proto"
	"google.golang.org/protobuf/proto"

	"github.com/micro/micro/v3/service/broker"
	"github.com/micro/micro/v3/service/client"
	"github.com/micro/micro/v3/service/logger"
	"github.com/micro/micro/v3/service/server"
)

type gameConfig struct {
	GameId      string
	GroupCount  int
	OffsetCount int
	SubType     int64
	NeedCount   int64
}

func NewManager(opts ...center.CenterOption) center.Manager {
	m := &defaultMgr{
		exited:       make(chan struct{}, 1),
		matchChannel: make(chan string, 2000),
		gameConfig:   &sync.Map{},
	}
	for _, o := range opts {
		o(&m.opts)
	}
	return m
}

type defaultMgr struct {
	opts         center.CenterOptions
	exited       chan struct{}
	matchChannel chan string
	gameConfig   *sync.Map //map[string]*gameConfig
}

func (m *defaultMgr) Start() error {
	go m.loop()
	return nil
}

func (m *defaultMgr) Stop() error {
	close(m.exited)
	return nil
}

func (m *defaultMgr) loop() {
	ticket := time.NewTicker(time.Second * 10)
	defer ticket.Stop()
	for {
		select {
		case <-m.exited:
			return
		case gameId := <-m.matchChannel:
			val, ok := m.gameConfig.Load(gameId)
			if ok {
				config, okk := val.(*gameConfig)
				if okk && config != nil {
					m.processTask(config)
				}
			}
		}
	}
}

func (m *defaultMgr) processTask(config *gameConfig) {
	var count int
	var err error
	ctx := context.Background()
	if count, err = db.Default.GetQueueCount(ctx, config.GameId, config.SubType); err != nil {
		logger.Errorf("processTask get GetQueueCount %s %d error %s", config.GameId, config.SubType, err.Error())
	}
	if count <= 0 {
		return
	}

	version, err := m.PublishPoolVersion(ctx, config.GameId, config.SubType)
	if err != nil {
		logger.Errorf("processTask AddPoolVersion %s %d error %s", config.GameId, config.SubType, err.Error())
		return
	}
	segCount := int(math.Ceil(float64(count) / float64(config.GroupCount)))
	reqList := make([]*match_process.MatchTaskReq, 0, 10)
	matchId := fmt.Sprintf("%s-%d", server.DefaultServer.Options().Id, time.Now().UnixNano()/1e6)
	evalhaskKey := utils.RandomString(15)
	EvalGroupId := matchId
	for i := 0; i < segCount; i++ {
		st := (i * config.GroupCount) + 1 - config.OffsetCount
		ed := (i+1)*config.GroupCount + config.OffsetCount
		if st <= 0 {
			st = 1
		}
		needStop := false
		if ed >= count {
			ed = count
			needStop = true
		}
		reqList = append(reqList, &match_process.MatchTaskReq{
			TaskId:             matchId,
			SubTaskId:          fmt.Sprintf("%s-%d", matchId, i+1),
			GameId:             config.GameId,
			SubType:            config.SubType,
			StartPos:           int64(st),
			EndPos:             int64(ed),
			EvalGroupId:        EvalGroupId,
			EvalGroupTaskCount: 0,
			EvalGroupSubId:     int64(i + 1),
			EvalhaskKey:        evalhaskKey,
			NeedCount:          config.NeedCount,
			Version:            version,
		})
		if needStop {
			break
		}
	}
	go func() {
		realSegCount := len(reqList)
		matchSrv := match_process.NewMatchProcessService("match_process", client.DefaultClient)
		for _, rr := range reqList {
			rr.EvalGroupTaskCount = int64(realSegCount)
			rsp, err := matchSrv.MatchTask(context.Background(), rr)
			if err != nil {
				logger.Infof("MatchTask error %+v", err)
			} else {
				logger.Infof("MatchTask result %+v", rsp)
			}
		}
	}()
}

func (m *defaultMgr) PublishPoolVersion(ctx context.Context, gameId string, subType int64) (int64, error) {
	version := time.Now().UnixNano()
	err := db.Default.AddPoolVersion(ctx, gameId, subType, version)
	if err != nil {
		return 0, err
	}
	msg := &match_process.PoolVersionMsg{
		GameId:  gameId,
		SubType: subType,
		Version: version,
	}
	by, _ := proto.Marshal(msg)
	err = broker.Publish("pool_version", &broker.Message{
		Header: map[string]string{"gameId": gameId},
		Body:   by,
	})
	if err != nil {
		logger.Errorf("PublishPoolVersion publish err : %s", err.Error())
	}
	return version, err
}
