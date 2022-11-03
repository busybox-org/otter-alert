package engine

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/xmapst/otteralert/internal/otter"
)

func (e *Engine) recoverChannel(channelID int64) (title, message string) {
	logrus.Warningln("恢复通道", channelID)
	otterApi := otter.NewOtter(e.manager.Endpoint, e.manager.Username, e.manager.Password)
	err := otterApi.Login()
	if err != nil {
		logrus.Errorln("登录Manager失败")
		title = "## Manager登录失败"
		message = title + fmt.Sprint(
			"\n- 错误: ", err.Error(),
			"\n- 请联系相关人员进行手工处理",
		)
		return
	}
	err = otterApi.StartChannel(channelID)
	if err != nil {
		logrus.Errorln("通道恢复失败", err)
		channel := e.selectChannel(channelID)
		title = fmt.Sprintf("## %s 通道恢复失败", *channel.Name)
		message = title + fmt.Sprint(
			"\n- 错误: ", err.Error(),
			"\n- 请联系相关人员进行手工处理",
		)
		return
	}
	logrus.Infof("通道%d恢复成功", channelID)
	return
}

func (e *Engine) restartChannel(channelID int64) (title, message string) {
	e.restartChannelService.Add(channelID)
	logrus.Warningln("重启通道", channelID)
	otterApi := otter.NewOtter(e.manager.Endpoint, e.manager.Username, e.manager.Password)
	err := otterApi.Login()
	if err != nil {
		logrus.Errorln("登录Manager失败")
		title = "## Manager登录失败"
		message = title + fmt.Sprint(
			"\n- 错误: ", err.Error(),
			"\n- 请联系相关人员进行手工处理",
		)
		return
	}
	err = otterApi.StopChannel(channelID)
	if err != nil {
		logrus.Errorln("通道停止失败", err)
		channel := e.selectChannel(channelID)
		title = fmt.Sprintf("## %s 通道停止失败", *channel.Name)
		message = title + fmt.Sprint(
			"\n- 错误: ", err.Error(),
			"\n- 请联系相关人员进行手工处理",
		)
		return
	}
	err = otterApi.StartChannel(channelID)
	if err != nil {
		logrus.Errorln("通道启动失败", err)
		channel := e.selectChannel(channelID)
		title = fmt.Sprintf("## %s 通道启动失败", *channel.Name)
		message = title + fmt.Sprint(
			"\n- 错误: ", err.Error(),
			"\n- 请联系相关人员进行手工处理",
		)
		return
	}
	logrus.Infof("通道%d重启成功", channelID)
	return
}
