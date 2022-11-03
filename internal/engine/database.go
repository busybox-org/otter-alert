package engine

import (
	"strconv"
	"time"
)

type Pipeline struct {
	ID          *int64    `gorm:"column:ID"`
	Name        *string   `gorm:"column:NAME"`
	Description *string   `gorm:"column:DESCRIPTION"`
	PARAMETERS  *string   `gorm:"column:PARAMETERS"`
	ChannelID   *int64    `gorm:"column:CHANNEL_ID"`
	GMTCreate   time.Time `gorm:"column:GMT_CREATE"`
	GMTModified time.Time `gorm:"column:GMT_MODIFIED"`
}

func (e *Engine) selectPipeline(id int64) (res Pipeline) {
	e.dbConn.Table("pipeline").First(&res, id)
	if res.Name == nil {
		var name = strconv.Itoa(int(id))
		res.Name = &name
	}
	return
}

func (e *Engine) selectAllPipeline() (res []Pipeline) {
	e.dbConn.Table("pipeline").Find(&res)
	return
}

type Channel struct {
	ID          *int64    `gorm:"column:ID"`
	Name        *string   `gorm:"column:NAME"`
	Description *string   `gorm:"column:DESCRIPTION"`
	PARAMETERS  *string   `gorm:"column:PARAMETERS"`
	GMTCreate   time.Time `gorm:"column:GMT_CREATE"`
	GMTModified time.Time `gorm:"column:GMT_MODIFIED"`
}

func (e *Engine) selectChannel(id int64) (res Channel) {
	e.dbConn.Table("channel").First(&res, id)
	if res.Name == nil {
		var name = strconv.Itoa(int(id))
		res.Name = &name
	}
	return
}

type DelayStat struct {
	ID          *int64    `gorm:"column:ID"`
	DelayTtime  *int64    `gorm:"column:DELAY_TIME"`
	DelayNumber *int64    `gorm:"column:DELAY_NUMBER"`
	PipelineID  *int64    `gorm:"column:PIPELINE_ID"`
	GMTCreate   time.Time `gorm:"column:GMT_CREATE"`
	GMTModified time.Time `gorm:"column:GMT_MODIFIED"`
}

func (e *Engine) selectDelayStat(pipelineID int64) (res DelayStat) {
	e.dbConn.Table("delay_stat").Where("PIPELINE_ID = ?", pipelineID).Last(&res)
	return
}

type LogRecord struct {
	ID          int64     `gorm:"column:ID"`
	NID         int64     `gorm:"column:NID"`
	ChannelID   int64     `gorm:"column:CHANNEL_ID"`
	PipelineID  int64     `gorm:"column:PIPELINE_ID"`
	Title       string    `gorm:"column:TITLE"`
	Message     string    `gorm:"column:MESSAGE"`
	GMTCreate   time.Time `gorm:"column:GMT_CREATE"`
	GMTModified time.Time `gorm:"column:GMT_MODIFIED"`
}

func (e *Engine) selectLogRecord() (res LogRecord) {
	// Get last record, ordered by primary key desc
	e.dbConn.Table("log_record").Last(&res)
	return
}

type Node struct {
	ID          int64  `gorm:"column:ID"`
	Name        string `gorm:"column:NAME"`
	IP          string `gorm:"column:IP"`
	Description string `gorm:"column:DESCRIPTION"`
}

func (e *Engine) selectNode(id int64) (res Node) {
	e.dbConn.Table("node").First(&res, id)
	return
}
