package wechat

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
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

func (w *Wechat) getSyncMsg() (msgs []Message, err error) {
	name := "webwxsync"
	syncResp := new(SyncResp)
	url := fmt.Sprintf("%s/%s?sid=%s&pass_ticket=%s&skey=%s", w.BaseUri, name, w.Request.Wxsid, w.Request.PassTicket, w.Request.Skey)
	params := SyncParams{
		BaseRequest: *w.Request,
		SyncKey:     w.SyncKeyStr,
		RR:          ^time.Now().Unix(),
	}
	data, err := json.Marshal(params)

	w.Log.Println(url)
	w.Log.Println(string(data))

	if err := w.Send(url, bytes.NewReader(data), syncResp); err != nil {
		w.Log.Printf("w.Send(%s,%s,%+v) with error:%v", url, string(data), syncResp, err)
		return nil, err
	}
	w.SyncKeys = syncResp.SyncKey.List
	w.SyncKeyStr = ""
	for i, item := range w.SyncKeys {
		if i == 0 {
			w.SyncKeyStr = strconv.Itoa(item.Key) + "_" + strconv.Itoa(item.Val)
			continue
		}
		w.SyncKeyStr += "|" + strconv.Itoa(item.Key) + "_" + strconv.Itoa(item.Val)
	}

	msgs = syncResp.AddMsgList
	return
}

func (w *Wechat) SyncDaemon(msgIn chan Message) {
	for {

		resp, err := w.SyncCheck()
		if err != nil {
			w.Log.Printf("w.SyncCheck() with error:%+v\n", err)
			continue
		}
		switch resp.RetCode {
		case 1100:
			w.Log.Println("从微信上登出")
		case 1101:
			w.Log.Println("从其他设备上登陆")
			break
		case 0:
			switch resp.Selector {
			case 2, 3: //有消息,未知
				msgs, err := w.getSyncMsg()
				w.Log.Printf("the msgs:%+v\n", msgs)

				if err != nil {
					w.Log.Printf("w.getSyncMsg() error:%+v\n", err)
				}

				for _, msg := range msgs {
					msgIn <- msg
				}
			case 4: //通讯录更新
				w.GetContacts()
			case 6: //可能是红包
				w.Log.Println("请速去手机抢红包")
			case 7:
				w.Log.Println("在手机上操作了微信")
			case 0:
				w.Log.Println("无事件")
			}
		default:
			w.Log.Printf("the resp:%+v", resp)
			continue
		}

		time.Sleep(time.Second * 4)
	}
}

func (w *Wechat) MsgDaemon(msgOut chan MessageOut) {
	msg := MessageOut{}
	for {
		select {
		case msg = <-msgOut:
			w.Log.Printf("the msg to send %+v", msg)
		}
	}
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
	w.SyncHost = SyncHosts[0]

	return
}

func (w *Wechat) SyncCheck() (resp SyncCheckResp, err error) {
	params := url.Values{}
	curTime := strconv.FormatInt(time.Now().Unix(), 10)
	params.Set("r", curTime)
	params.Set("sid", w.Request.Wxsid)
	params.Set("uin", strconv.FormatInt(int64(w.Request.Wxuin), 10))
	params.Set("skey", w.Request.Skey)
	params.Set("deviceid", w.Request.DeviceID)
	params.Set("synckey", w.SyncKeyStr)
	params.Set("_", curTime)
	checkUrl := fmt.Sprintf("https://%s/cgi-bin/mmwebwx-bin/synccheck", w.SyncHost)
	Url, err := url.Parse(checkUrl)
	if err != nil {
		return
	}
	Url.RawQuery = params.Encode()
	w.Log.Println(Url.String())

	ret, err := w.Client.Get(Url.String())
	if err != nil {
		w.Log.Printf("the error is :%+v", err)
		return
	}
	defer ret.Body.Close()

	body, err := ioutil.ReadAll(ret.Body)

	if err != nil {
		return
	}
	w.Log.Println(string(body))
	resp = SyncCheckResp{}
	reRedirect := regexp.MustCompile(`window.synccheck={retcode:"(\d+)",selector:"(\d+)"}`)
	pmSub := reRedirect.FindStringSubmatch(string(body))
	w.Log.Printf("the data:%+v", pmSub)
	if len(pmSub) != 0 {
		resp.RetCode, err = strconv.Atoi(pmSub[1])
		resp.Selector, err = strconv.Atoi(pmSub[2])
		w.Log.Printf("the resp:%+v", resp)

	} else {
		err = errors.New("regex error in window.redirect_uri")
		return
	}
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
		w.Log.Printf("the error:%+v", err)
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
