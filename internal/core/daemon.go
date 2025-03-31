package core

import (
	"fmt"
	"time"

	"github.com/kardianos/service"
	"github.com/robfig/cron/v3"
	"github.com/spf13/pflag"
	"github.com/xmapst/logx"

	"github.com/busybox-org/otteralert/internal/dingtalk"
	"github.com/busybox-org/otteralert/internal/otter"
)

type sProgram struct {
	flags     *pflag.FlagSet
	cron      *cron.Cron
	alert     *dingtalk.DingTalk
	otter     otter.IOtter
	alertData map[string]int
}

func New(flags *pflag.FlagSet) service.Interface {
	p := &sProgram{
		flags:     flags,
		alertData: make(map[string]int),
		cron: cron.New(cron.WithParser(cron.NewParser(
			cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
		))),
	}
	alertAk := p.flags.Lookup("alert_ak").Value.String()
	p.alert = dingtalk.InitDingTalk([]string{alertAk}, "")
	alertSK := p.flags.Lookup("alert_sk").Value.String()
	if len(alertSK) != 0 {
		p.alert = dingtalk.InitDingTalkWithSecret(alertAk, alertSK)
	}
	return p
}

func (p *sProgram) Start(s service.Service) (err error) {
	p.otter, err = otter.New(
		p.flags.Lookup("manager.endpoint").Value.String(),
		p.flags.Lookup("manager.zookeeper").Value.String(),
		p.flags.Lookup("manager.database").Value.String(),
	)
	if err != nil {
		return err
	}
	p.cron.Start()
	spec := p.flags.Lookup("cron").Value.String()
	_, err = p.cron.AddFunc(spec, func() {
		var channelIDs []int64
		channelIDs, err = p.otter.GetAllAliveChannel()
		if err != nil {
			return
		}
		for _, channelID := range channelIDs {
			state, err := p.otter.GetChannelState(channelID)
			if err != nil {
				logx.Errorf("获取channel状态失败: %v", err)
				continue
			}
			channel, err := p.otter.GetChannel(channelID)
			if err != nil {
				logx.Errorf("获取channel失败: %v", err)
				continue
			}
			piplines, err := p.otter.GetAllPipeline(channelID)
			if err != nil {
				logx.Errorf("获取channel pipeline失败: %v", err)
				continue
			}
			for _, pipline := range piplines {
				binlog, err := p.otter.GetBinlog(pipline)
				if err != nil {
					logx.Errorf("获取binlog失败: %v", err)
					continue
				}
				delaystat, err := p.otter.GetDelayStat(pipline.ID)
				if err != nil {
					logx.Errorf("获取delaystat失败: %v", err)
					continue
				}
				delayTime := fmt.Sprintf("%.2f", float32(delaystat.DelayTtime)/float32(1000))
				syncTime := delaystat.GMTCreate.In(time.Local).Format("2006-01-02 15:04:05")
				// 延迟过大检查
				if state == "START" && len(state) != 0 && state != "STOP" {
					if len(channel.Name) != 0 && delaystat.DelayTtime > 60000 {
						logx.Infof("ChannelName:%s ChannelId:%d State:%s Delay:%ss 延迟过大", channel.Name, channelID, state, delayTime)
						p.alertDelay(channel.Name, state, delayTime, syncTime, binlog)
						p.alertData[channel.Name] = 1
					} else {
						if aType, ok := p.alertData[channel.Name]; ok {
							delete(p.alertData, channel.Name)
							p.alertResolve(channel.Name, state, delayTime, syncTime, binlog, aType)
						}
					}
				}
				// 挂起检查
				if state != "START" && len(state) != 0 && state != "STOP" {
					if len(channel.Name) != 0 {
						log, err := p.otter.GetLogRecord(channelID)
						if err != nil {
							log = fmt.Sprintf("获取日志失败 %v", err)
						}
						logx.Infof("ChannelName:%s ChannelId:%d State:%s Log:%s", channel.Name, channelID, state, log)
						p.alertSend(channel.Name, state, delayTime, syncTime, binlog, pipline, log)
						p.alertData[channel.Name+"_pause"]++
						go func() {
							for i := 0; i < 3; i++ {
								err = p.otter.StartChannel(channelID)
								if err == nil {
									return
								}
								time.Sleep(5 * time.Second)
							}
						}()
					}
				} else {
					if atype, ok := p.alertData[channel.Name+"_pause"]; ok {
						// 删除map数据发送恢复正常信息
						delete(p.alertData, channel.Name+"_pause")
						p.alertResolve(channel.Name, state, delayTime, syncTime, binlog, atype)
					}
					logx.Infof("ChannelName:%s ChannelId:%d State:%s Delay:%ss", channel.Name, channelID, state, delayTime)
				}
			}
		}
	})
	if err != nil {
		return err
	}
	return
}

