package engine

import (
	"fmt"
	"time"
)

func (e *Engine) logRecord() {
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
			"\n- 管道: ", *pipeline.Name,
			"\n- 节点: ", node.Name,
			message,
		)
	}
	return title, title + message
}
