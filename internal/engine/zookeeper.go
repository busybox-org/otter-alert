package engine

import (
	"bytes"
	"fmt"
	"github.com/go-zookeeper/zk"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"strconv"
)

var chPrefix = "/otter/channel"
var json = jsoniter.ConfigCompatibleWithStandardLibrary

type pipelineState struct {
	Active     bool   `json:"active"`
	PipelineID int64  `json:"pipelineId"`
	Status     string `json:"status"`
}

var ChState = map[string]int{
	"START": 0,
	"STOP":  1,
	"PAUSE": 2,
	"NONE":  3,
}

func (e *Engine) getChannelState(channelID int64) string {
	path := fmt.Sprintf("%s/%d", chPrefix, channelID)
	cStatus, _, err := e.zkConn.Get(path)
	if err != nil {
		logrus.Error(err)
		return ""
	}
	return string(bytes.ReplaceAll(cStatus, []byte(`"`), []byte("")))
}

func (e *Engine) getPipelineState(cid string, pid string) (*pipelineState, error) {
	path := fmt.Sprintf("%s/%s/%s/mainstem", chPrefix, cid, pid)
	res, _, err := e.zkConn.Get(path)
	if err != nil {
		return nil, err
	}
	var state = new(pipelineState)
	err = json.Unmarshal(res, &state)
	if err != nil {
		return nil, err
	}
	return state, nil
}

func (e *Engine) getAllChannelPipelineState() {
	list, _, err := e.zkConn.Children(chPrefix)
	if err != nil {
		logrus.Error("zk连接异常,获取不到节点信息")
		return
	}
	for _, cid := range list {
		go e.allChannelState(cid)
	}
}

func (e *Engine) allChannelState(cid string) {
	channelId, err := strconv.ParseInt(cid, 10, 64)
	if err != nil {
		logrus.Errorln(err)
		return
	}
	status := e.getChannelState(channelId)
	if status == "" {
		return
	}
	if ChState[status] != 0 {
		e.channelService.Add(channelId, status)
	} else {
		e.channelService.Delete(channelId, status)
	}
	var pipelineList []string
	pipelineList, _, err = e.zkConn.Children(fmt.Sprintf("%s/%s", chPrefix, cid))
	if err != nil {
		logrus.Error(err)
		return
	}
	for _, pid := range pipelineList {
		go e.allPipelineState(cid, pid)
	}
}

func (e *Engine) allPipelineState(cid, pid string) {
	state, err := e.getPipelineState(cid, pid)
	if err != nil {
		if err != zk.ErrNoNode {
			logrus.Errorln(err)
		}
		return
	}
	if !state.Active {
		e.pipelineService.Add(state.PipelineID, state.Status)
	} else {
		e.pipelineService.Delete(state.PipelineID, state.Status)
	}
}