func (p *sProgram) alertSend(channelName, state, delayTime, syncTime string, binlog *otter.Binlog, pipeline *otter.Pipeline, log string) {
	dm := dingtalk.DingMap()
	dm.Set(fmt.Sprintf("通道: %s状态异常", channelName), dingtalk.H2)
	dm.Set(fmt.Sprintf("状态: %s", state), dingtalk.RED)
	dm.Set(fmt.Sprintf("延迟: %ss", delayTime), dingtalk.N)
	dm.Set(fmt.Sprintf("源库: %s:%d", binlog.Identity.SourceAddress.Address, binlog.Identity.SourceAddress.Port), dingtalk.N)
	dm.Set(fmt.Sprintf("位点: %s | %d", binlog.Postion.JournalName, binlog.Postion.Position), dingtalk.N)
	dm.Set(fmt.Sprintf("位点时间: %s", binlog.Postion.TimeString), dingtalk.N)
	dm.Set(fmt.Sprintf("最后同步: %s", syncTime), dingtalk.N)
	dm.Set(fmt.Sprintf("报警时间: %s", time.Now().Format("2006-01-02 15:04:05")), dingtalk.N)
	dm.Set(fmt.Sprintf("错误日志:\n```\n%s\n```\n", log), dingtalk.N)
	dm.Set(fmt.Sprintf("[日志传送门](%s/log_record_tab.htm?pipelineId=%d)", p.otter.WebUrl(), pipeline.ID), dingtalk.N)
	dm.Set(fmt.Sprintf("自动解挂任务已开启"), dingtalk.BLUE)
	_ = p.alert.SendMarkDownMessageBySlice("otter同步异常", dm.Slice())
}

func (p *sProgram) alertDelay(channelName, state, delayTime, syncTime string, binlog *otter.Binlog) {
	dm := dingtalk.DingMap()
	dm.Set(fmt.Sprintf("通道: %s延迟过大", channelName), dingtalk.H2)
	dm.Set(fmt.Sprintf("状态: %s", state), dingtalk.RED)
	dm.Set(fmt.Sprintf("时间: %s", time.Now().Format("2006-01-02 15:04:05")), dingtalk.N)
	dm.Set(fmt.Sprintf("延迟: %ss", delayTime), dingtalk.N)
	dm.Set(fmt.Sprintf("源库: %s:%d", binlog.Identity.SourceAddress.Address, binlog.Identity.SourceAddress.Port), dingtalk.N)
	dm.Set(fmt.Sprintf("位点: %s | %d", binlog.Postion.JournalName, binlog.Postion.Position), dingtalk.N)
	dm.Set(fmt.Sprintf("位点时间: %s", binlog.Postion.TimeString), dingtalk.N)
	dm.Set(fmt.Sprintf("最后同步: %s", syncTime), dingtalk.N)
}

func (p *sProgram) alertResolve(channelName, state, delayTime, syncTime string, binlog *otter.Binlog, aType int) {
	dm := dingtalk.DingMap()
	if aType == 1 {
		dm.Set(fmt.Sprintf("通道: %s 延迟过大恢复正常", channelName), dingtalk.H2)
	} else {
		dm.Set(fmt.Sprintf("通道: %s恢复正常", channelName), dingtalk.H2)
	}
	dm.Set(fmt.Sprintf("状态: %s", state), dingtalk.GREEN)
	dm.Set(fmt.Sprintf("延迟: %ss", delayTime), dingtalk.N)
	dm.Set(fmt.Sprintf("源库: %s:%d", binlog.Identity.SourceAddress.Address, binlog.Identity.SourceAddress.Port), dingtalk.N)
	dm.Set(fmt.Sprintf("位点: %s | %d", binlog.Postion.JournalName, binlog.Postion.Position), dingtalk.N)
	dm.Set(fmt.Sprintf("位点时间: %s", binlog.Postion.TimeString), dingtalk.N)
	dm.Set(fmt.Sprintf("最后同步: %s", syncTime), dingtalk.N)
	dm.Set(fmt.Sprintf("恢复时间: %s", time.Now().Format("2006-01-02 15:04:05")), dingtalk.N)
	_ = p.alert.SendMarkDownMessageBySlice("otter恢复正常", dm.Slice())
}

func (p *sProgram) Stop(service.Service) error {
	p.cron.Stop()
	return nil
}
