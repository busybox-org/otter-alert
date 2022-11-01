package engine

import (
	"bytes"
	"fmt"
	"github.com/go-zookeeper/zk"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"github.com/xmapst/otter-alert/dataprovider"
	"strconv"
)

var chPrefix = "/otter/channel"
var json = jsoniter.ConfigCompatibleWithStandardLibrary

type ChannelPipelineState struct {
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
	chPath := fmt.Sprintf("%s/%d", chPrefix, channelID)
	cStatus, _, err := e.zkConn.Get(chPath)
	if err != nil {
		logrus.Error(err)
		return ""
	}
	return string(bytes.ReplaceAll(cStatus, []byte(`"`), []byte("")))
}

func (e *Engine) getAllChannelState(db *dataprovider.ChannelState) {
	list, _, err := e.zkConn.Children(chPrefix)
	if err != nil {
		logrus.Error("zk连接异常,获取不到节点信息")
		return
	}
	for _, cID := range list {
		chPath := chPrefix + "/" + cID
		var cStatus []byte
		cStatus, _, err = e.zkConn.Get(chPath)
		if err != nil {
			logrus.Error(err)
			continue
		}
		channelId, _ := strconv.Atoi(cID)
		status := string(bytes.ReplaceAll(cStatus, []byte(`"`), []byte("")))
		if ChState[status] != 0 {
			db.Add(cID, int64(channelId), 0, status)
		} else {
			db.Del(cID, status)
		}
		var pipelineList []string
		pipelineList, _, err = e.zkConn.Children(chPath)
		if err != nil {
			logrus.Error(err)
			continue
		}
		for _, pID := range pipelineList {
			var cps = ChannelPipelineState{}
			var res []byte
			res, _, err = e.zkConn.Get(chPath + "/" + pID + "/" + "mainstem")
			if err != nil {
				if err != zk.ErrNoNode {
					logrus.Error(err)
				}
				continue
			}
			err = json.Unmarshal(res, &cps)
			if err != nil {
				logrus.Error(err)
				continue
			}
			pid := fmt.Sprintf("c_%s-p_%d", cID, cps.PipelineID)
			if !cps.Active {
				db.Add(pid, int64(channelId), cps.PipelineID, cps.Status)
			} else {
				db.Del(pid, cps.Status)
			}
		}
	}
}
