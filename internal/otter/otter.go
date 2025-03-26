package otter

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/imroc/req/v3"
	"github.com/xmapst/logx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type IOtter interface {
	WebUrl() string
	StartChannel(channelID int64) error
	StopChannel(channelID int64) error
	GetChannel(channelID int64) (channel *Channel, err error)
	GetAllAliveChannel() (channels []int64, err error)
	GetAllPipeline(channelID int64) (pipelines []*Pipeline, err error)
	GetPipeline(channelID, piplineID int64) (pipeline *Pipeline, err error)
	GetDelayStat(pipelineId int64) (delayStat *DelayStat, err error)
	GetLogRecord(channelID int64) (log string, err error)
	GetNode(nodeID int64) (node *Node, err error)
	GetAllNode() (nodes []*Node, err error)
	GetBinlog(info *Pipeline) (*Binlog, error)
	GetChannelState(channelID int64) (string, error)
	GetPipelineState(channelID, piplineID int64) (state *PipelineState, err error)
	GetAllPipelineState() (pipeline []*PipelineState, err error)
}

type sOtter struct {
	client    *req.Client
	zkConn    *zk.Conn
	dbConn    *gorm.DB
	cookies   string
	baseUrl   string
	webUser   string
	webPasswd string
}

// New otter interfac
// New otter interface
// endpoint: http(s)://username:password@domain:port
// zk: host1:port,host2:port,host3:port
// db: user:password@tcp(host:port)/database?charset=utf8&parseTime=True&loc=Local
func New(httpSni, zkSni, dbSni string) (IOtter, error) {
	o := &sOtter{
		client: req.NewClient().
			EnableHTTP3().
			EnableDumpAllAsync().
			ImpersonateChrome().
			SetLogger(logx.GetSubLogger()).
			SetRedirectPolicy(req.NoRedirectPolicy()),
	}
	_url, err := url.Parse(httpSni)
	if err != nil {
		logx.Errorf("parse url error: %v", err)
		return nil, err
	}
	if _url.User == nil {
		return nil, fmt.Errorf("missing username and password")
	}
	o.webUser = _url.User.Username()
	o.webPasswd, _ = _url.User.Password()
	_url.User = nil
	o.client.SetBaseURL(_url.String())
	if err = o.login(); err != nil {
		return nil, err
	}
	if o.zkConn, _, err = zk.Connect(
		strings.Split(zkSni, ","),
		time.Second*60,
		zk.WithLogger(logx.GetSubLogger()),
	); err != nil {
		logx.Errorf("zk connect error: %v", err)
		return nil, err
	}
	if o.dbConn, err = gorm.Open(mysql.Open(dbSni), &gorm.Config{
		Logger: logger.New(logx.GetSubLogger(), logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: false,
			Colorful:                  false,
		}),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	}); err != nil {
		logx.Errorf("db connect error: %v", err)
		return nil, err
	}
	return o, nil
}

func (o *sOtter) WebUrl() string {
	return o.baseUrl
}

func (o *sOtter) login() error {
	request := o.client.R().SetHeaders(map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}).SetFormDataFromValues(
		url.Values{
			"action": {
				"user_action",
			},
			"event_submit_do_login": {
				"1",
			},
			"_fm.l._0.n": {
				o.webUser,
			},
			"_fm.l._0.p": {
				o.webPasswd,
			},
		},
	)
	response, err := request.Post("/login.htm")
	if err != nil {
		logx.Errorf("login error: %v", err)
		return err
	}
	if response.StatusCode != http.StatusFound {
		return fmt.Errorf("unacceptable response code %d", response.StatusCode)
	}
	var cook []string
	for _, cookie := range response.Cookies() {
		cook = append(cook, cookie.Name+"="+cookie.Value)
	}
	if len(cook) == 0 {
		return fmt.Errorf("missing cookies")
	}
	o.cookies = strings.Join(cook, ";")
	o.baseUrl = strings.TrimSuffix(response.Header.Get("Location"), "/channel_list.htm")
	return nil
}

func (o *sOtter) channelManager(id int64, action string) error {
	request := o.client.R().SetQueryParamsAnyType(map[string]any{
		"action":              "channelAction",
		"channelId":           id,
		"status":              action,
		"pageIndex":           1,
		"searchKey":           "",
		"eventSubmitDoStatus": true,
	}).SetHeaders(map[string]string{
		"Cookie": o.cookies,
	})
	response, err := request.Get("")
	if err != nil {
		logx.Errorf("channelManager error: %v", err)
		return err
	}
	if response.StatusCode == http.StatusFound {
		return nil
	}
	return fmt.Errorf("unacceptable response code %d", response.StatusCode)
}

func (o *sOtter) StartChannel(id int64) error {
	return o.channelManager(id, "start")
}

func (o *sOtter) StopChannel(id int64) error {
	return o.channelManager(id, "stop")
}
