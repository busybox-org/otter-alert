package otter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type Binlog struct {
	Type     string   `json:"@type"`
	Identity Identity `json:"identity"`
	Postion  Postion  `json:"postion"`
}

type Identity struct {
	SlaveID       int           `json:"slaveId"`
	SourceAddress SourceAddress `json:"sourceAddress"`
}

type SourceAddress struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

type Postion struct {
	Gtid        string `json:"gtid"`
	Included    bool   `json:"included"`
	JournalName string `json:"journalName"`
	Position    int    `json:"position"`
	ServerID    int    `json:"serverId"`
	Timestamp   int64  `json:"timestamp"`
	TimeString  string
}

func (o *sOtter) GetBinlog(info *Pipeline) (*Binlog, error) {
	path := fmt.Sprintf("/otter/canal/destinations/%s/%d/cursor", info.DestinationName, info.ID)
	data, _, err := o.zkConn.Get(path)
	if err != nil {
		return nil, err
	}
	binlog := &Binlog{}
	if err = json.Unmarshal(data, binlog); err != nil {
		return nil, err
	}
	seconds := binlog.Postion.Timestamp / 1000
	timestamp := time.Unix(seconds, 0)
	beijingTime := timestamp.In(time.Local)
	binlog.Postion.TimeString = beijingTime.Format("2006-01-02 15:04:05")
	return binlog, nil
}

func (o *sOtter) GetChannelState(id int64) (string, error) {
	path := fmt.Sprintf("/otter/channel/%d", id)
	data, _, err := o.zkConn.Get(path)
	if err != nil {
		return "", err
	}
	return string(bytes.ReplaceAll(data, []byte(`"`), []byte(""))), nil
}

type PipelineState struct {
	Active     bool   `json:"active"`
	ChannelID  int64  `json:"channelId"`
	PipelineID int64  `json:"pipelineId"`
	Status     string `json:"status"`
}

func (o *sOtter) GetPipelineState(channelID, piplineID int64) (state *PipelineState, err error) {
	path := fmt.Sprintf("/otter/channel/%d/%d/mainstem", channelID, piplineID)
	data, _, err := o.zkConn.Get(path)
	if err != nil {
		return
	}
	if err = json.Unmarshal(data, &state); err != nil {
		return
	}
	state.ChannelID = channelID
	return
}
func (o *sOtter) GetAllPipelineState() (pipeline []*PipelineState, err error) {
	channels, err := o.GetAllAliveChannel()
	for _, channel := range channels {
		var list []*Pipeline
		list, err = o.GetAllPipeline(channel)
		if err != nil {
			return
		}
		for _, info := range list {
			var state *PipelineState
			state, err = o.GetPipelineState(info.ChannelID, info.ID)
			if err != nil {
				return
			}
			pipeline = append(pipeline, state)
		}
	}
	return
}

func (o *sOtter) GetAllAliveChannel() (channels []int64, err error) {
	childs, _, err := o.zkConn.Children("/otter/channel")
	if err != nil {
		return
	}
	for _, child := range childs {
		var channelId int64
		channelId, err = strconv.ParseInt(child, 10, 64)
		if err != nil {
			return
		}
		channels = append(channels, channelId)
	}
	return
}
