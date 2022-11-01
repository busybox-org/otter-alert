package otter

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (o *Otter) Login() error {
	params := url.Values{}
	params.Add("action", `user_action`)
	params.Add("event_submit_do_login", `1`)
	params.Add("_fm.l._0.n", o.Username)
	params.Add("_fm.l._0.p", o.Password)
	body := strings.NewReader(params.Encode())

	req, err := http.NewRequest(http.MethodPost, o.Endpoint+"/login.htm", body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := o.client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	if resp.StatusCode != http.StatusFound {
		return fmt.Errorf("failed to login")
	}
	var JSESSIONID, OtterWebXJsessionid0 string
	for _, v := range resp.Cookies() {
		if v.Name == "JSESSIONID" {
			JSESSIONID = fmt.Sprintf("%s=%s", v.Name, v.Value)
		}
		if v.Name == "OTTER_WEBX_JSESSIONID0" {
			OtterWebXJsessionid0 = fmt.Sprintf("%s=%s", v.Name, v.Value)
		}
	}
	if JSESSIONID == "" || OtterWebXJsessionid0 == "" {
		return fmt.Errorf("failed to login")
	}
	o.cookieStr = JSESSIONID + ";" + OtterWebXJsessionid0
	return nil
}
