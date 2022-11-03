package notification

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/xmapst/otteralert/internal/utils"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type dingTalkNotificationResponse struct {
	ErrorMessage string `json:"errmsg"`
	ErrorCode    int    `json:"errcode"`
}

type dingTalkNotification struct {
	MessageType string                          `json:"msgtype"`
	Text        *dingTalkNotificationText       `json:"text,omitempty"`
	Link        *dingTalkNotificationLink       `json:"link,omitempty"`
	Markdown    *dingTalkNotificationMarkdown   `json:"markdown,omitempty"`
	ActionCard  *dingTalkNotificationActionCard `json:"actionCard,omitempty"`
	At          *dingTalkNotificationAt         `json:"at,omitempty"`
}

type dingTalkNotificationText struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type dingTalkNotificationLink struct {
	Title      string `json:"title"`
	Text       string `json:"text"`
	MessageURL string `json:"messageUrl"`
	PictureURL string `json:"picUrl"`
}

type dingTalkNotificationMarkdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type dingTalkNotificationAt struct {
	AtMobiles []string `json:"atMobiles,omitempty"`
	IsAtAll   bool     `json:"isAtAll,omitempty"`
}

type dingTalkNotificationActionCard struct {
	Title             string                       `json:"title"`
	Text              string                       `json:"text"`
	HideAvatar        string                       `json:"hideAvatar"`
	ButtonOrientation string                       `json:"btnOrientation"`
	Buttons           []DingTalkNotificationButton `json:"btns,omitempty"`
	SingleTitle       string                       `json:"singleTitle,omitempty"`
	SingleURL         string                       `json:"singleURL"`
}

type DingTalkNotificationButton struct {
	Title     string `json:"title"`
	ActionURL string `json:"actionURL"`
}

type dingTalk struct {
	Url    *url.URL
	Secret string
}

func newDingTalk(addr, secret string) (Notification, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	return &dingTalk{
		Url:    u,
		Secret: secret,
	}, nil
}

func (d *dingTalk) SendMarkdown(title, message string) {
	var notification = &dingTalkNotification{
		MessageType: "markdown",
		Markdown: &dingTalkNotificationMarkdown{
			Title: title,
			Text:  message,
		},
		At: &dingTalkNotificationAt{
			IsAtAll: true,
		},
	}
	go d.retrySend(notification)
}

func (d *dingTalk) retrySend(message interface{}) {
	_ = utils.Retry(3, "发送消息失败", func() error {
		return d.Send(message)
	})
}

func (d *dingTalk) Send(message interface{}) error {
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
	var robotResp dingTalkNotificationResponse
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

func (d *dingTalk) generateSign() {
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
