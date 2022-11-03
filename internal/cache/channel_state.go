package cache

import (
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/sirupsen/logrus"
	"time"
)

const channelStatKeyPrefix = "channel_stat:"

type ChannelState struct {
	ChannelID int64     `json:"ChannelID"` // 通道ID
	Status    string    `json:"Status"`    // 状态
	StartTime time.Time `json:"StartTime"` // 开始时间
	LastTime  time.Time `json:"LastTime"`  // 上一次告警时间
	Active    bool      `json:"Active"`    // 是否恢复
}

type ChannelStateService struct {
	Ch chan ChannelState
}

func (c *ChannelStateService) Add(id int64, status string) {
	key := fmt.Sprintf("%s%d", channelStatKeyPrefix, id)
	data := ChannelState{
		ChannelID: id,
		Status:    status,
	}
	if !cache.Has(key) {
		data.StartTime = time.Now()
		data.LastTime = time.Now()
		c.Ch <- data
	}
	err := cache.Add(key, &data)
	if err != nil {
		logrus.Errorln(err)
	}
}
func (c *ChannelStateService) Delete(id int64, status string) {
	key := fmt.Sprintf("%s%d", channelStatKeyPrefix, id)
	res := ChannelState{}
	err := cache.Get(key, &res)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			logrus.Errorln(err)
		}
		return
	}
	err = cache.Delete(key)
	if err != nil {
		logrus.Errorln(err)
	}
	res.Active = true
	res.Status = status
	c.Ch <- res
}

func (c *ChannelStateService) Trigger(interval time.Duration) {
	for range time.Tick(interval) {
		for k, v := range cache.Iterator(channelStatKeyPrefix) {
			go func(k string, v []byte) {
				var data = ChannelState{}
				err := json.Unmarshal(v, &data)
				if err != nil {
					logrus.Errorln(err)
					return
				}
				// 第一次或间隔小于定时间隔的
				if time.Now().Sub(data.LastTime) < interval {
					return
				}
				c.Ch <- data
				data.LastTime = time.Now()
				err = cache.Add(k, &data)
				if err != nil {
					logrus.Errorln(err)
				}
			}(k, v)
		}
	}
}
