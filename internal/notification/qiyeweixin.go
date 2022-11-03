package notification

import (
	"bytes"
	"errors"
	"github.com/xmapst/otteralert/internal/utils"
	"io"
	"net/http"
	"net/url"
)

type markdownMessage struct {
	MsgType  string   `json:"msgtype"`
	Markdown Markdown `json:"markdown"`
}

type Markdown struct {
	Content string `json:"content"`
}

type wxWorkResponse struct {
	ErrorCode    int    `json:"errcode"`
	ErrorMessage string `json:"errmsg"`
}

type qiyeweixin struct {
	Url    *url.URL
	Secret string
}

func (q *qiyeweixin) SendMarkdown(_, message string) {
	var notification = &markdownMessage{
		MsgType: "markdown",
		Markdown: Markdown{
			Content: message,
		},
	}
	go q.retrySend(notification)
}

func (q *qiyeweixin) retrySend(message interface{}) {
	_ = utils.Retry(3, "发送消息失败", func() error {
		return q.Send(message)
	})
}

func (q *qiyeweixin) Send(message interface{}) error {
	msgBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, q.Url.String(), bytes.NewBuffer(msgBytes))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json;charset=utf-8")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	body, _ := io.ReadAll(resp.Body)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	var wxWorkResp wxWorkResponse
	err = json.Unmarshal(body, &wxWorkResp)
	if err != nil {
		return err
	}
	if wxWorkResp.ErrorCode != 0 && wxWorkResp.ErrorMessage != "" {
		return errors.New(string(body))
	}
	return nil
}

func newQiYeWeiXin(addr, secret string) (Notification, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	return &qiyeweixin{
		Url:    u,
		Secret: secret,
	}, nil
}
