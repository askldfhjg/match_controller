package manager

import (
	"context"
	"fmt"
	"match_controller/controller"
	"match_controller/internal/db"
	"match_controller/utils"
	"sync"
	"time"

	match_process "github.com/askldfhjg/match_apis/match_process/proto"
	"google.golang.org/protobuf/proto"

	"github.com/micro/micro/v3/service/broker"
	"github.com/micro/micro/v3/service/client"
	"github.com/micro/micro/v3/service/logger"
)

type poolConfig struct {
	GameId      string
	GroupCount  int
	OffsetCount int
	SubType     int64
	NeedCount   int64
}

type gameConfig struct {
	GameId string
	Pools  []*poolConfig
}

func NewManager(opts ...controller.CenterOption) controller.Manager {
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
	opts         controller.CenterOptions
	exited       chan struct{}
	matchChannel chan string
	gameConfig   *sync.Map //map[string]*gameConfig
}

func (m *defaultMgr) Start() error {
	m.gameConfig.Store("aaaa", &gameConfig{
		GameId: "aaaa",
		Pools: []*poolConfig{
			{
				GameId:      "aaaa",
				GroupCount:  3000,
				OffsetCount: 50,
				SubType:     1,
				NeedCount:   3,
			},
		},
	})
	err := m.initPoolVersion()
	if err != nil {
		return err
	}
	go m.loop()
	return nil
}

func (m *defaultMgr) Stop() error {
	close(m.exited)
	return nil
}

func (m *defaultMgr) AddTask(gameId string) error {
	m.matchChannel <- gameId
	return nil
}

func (m *defaultMgr) initPoolVersion() error {
	version := time.Now().UnixNano()
	var err error
	m.gameConfig.Range(func(key interface{}, val interface{}) bool {
		config, ok := val.(*gameConfig)
		if ok {
			for _, cfg := range config.Pools {
				err = db.Default.InitPoolVersion(context.Background(), cfg.GameId, cfg.SubType, version)
				if err != nil {
					return false
				}
			}
		}
		return true
	})
	return err
}

func (m *defaultMgr) loop() {
	for {
		select {
		case <-m.exited:
			return
		case gameId := <-m.matchChannel:
			val, ok := m.gameConfig.Load(gameId)
			if ok {
				config, okk := val.(*gameConfig)
				if okk && config != nil {
					for _, poolCfg := range config.Pools {
						go m.processTask(poolCfg)
					}
				}
			}
		}
	}
}

func (m *defaultMgr) processTask(config *poolConfig) {
	var groups []int
	var err error
	//logger.Infof("processTask %+v", config)
	logger.Infof("processTask group result %+v", groups)
	version, oldV, err := m.PublishPoolVersion(config.GameId, config.SubType)
	if err != nil {
		logger.Errorf("processTask AddPoolVersion %s %d error %s", config.GameId, config.SubType, err.Error())
		return
	}

	if groups, err = db.Default.GetQueueCounts(context.Background(), oldV, config.GameId, config.SubType, config.GroupCount); err != nil {
		logger.Errorf("processTask get GetQueueCount %s %d error %s", config.GameId, config.SubType, err.Error())
		return
	}
	if len(groups) <= 0 {
		return
	}

	segCount := len(groups)
	reqList := make([]*match_process.MatchTaskReq, segCount)
	matchId := fmt.Sprintf("%s-%d-%d", config.GameId, config.SubType, time.Now().UnixNano())
	//evalhaskKey := utils.RandomString(15)
	//EvalGroupId := evalhaskKey
	startTime := time.Now().UnixNano() / 1e6
	realSegCount := 0
	for i := 0; i+1 < segCount; i++ {
		evalhaskKey := utils.RandomString(15)
		EvalGroupId := evalhaskKey
		st := groups[i] + 1
		if i == 0 {
			st = groups[i]
		}
		ed := groups[i+1]
		reqList[i] = &match_process.MatchTaskReq{
			TaskId:             matchId,
			SubTaskId:          fmt.Sprintf("%s-%d", matchId, i+1),
			GameId:             config.GameId,
			SubType:            config.SubType,
			StartPos:           int64(st),
			EndPos:             int64(ed),
			EvalGroupId:        EvalGroupId,
			EvalGroupTaskCount: 1,
			EvalGroupSubId:     1,
			EvalhaskKey:        evalhaskKey,
			NeedCount:          config.NeedCount,
			Version:            version,
			StartTime:          startTime,
			OldVersion:         oldV,
		}
		realSegCount++
	}
	// go func() {
	// 	matchSrv := match_process.NewMatchProcessService("match_process", client.DefaultClient)
	// 	for _, rr := range reqList {
	// 		if rr == nil {
	// 			continue
	// 		}
	// 		//rr.EvalGroupTaskCount = int64(realSegCount)
	// 		//logger.Infof("result %+v", rr)
	// 		_, err := matchSrv.MatchTask(context.Background(), rr)
	// 		if err != nil {
	// 			logger.Infof("processTask send error %+v", err)
	// 		} else {
	// 			//logger.Infof("processTask send result %+v", rsp)
	// 		}
	// 	}
	// }()
	for _, rr := range reqList {
		if rr == nil {
			continue
		}
		//rr.EvalGroupTaskCount = int64(realSegCount)
		//logger.Infof("result %+v", rr)
		req := rr
		go func() {
			matchSrv := match_process.NewMatchProcessService("match_process", client.DefaultClient)
			_, err := matchSrv.MatchTask(context.Background(), req)
			if err != nil {
				logger.Infof("processTask send error %+v", err)
			} else {
				//logger.Infof("processTask send result %+v", rsp)
			}
		}()
	}
}

func (m *defaultMgr) PublishPoolVersion(gameId string, subType int64) (int64, int64, error) {
	version := time.Now().UnixNano()
	oldV, err := db.Default.AddPoolVersion(context.Background(), gameId, subType, version)
	if err != nil {
		return 0, 0, err
	}
	msg := &match_process.PoolVersionMsg{
		GameId:     gameId,
		SubType:    subType,
		Version:    version,
		OldVersion: oldV,
	}
	by, _ := proto.Marshal(msg)
	err = broker.Publish("pool_version", &broker.Message{
		Header: map[string]string{"gameId": gameId},
		Body:   by,
	})
	if err != nil {
		db.Default.DelPoolVersion(context.Background(), gameId, subType)
		logger.Errorf("PublishPoolVersion publish err : %s", err.Error())
		return 0, 0, err
	}
	return version, oldV, err
}
