package notification

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"github.com/xmapst/otter-alert/utils"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
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

type DingTalk struct {
	Url    *url.URL
	Secret string
}

func newDingTalk(addr, secret string) (Notification, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	return &DingTalk{
		Url:    u,
		Secret: secret,
	}, nil
}

func (d *DingTalk) SendMarkdown(title, message string) {
	var notification = &utils.DingTalkNotification{
		MessageType: "markdown",
		Markdown: &utils.DingTalkNotificationMarkdown{
			Title: title,
			Text:  message,
		},
		At: &utils.DingTalkNotificationAt{
			IsAtAll: true,
		},
	}
	go d.retrySend(notification)
}

func (d *DingTalk) retrySend(message interface{}) {
	_ = utils.Retry(3, "发送消息失败", func() error {
		return d.Send(message)
	})
}

func (d *DingTalk) Send(message interface{}) error {
	d.generateSign()
	body, err := json.Marshal(message)
	if err != nil {
		logrus.Error(err, "error encoding DingTalk request")
		return err
	}
	httpReq, err := http.NewRequest("POST", d.Url.String(), bytes.NewReader(body))
	if err != nil {
		logrus.Error(err, "error building DingTalk request")
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		logrus.Error(err, "error sending notification to DingTalk")
		return err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != 200 {
		logrus.Errorf("unacceptable response code %d", resp.StatusCode)
		return fmt.Errorf("unacceptable response code %d", resp.StatusCode)
	}
	var robotResp utils.DingTalkNotificationResponse
	enc := json.NewDecoder(resp.Body)
	if err = enc.Decode(&robotResp); err != nil {
		logrus.Error(err, "error decoding response from DingTalk")
		return err
	}
	if robotResp.ErrorCode != 0 {
		logrus.Error("Failed to send notification to DingTalk  respCode ", robotResp.ErrorCode, " respMsg ", robotResp.ErrorMessage)
		return fmt.Errorf("Failed to send notification to DingTalk  respCode %d respMsg %s", robotResp.ErrorCode, robotResp.ErrorMessage)
	}
	logrus.Info("message sent successfully")
	return nil
}

func (d *DingTalk) generateSign() {
	if d.Secret != "" {
		timestamp := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
		stringToSign := []byte(timestamp + "\n" + d.Secret)
		mac := hmac.New(sha256.New, []byte(d.Secret))
		mac.Write(stringToSign) // nolint: errcheck
		signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
		qs := d.Url.Query()
		qs.Set("timestamp", timestamp)
		qs.Set("sign", signature)
		d.Url.RawQuery = qs.Encode()
	}
}
