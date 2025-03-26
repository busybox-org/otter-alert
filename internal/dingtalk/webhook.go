package dingtalk

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/imroc/req/v3"
	"github.com/xmapst/logx"
)

var (
	client = req.NewClient().
		EnableHTTP3().
		ImpersonateChrome().
		EnableDumpAllAsync().
		SetBaseURL(dingTalkURL).
		SetLogger(logx.GetSubLogger())
)

func InitDingTalk(tokens []string, key string, opts ...InitOption) *DingTalk {
	if len(tokens) == 0 {
		panic("no token")
	}
	dt := &DingTalk{
		robotToken: tokens,
		keyWord:    key,
		http:       client,
	}
	for _, opt := range opts {
		opt.applyInit(&dt.InitModel)
	}
	return dt
}

func InitDingTalkWithSecret(tokens string, secret string, opts ...InitOption) *DingTalk {
	if len(tokens) == 0 || secret == "" {
		panic("no token")
	}
	dt := &DingTalk{
		robotToken: []string{tokens},
		secret:     secret,
		http:       client,
	}
	for _, opt := range opts {
		opt.applyInit(&dt.InitModel)
	}
	return dt
}

func (d *DingTalk) sendMessageNoCtx(msg iDingMsg) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.InitModel.GetSendTimeout())
	defer cancel()
	return d.sendMessage(ctx, msg)
}

func (d *DingTalk) sendMessage(ctx context.Context, msg iDingMsg) error {
	_req := d.http.NewRequest().
		SetContext(ctx).
		SetHeaders(map[string]string{
			"Content-type": "application/json",
		}).
		SetQueryParam("access_token", d.robotToken[rand.Intn(len(d.robotToken))])
	if d.secret != "" {
		timestamp := time.Now().UnixNano() / 1e6
		_req.SetQueryParamsAnyType(map[string]any{
			"timestamp": timestamp,
			"sign":      d.sign(timestamp, d.secret),
		})

	}
	resp, err := _req.SetBodyBytes(msg.Marshaler()).Post("")
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("send msg err. http code: %d, token: %s, msg: %s", resp.StatusCode, d.robotToken, msg.Marshaler())
	}
	var respMsg responseMsg
	err = resp.UnmarshalJson(&respMsg)
	if err != nil {
		return err
	}
	if respMsg.ErrCode != 0 {
		return fmt.Errorf("send msg err. err msg: %s", respMsg.ErrMsg)
	}
	return nil
}

func (d *DingTalk) sign(t int64, secret string) string {
	strToHash := fmt.Sprintf("%d\n%s", t, secret)
	hmac256 := hmac.New(sha256.New, []byte(secret))
	hmac256.Write([]byte(strToHash))
	data := hmac256.Sum(nil)
	return base64.StdEncoding.EncodeToString(data)
}

func (d *DingTalk) OutGoing(r io.Reader) (outGoingMsg OutGoingModel, err error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return
	}
	err = json.Unmarshal(buf, &outGoingMsg)
	return
}

func (d *DingTalk) SendTextMessage(content string, opt ...atOption) error {
	content = content + d.keyWord
	return d.sendMessageNoCtx(NewTextMsg(content, opt...))
}

func (d *DingTalk) SendMarkDownMessage(title, text string, opts ...atOption) error {
	title = title + d.keyWord
	return d.sendMessageNoCtx(NewMarkDownMsg(title, text, opts...))
}

// SendDTMDMessage 利用dtmd发送点击消息
func (d *DingTalk) SendDTMDMessage(title string, dtmdMap *dingMap, opt ...atOption) error {
	title = title + d.keyWord
	return d.sendMessageNoCtx(NewDTMDMsg(title, dtmdMap, opt...))
}

func (d *DingTalk) SendMarkDownMessageBySlice(title string, textList []string, opts ...atOption) error {
	title = title + d.keyWord
	text := ""
	for _, t := range textList {
		text = text + "\n" + t
	}
	return d.sendMessageNoCtx(NewMarkDownMsg(title, text, opts...))
}

func (d *DingTalk) SendLinkMessage(title, text, picUrl, msgUrl string) error {
	title = title + d.keyWord
	return d.sendMessageNoCtx(NewLinkMsg(title, text, picUrl, msgUrl))
}

func (d *DingTalk) SendActionCardMessage(title, text string, opts ...actionCardOption) error {
	title = title + d.keyWord
	return d.sendMessageNoCtx(NewActionCardMsg(title, text, opts...))
}

func (d *DingTalk) SendActionCardMessageBySlice(title string, textList []string, opts ...actionCardOption) error {
	title = title + d.keyWord
	text := ""
	for _, t := range textList {
		text = text + "\n" + t
	}
	return d.sendMessageNoCtx(NewActionCardMsg(title, text, opts...))
}

func (d *DingTalk) SendFeedCardMessage(feedCard []FeedCardLinkModel) error {
	if len(feedCard) > 0 {
		feedCard[0].Title = feedCard[0].Title + d.keyWord
	}
	return d.sendMessageNoCtx(NewFeedCardMsg(feedCard))
}
