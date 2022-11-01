package dataprovider

import (
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type AnalysisTopState struct {
	FailedCh chan AFailed
	faileds  map[string]*AFailed
	// map并发安全读写锁
	lock *sync.RWMutex
}

type AFailed struct {
	ChannelID  int64
	PipelineID int64
	Interval   time.Duration
	First      bool
}

func NewAnalysisTopState() *AnalysisTopState {
	return &AnalysisTopState{
		faileds: make(map[string]*AFailed),
		lock:    &sync.RWMutex{},
	}
}

func (ats *AnalysisTopState) Add(id string, channelID, pipelineID int64, interval time.Duration) {
	ats.lock.Lock()
	defer ats.lock.Unlock()
	if ats.faileds[id] == nil {
		ats.faileds[id] = &AFailed{
			ChannelID:  channelID,
			PipelineID: pipelineID,
			First:      true,
		}
	}
	ats.faileds[id].Interval = interval
}
func (ats *AnalysisTopState) Del(id string) {
	ats.lock.Lock()
	defer ats.lock.Unlock()
	if ats.faileds[id] != nil {
		delete(ats.faileds, id)
	}
}

func (ats *AnalysisTopState) Trigger(interval time.Duration) {
	go func() {
		for {
			for id, failed := range ats.faileds {
				// 第一次触发告警
				if !failed.First {
					continue
				}
				logrus.Info("first trigger")
				ats.FailedCh <- *failed
				ats.lock.RLock()
				ats.faileds[id].First = false
				ats.lock.RUnlock()
			}
			time.Sleep(150 * time.Millisecond)
		}
	}()
	for range time.Tick(interval) {
		// 后续定期告警
		for _, failed := range ats.faileds {
			if failed.First {
				continue
			}
			logrus.Info("periodic trigger")
			ats.FailedCh <- *failed
		}
	}
}
