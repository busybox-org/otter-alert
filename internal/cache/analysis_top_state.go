package cache

import (
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/sirupsen/logrus"
	"time"
)

const analysisTopStatKeyPrefix = "analysis_top_stat:"

type AnalysisTopState struct {
	ChannelID  int64         `json:"ChannelID"`  // 通道ID
	PipelineID int64         `json:"PipelineID"` // 管道ID
	Interval   time.Duration `json:"Interval"`   // 延迟时间
	LastTime   time.Time     `json:"LastTime"`   // 上一次告警时间
}

type AnalysisTopStateService struct {
	Ch chan AnalysisTopState
}

func (a *AnalysisTopStateService) Add(id, channelID int64, interval time.Duration) {
	key := fmt.Sprintf("%s%d", analysisTopStatKeyPrefix, id)
	data := AnalysisTopState{
		ChannelID:  channelID,
		PipelineID: id,
		Interval:   interval,
	}
	if !cache.Has(key) {
		a.Ch <- data
		data.LastTime = time.Now()
	}
	err := cache.Add(key, &data)
	if err != nil {
		logrus.Errorln(err)
	}
}
func (a *AnalysisTopStateService) Delete(id int64) {
	key := fmt.Sprintf("%s%d", analysisTopStatKeyPrefix, id)
	err := cache.Delete(key)
	if err != nil && err != badger.ErrKeyNotFound {
		logrus.Errorln(err)
	}
}

func (a *AnalysisTopStateService) Trigger(interval time.Duration) {
	for range time.Tick(interval) {
		for k, v := range cache.Iterator(analysisTopStatKeyPrefix) {
			go func(k string, v []byte) {
				var data = AnalysisTopState{}
				err := json.Unmarshal(v, &data)
				if err != nil {
					logrus.Errorln(err)
					return
				}
				// 第一次或间隔小于定时间隔的
				if time.Now().Sub(data.LastTime) < interval {
					return
				}
				a.Ch <- data
				data.LastTime = time.Now()
				err = cache.Add(k, &data)
				if err != nil {
					logrus.Errorln(err)
				}
			}(k, v)
		}
	}
}
