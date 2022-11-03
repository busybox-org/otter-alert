package notification

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/xmapst/otteralert/internal/utils"
	"net/http"
)

var (
	httpClient = &http.Client{
		Transport: &http.Transport{
			Proxy:             http.ProxyFromEnvironment,
			DisableKeepAlives: true,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

type Notification interface {
	SendMarkdown(title, message string)
	Send(message interface{}) error
}

func New(mode, addr, secret string) (Notification, error) {
	if addr == "" {
		return newDefault()
	}
	err := utils.CheckURL(addr)
	if err != nil {
		return nil, err
	}
	switch mode {
	case "dingtalk":
		return newDingTalk(addr, secret)
	case "qiyeweixin":
		return newQiYeWeiXin(addr, secret)
	default:
		return newDefault()
	}
}
