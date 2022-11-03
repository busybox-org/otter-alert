package engine

import (
	"fmt"
	"github.com/go-zookeeper/zk"
	"github.com/sirupsen/logrus"
	"github.com/xmapst/otteralert/internal/cache"
	"github.com/xmapst/otteralert/internal/config"
	"github.com/xmapst/otteralert/internal/notification"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

type Engine struct {
	interval time.Duration
	zkConn   *zk.Conn
	dbConn   *gorm.DB

	notification          notification.Notification
	manager               *config.Manager
	channelService        *cache.ChannelStateService
	restartChannelService *cache.RestartChannelService
	pipelineService       *cache.PipelineStateService
}

func New() *Engine {
	var e = &Engine{
		interval: config.App.Interval,
		manager:  config.App.Manager,
		channelService: &cache.ChannelStateService{
			Ch: make(chan cache.ChannelState, 10),
		},
		restartChannelService: &cache.RestartChannelService{},
		pipelineService: &cache.PipelineStateService{
			Ch: make(chan cache.PipelineState, 10),
		},
	}
	var err error
	e.zkConn, _, err = zk.Connect(config.App.Zookeeper, time.Second*1, zk.WithLogInfo(false))
	if err != nil {
		logrus.Fatalln(err)
	}
	e.dbConn, err = gorm.Open(mysql.Open(e.manager.DatabaseUrl), &gorm.Config{
		Logger: logger.New(
			logrus.StandardLogger(),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Silent,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		),
	})
	if err != nil {
		logrus.Fatalln(err)
	}
	e.notification, err = notification.New(config.App.Notification.Type, config.App.Notification.Url, config.App.Notification.Secret)
	if err != nil {
		logrus.Fatalln(err)
	}
	err = cache.New()
	if err != nil {
		logrus.Fatalln(err)
	}
	return e
}

func (e *Engine) Run() {
	// 定期查询log_record
	go e.logRecord()

	//定期查询delay_stat
	go e.dalyState()

	// 状态告警
	go func() {
		for {
			var title, message string
			select {
			case failed := <-e.channelService.Ch:
				title, message = e.createChannelStateMsgText(failed)
				if e.restartChannelService.Has(failed.ChannelID) {
					if failed.Active {
						e.restartChannelService.Delete(failed.ChannelID)
					}
					continue
				}
				// 解挂
				if ChState[failed.Status] == 2 {
					title, message = e.recoverChannel(failed.ChannelID)
				}
			case failed := <-e.pipelineService.Ch:
				title, message = e.createPipelineStateMsgText(failed)
			}
			if title != "" && message != "" {
				e.notification.SendMarkdown(title, message)
			}
		}
	}()
	// 数据采样
	go func() {
		// 10秒获取一次
		for range time.Tick(10 * time.Second) {
			e.getAllChannelPipelineState()
		}
	}()
	// 触发
	go e.pipelineService.Trigger(e.interval)
	go e.channelService.Trigger(e.interval)

	select {}
}

func (e *Engine) createChannelStateMsgText(failed cache.ChannelState) (string, string) {
	// 查找数据库
	channel := e.selectChannel(failed.ChannelID)
	interval := time.Now().Sub(failed.StartTime)
	state := "异常"
	if failed.Active {
		state = "恢复"
	}
	title := fmt.Sprintf("## 通道 %s %s", *channel.Name, state)
	message := fmt.Sprint(
		"\n- 状态: ", failed.Status,
		"\n- 持续: ", interval.String(),
		"\n- 时间: ", time.Now().Format("2006-01-02 15:04:05"),
	)
	return title, title + message
}

func (e *Engine) createPipelineStateMsgText(failed cache.PipelineState) (string, string) {
	// 查找数据库
	pipeline := e.selectPipeline(failed.PipelineID)
	channel := e.selectChannel(*pipeline.ChannelID)
	interval := time.Now().Sub(failed.StartTime)
	state := "异常"
	if failed.Active {
		state = "恢复"
	}
	title := fmt.Sprintf("## 管道 %s %s", *pipeline.Name, state)
	message := fmt.Sprint(
		"\n- 通道: ", channel.Name,
		"\n- 状态: ", failed.Status,
		"\n- 持续: ", interval.String(),
		"\n- 时间: ", time.Now().Format("2006-01-02 15:04:05"),
	)
	return title, title + message
}
