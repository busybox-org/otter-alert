package engine

import (
	"fmt"
	"github.com/go-zookeeper/zk"
	"github.com/sirupsen/logrus"
	"github.com/xmapst/otter-alert/dataprovider"
	"github.com/xmapst/otter-alert/internal/config"
	"github.com/xmapst/otter-alert/internal/notification"
	"github.com/xmapst/otter-alert/utils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strconv"
	"time"
)

type Engine struct {
	interval     time.Duration
	zkConn       *zk.Conn
	dbConn       *gorm.DB
	notification notification.Notification
	manager      *config.Manager
}

func New() *Engine {
	err := utils.CheckURL(config.App.Notification.Url)
	if err != nil {
		logrus.Fatalln(err)
	}
	var e = &Engine{
		interval: config.App.Interval,
		manager:  config.App.Manager,
	}
	e.zkConn, _, err = zk.Connect(config.App.Zookeeper, time.Second*1, zk.WithLogInfo(false))
	if err != nil {
		logrus.Fatalln(err)
	}
	e.dbConn, err = gorm.Open(mysql.Open(e.manager.DatabaseUrl), &gorm.Config{})
	if err != nil {
		logrus.Fatalln(err)
	}
	e.notification, err = notification.New(config.App.Notification.Type, config.App.Notification.Url, config.App.Notification.Secret)
	if err != nil {
		logrus.Fatalln(err)
	}
	return e
}

func (e *Engine) Run() {
	// 状态采集及告警
	go func() {
		cache := dataprovider.NewChannelState()
		cache.FailedCh = make(chan dataprovider.CFailed, 10)
		go func() {
			for failed := range cache.FailedCh {
				// 解挂
				if failed.Status == "PAUSE" && failed.PipelineID == 0 {
					title, message := e.recoverChannel(failed.ChannelID)
					if title != "" && message != "" {
						e.notification.SendMarkdown(title, message)
					}
					continue
				}
				// 发送告警
				title, message := e.createStateMsgText(failed)
				e.notification.SendMarkdown(title, message)
			}
		}()
		// 数据采样
		go func() {
			for range time.Tick(150 * time.Millisecond) {
				e.getAllChannelState(cache)
			}
		}()
		// 触发
		cache.Trigger(e.interval)
	}()

	// 定期查询log_record
	go func() {
		var id int64
		var interval = 15 * time.Second
		for range time.Tick(interval) {
			res := e.selectLogRecord()
			if id != res.ID && time.Now().Sub(res.GMTCreate) <= interval {
				id = res.ID
				title, message := e.createLogRecordMsgText(res)
				e.notification.SendMarkdown(title, message)
			}
		}
	}()

	//定期查询delay_stat
	go func() {
		cache := dataprovider.NewAnalysisTopState()
		cache.FailedCh = make(chan dataprovider.AFailed, 10)
		go func() {
			for failed := range cache.FailedCh {
				if e.getChannelState(failed.ChannelID) != "START" {
					continue
				}
				logrus.Warnf("通道%d延时", failed.Interval)
				title, message := e.restartChannel(failed.ChannelID)
				if title != "" && message != "" {
					e.notification.SendMarkdown(title, message)
				}
			}
		}()

		// 数据采样
		go func() {
			for range time.Tick(15 * time.Second) {
				pipelines := e.selectAllPipeline()
				for _, pipeline := range pipelines {
					go func(pipeline Pipeline) {
						_p := e.selectPipeline(*pipeline.ID)
						state := e.selectDelayStat(*pipeline.ID)
						interval := time.Now().Sub(state.GMTModified)
						var id = strconv.FormatInt(*pipeline.ID, 10)
						if interval > 15*time.Minute {
							cache.Add(id, *_p.ChannelID, *_p.ID, interval)
						} else {
							cache.Del(id)
						}
					}(pipeline)
				}
			}
		}()

		// 触发
		cache.Trigger(e.interval)
	}()
	// 阻塞
	select {}
}

func (e *Engine) createStateMsgText(failed dataprovider.CFailed) (string, string) {
	var name string
	// 查找数据库
	if failed.PipelineID != 0 {
		pipeline := e.selectPipeline(failed.PipelineID)
		name = *pipeline.Name
	} else {
		channel := e.selectChannel(failed.ChannelID)
		name = *channel.Name
	}
	count := time.Now().Unix() - failed.StartTime
	end := fmt.Sprintf("%d秒", count)
	if count >= 60 {
		count = count / 60
		end = fmt.Sprintf("%d分钟", count)
	}
	var title = fmt.Sprintf("## %s 异常", name)
	if failed.Active {
		title = fmt.Sprintf("## %s 恢复", name)
		failed.StartTime = time.Now().Unix()
	}
	message := fmt.Sprint(
		"\n- 状态: ", failed.Status,
		"\n- 持续: ", end,
		"\n- 时间: ", time.Unix(failed.StartTime, 0).Format("2006-01-02 15:04:05"),
	)
	return title, title + message
}

func (e *Engine) createLogRecordMsgText(record LogRecord) (string, string) {
	var title = fmt.Sprintf("## Manager 异常")
	var message = fmt.Sprint(
		"\n- 时间: ", record.GMTCreate,
		"\n- 内容: ", record.Message,
	)
	if record.NID > 0 {
		// 查询节点名称
		node := e.selectNode(record.NID)
		channel := e.selectChannel(record.ChannelID)
		pipeline := e.selectPipeline(record.PipelineID)
		title = fmt.Sprintf("## %s 异常", *pipeline.Name)
		message = fmt.Sprint(
			"\n- 通道: ", *channel.Name,
			"\n - 管道: ", *pipeline.Name,
			"\n - 节点: ", node.Name,
			message,
		)
	}
	return title, title + message
}
