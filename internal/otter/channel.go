package otter

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
)

func (o *Otter) StartChannel(id int64) error {
	return o.channel(id, "start")
}

func (o *Otter) StopChannel(id int64) error {
	return o.channel(id, "stop")
}

func (o *Otter) channel(id int64, action string) error {
	_url := fmt.Sprintf("%s/?action=channelAction&channelId=%d&status=%s&pageIndex=1&searchKey=&eventSubmitDoStatus=true", o.Endpoint, id, action)
	req, err := http.NewRequest(http.MethodGet, _url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Cookie", o.cookieStr)
	resp, err := o.client.Do(req)
	if err != nil {
		logrus.Fatalln(err)
	}
	defer resp.Body.Close()
	return nil
}
