package wechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

func (w *Wechat) GetContacts() (err error) {

	name, resp := "webwxgetcontact", new(MemberResp)
	apiURI := fmt.Sprintf("%s/%s?pass_ticket=%s&skey=%s&r=%s", w.BaseUri, name, w.Request.PassTicket, w.Request.Skey, w.GetUnixTime())
	if err := w.Send(apiURI, nil, resp); err != nil {
		return err
	}
	w.MemberList, w.TotalMember = make([]*Member, 0, resp.MemberCount/5*2), resp.MemberCount
	for _, member := range resp.MemberList {
		if member.IsNormal(w.User.UserName) {
			w.MemberList = append(w.MemberList, member)
		}
	}
	return
}

func (w *Wechat) StatusNotify() (err error) {
	statusURL := w.BaseUri + fmt.Sprintf("/webwxstatusnotify?lang=zh_CN&pass_ticket=%s", w.Request.PassTicket)
	resp := new(NotifyResp)
	data, err := json.Marshal(NotifyParams{
		BaseRequest:  w.Request,
		Code:         3,
		FromUserName: w.User.UserName,
		ToUserName:   w.User.UserName,
		ClientMsgId:  w.GetUnixTime(),
	})

	if err := w.Send(statusURL, bytes.NewReader(data), resp); err != nil {
		return err
	}

	return
}

func (w *Wechat) GetContactsInBatch() (err error) {
	resp := new(MemberResp)
	apiUrl := fmt.Sprintf("https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxbatchgetcontact?type=ex&r=%s&pass_ticket=%s", w.GetUnixTime(), w.Request.PassTicket)
	if err := w.Send(apiUrl, nil, resp); err != nil {
		return err
	}
	return
}

func (w *Wechat) TestCheck() (err error) {
	/*for _, host := range Hosts {
		w.SyncHost = host

	}*/
	return
}

func (w *Wechat) SyncCheck() (err error) {
	//checkUrl:=fmt.Sprintf("https://%s/cgi-bin/mmwebwx-bin/synccheck?sid=%s&uin=%s&skey=%s&deviceid=%s&synckey=%s&_=%s"+, w.SyncHost,w.Sid,w.Uin,w.Skey,w.DeviceId,w.SyncKeys[0])
	return
}

func (w *Wechat) SendMsg(name, word string, isFile bool) (err error) {

	return
}

func (w *Wechat) SendMsgToAll(word string) (err error) {

	return
}

func (w *Wechat) SendImage(name, fileName string) (err error) {

	return
}

func (w *Wechat) AddMember(name string) (err error) {

	return
}

func (w *Wechat) CreateRoom(name string) (err error) {

	return
}

func (w *Wechat) PullMsg() {
	return
}

func (w *Wechat) Post(url string, data url.Values, jsonFmt bool) (result string) {
	//req.Header.Set("User-agent", UserAgent)

	resp, err := w.Client.PostForm(url, data)

	//req.Header.Set("ContentType", "application/json; charset=UTF-8")

	if err != nil {
		return
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	result = string(respBody)
	return
}

func (w *Wechat) Send(apiURI string, body io.Reader, call Caller) (err error) {
	method := "GET"
	if body != nil {
		method = "POST"
	}

	req, err := http.NewRequest(method, apiURI, body)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	resp, err := w.Client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	reader := resp.Body.(io.Reader)

	if err = json.NewDecoder(reader).Decode(call); err != nil {
		return
	}
	if !call.IsSuccess() {
		return call.Error()
	}
	return
}

func (w *Wechat) SetCookies() {
	//w.Client.Jar.SetCookies(u, cookies)

}

func (w *Wechat) GetUnixTime() int {
	return int(time.Now().Unix())
}
