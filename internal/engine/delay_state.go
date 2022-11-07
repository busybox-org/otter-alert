package engine

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/xmapst/otteralert/internal/cache"
	"github.com/xmapst/otteralert/internal/utils"
	"time"
)

func (e *Engine) dalyState() {
	analysisTopState := cache.AnalysisTopStateService{
		Ch: make(chan cache.AnalysisTopState, 10),
	}
	go func() {
		for failed := range analysisTopState.Ch {
			state := e.getChannelState(failed.ChannelID)
			if state != "START" {
				continue
			}
			logrus.Warnf("通道%d延时 %s", failed.ChannelID, utils.FmtDuration(failed.Interval))
			title, message := e.restartChannel(failed.ChannelID)
			if title == "" && message == "" {
				channel := e.selectChannel(failed.ChannelID)
				title = fmt.Sprintf("## 通道%s恢复成功", *channel.Name)
				message = title + fmt.Sprintf("\n- 延时: %s", utils.FmtDuration(failed.Interval))
			}
			e.notification.SendMarkdown(title, message)
		}
	}()

	// 数据采样
	go func() {
		for range time.Tick(60 * time.Second) {
			pipelines := e.selectAllPipeline()
			for _, pipeline := range pipelines {
				go func(pipeline Pipeline) {
					_p := e.selectPipeline(*pipeline.ID)
					state := e.selectDelayStat(*pipeline.ID)
					interval := time.Now().Sub(state.GMTModified)
					if interval >= 15*time.Minute {
						analysisTopState.Add(*pipeline.ID, *_p.ChannelID, interval)
					} else {
						analysisTopState.Delete(*pipeline.ID)
					}
				}(pipeline)
			}
		}
	}()

	// 触发
	analysisTopState.Trigger(e.interval)
}
