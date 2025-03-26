package otter

import (
	"regexp"
	"time"

	"github.com/tidwall/gjson"
)

type Channel struct {
	ID          int64     `gorm:"column:ID"`
	Name        string    `gorm:"column:NAME"`
	Description string    `gorm:"column:DESCRIPTION"`
	PARAMETERS  string    `gorm:"column:PARAMETERS"`
	GMTCreate   time.Time `gorm:"column:GMT_CREATE"`
	GMTModified time.Time `gorm:"column:GMT_MODIFIED"`
}

func (o *sOtter) GetChannel(channelID int64) (channel *Channel, err error) {
	err = o.dbConn.Table("CHANNEL").Where("ID = ?", channelID).First(&channel).Error
	return
}

func (o *sOtter) GetAllChannel() (channels []*Channel, err error) {
	err = o.dbConn.Table("CHANNEL").Find(&channels).Error
	return
}

type Pipeline struct {
	ID              int64     `gorm:"column:ID"`
	Name            string    `gorm:"column:NAME"`
	Description     string    `gorm:"column:DESCRIPTION"`
	PARAMETERS      string    `gorm:"column:PARAMETERS"`
	ChannelID       int64     `gorm:"column:CHANNEL_ID"`
	GMTCreate       time.Time `gorm:"column:GMT_CREATE"`
	GMTModified     time.Time `gorm:"column:GMT_MODIFIED"`
	DestinationName string
}

func (o *sOtter) GetPipeline(channelID, piplineID int64) (pipeline *Pipeline, err error) {
	err = o.dbConn.Table("PIPELINE").Where("ID = ? AND CHANNEL_ID = ?", piplineID, channelID).First(&pipeline).Error
	if err != nil {
		return
	}
	pipeline.DestinationName = gjson.Get(pipeline.PARAMETERS, "destinationName").String()
	return
}

func (o *sOtter) GetAllPipeline(channelID int64) (pipelines []*Pipeline, err error) {
	err = o.dbConn.Table("PIPELINE").Where("CHANNEL_ID = ?", channelID).First(&pipelines).Error
	if err != nil {
		return
	}
	for _, pipeline := range pipelines {
		pipeline.DestinationName = gjson.Get(pipeline.PARAMETERS, "destinationName").String()
	}
	return
}

type DelayStat struct {
	ID          int64     `gorm:"column:ID"`
	DelayTtime  int64     `gorm:"column:DELAY_TIME"`
	DelayNumber int64     `gorm:"column:DELAY_NUMBER"`
	PipelineID  int64     `gorm:"column:PIPELINE_ID"`
	GMTCreate   time.Time `gorm:"column:GMT_CREATE"`
	GMTModified time.Time `gorm:"column:GMT_MODIFIED"`
}

func (o *sOtter) GetDelayStat(pipelineId int64) (delayStat *DelayStat, err error) {
	err = o.dbConn.Table("DELAY_STAT").Where("PIPELINE_ID = ?", pipelineId).Order("ID DESC").First(&delayStat).Error
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

var regexps = []*regexp.Regexp{
	regexp.MustCompile(`(?m)PreparedStatementCallback.*`),
	regexp.MustCompile(`(?m)MySQLSyntaxErrorException.*`),
	regexp.MustCompile(`(?m)TransformException.*`),
}

func (o *sOtter) GetLogRecord(channelID int64) (log string, err error) {
	var logRecord LogRecord
	err = o.dbConn.Table("LOG_RECORD").Where("CHANNEL_ID = ?", channelID).Order("GMT_CREATE DESC").First(&logRecord).Error
	if err != nil {
		return
	}
	for _, re := range regexps {
		// regexp FindAllString
		for _, match := range re.FindAllString(logRecord.Message, -1) {
			log += match
		}
	}
	return
}

type Node struct {
	ID          int64  `gorm:"column:ID"`
	Name        string `gorm:"column:NAME"`
	IP          string `gorm:"column:IP"`
	Description string `gorm:"column:DESCRIPTION"`
}

func (o *sOtter) GetNode(id int64) (node *Node, err error) {
	err = o.dbConn.Table("NODE").Where("ID = ?", id).First(&node).Error
	return
}

func (o *sOtter) GetAllNode() (nodes []*Node, err error) {
	err = o.dbConn.Table("NODE").Find(&nodes).Error
	return
}
