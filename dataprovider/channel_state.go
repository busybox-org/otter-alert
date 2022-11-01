package dataprovider

import (
	"sync"
	"time"
)

type ChannelState struct {
	FailedCh chan CFailed
	faileds  map[string]*CFailed
	// map并发安全读写锁
	lock *sync.RWMutex
}

type CFailed struct {
	ChannelID  int64     // 通道ID
	PipelineID int64     // 流水线ID
	StartTime  int64     // 开始时间
	First      bool      // 第一次
	Status     string    // 状态
	LastTime   time.Time // 上一次告警时间
	Active     bool      // 是否恢复
}

func NewChannelState() *ChannelState {
	return &ChannelState{
		faileds: make(map[string]*CFailed),
		lock:    &sync.RWMutex{},
	}
}

func (cs *ChannelState) Add(id string, channelID, pipelineID int64, status string) {
	cs.lock.Lock()
	defer cs.lock.Unlock()
	if cs.faileds[id] == nil {
		cs.faileds[id] = &CFailed{
			ChannelID:  channelID,
			PipelineID: pipelineID,
			StartTime:  time.Now().Unix(),
			First:      true,
			Status:     status,
		}
		return
	}
}
func (cs *ChannelState) Del(id string, status string) {
	cs.lock.Lock()
	defer cs.lock.Unlock()
	if cs.faileds[id] != nil {
		defer delete(cs.faileds, id)
		failed := cs.faileds[id]
		cs.FailedCh <- CFailed{
			ChannelID:  failed.ChannelID,
			PipelineID: failed.PipelineID,
			StartTime:  failed.StartTime,
			First:      failed.First,
			LastTime:   failed.LastTime,
			Active:     true,   // 已恢复
			Status:     status, // 当前状态
		}
	}
}

func (cs *ChannelState) Trigger(interval time.Duration) {
	go func() {
		for {
			for id, failed := range cs.faileds {
				// 第一次触发告警
				if !failed.First {
					continue
				}
				cs.FailedCh <- *failed
				cs.lock.RLock()
				cs.faileds[id].First = false
				cs.faileds[id].LastTime = time.Now()
				cs.lock.RUnlock()
			}
			time.Sleep(150 * time.Millisecond)
		}
	}()
	for range time.Tick(interval) {
		// 后续定期告警
		for _, failed := range cs.faileds {
			// 第一次或间隔小于定时间隔的
			if failed.First || time.Now().Sub(failed.LastTime) < interval {
				continue
			}
			cs.FailedCh <- *failed
		}
	}
}
