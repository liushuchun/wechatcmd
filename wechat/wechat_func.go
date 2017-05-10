package wechat

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func (w *Wechat) GetContacts() (err error) {

	name, resp := "webwxgetcontact", new(MemberResp)
	apiURI := fmt.Sprintf("%s/%s?pass_ticket=%s&skey=%s&r=%s", w.BaseUri, name, w.Request.PassTicket, w.Request.Skey, w.GetUnixTime())
	if err := w.Send(apiURI, nil, resp); err != nil {
		return err
	}

	w.MemberList = resp.MemberList
	w.TotalMember = resp.MemberCount
	for _, member := range w.MemberList {
		w.MemberMap[member.UserName] = member
		if member.UserName[:2] == "@@" {
			w.GroupMemberList = append(w.GroupMemberList, member) //群聊

		} else if member.VerifyFlag&8 != 0 {
			w.PublicUserList = append(w.PublicUserList, member) //公众号
		} else if member.UserName[:1] == "@" {
			w.ContactList = append(w.ContactList, member)
		}
	}
	mb := Member{}
	mb.NickName = w.User.NickName
	mb.UserName = w.User.UserName
	w.MemberMap[w.User.UserName] = mb
	for _, user := range w.ChatSet {
		exist := false
		for _, initUser := range w.InitContactList {
			if user == initUser.UserName {
				exist = true
				break
			}
		}
		if !exist {
			value, ok := w.MemberMap[user]
			if ok {
				contact := User{
					UserName:  value.UserName,
					NickName:  value.NickName,
					Signature: value.Signature,
				}

				w.InitContactList = append(w.InitContactList, contact)
			}
		}

	}

	return
}

func (w *Wechat) getWechatRoomMember(roomID, userId string) (roomName, userName string, err error) {
	apiUrl := fmt.Sprintf("%s/webwxbatchgetcontact?type=ex&r=%s&pass_ticket=%s", w.BaseUri, w.GetUnixTime(), w.Request.PassTicket)
	params := make(map[string]interface{})
	params["BaseRequest"] = *w.Request
	params["Count"] = 1
	params["List"] = []map[string]string{}
	l := []map[string]string{}
	params["List"] = append(l, map[string]string{
		"UserName":   roomID,
		"ChatRoomId": "",
	})
	fmt.Println(apiUrl, params)

	return "", "", nil
}

func (w *Wechat) getSyncMsg() (msgs []interface{}, err error) {
	name := "webwxsync"
	syncResp := new(SyncResp)
	url := fmt.Sprintf("%s/%s?sid=%s&pass_ticket=%s&skey=%s", w.BaseUri, name, w.Request.Wxsid, w.Request.PassTicket, w.Request.Skey)
	params := SyncParams{
		BaseRequest: *w.Request,
		SyncKey:     w.SyncKey,
		RR:          ^time.Now().Unix(),
	}
	data, err := json.Marshal(params)

	w.Log.Println(url)
	w.Log.Println(string(data))

	if err := w.Send(url, bytes.NewReader(data), syncResp); err != nil {
		w.Log.Printf("w.Send(%s,%s,%+v) with error:%v", url, string(data), syncResp, err)
		return nil, err
	}
	if syncResp.BaseResponse.Ret == 0 {
		w.SyncKey = syncResp.SyncKey
		w.SyncKeyStr = ""
		for i, item := range w.SyncKey.List {
			if i == 0 {
				w.SyncKeyStr = strconv.Itoa(item.Key) + "_" + strconv.Itoa(item.Val)
				continue
			}
			w.SyncKeyStr += "|" + strconv.Itoa(item.Key) + "_" + strconv.Itoa(item.Val)
		}
	}

	msgs = syncResp.AddMsgList
	return
}

//同步守护goroutine
func (w *Wechat) SyncDaemon(msgIn chan Message) {
	for {
		w.lastCheckTs = time.Now()
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

				for _, m := range msgs {
					msg := Message{}
					msgType := m.(map[string]interface{})["MsgType"].(float64)
					msg.MsgType = int(msgType)
					msg.FromUserName = m.(map[string]interface{})["FromUserName"].(string)
					if nickNameFrom, ok := w.MemberMap[msg.FromUserName]; ok {
						msg.FromUserNickName = nickNameFrom.NickName
					}

					msg.ToUserName = m.(map[string]interface{})["ToUserName"].(string)
					if nickNameTo, ok := w.MemberMap[msg.ToUserName]; ok {
						msg.ToUserNickName = nickNameTo.NickName
					}

					msg.Content = m.(map[string]interface{})["Content"].(string)
					msg.Content = strings.Replace(msg.Content, "&lt;", "<", -1)
					msg.Content = strings.Replace(msg.Content, "&gt;", ">", -1)
					msg.Content = strings.Replace(msg.Content, " ", " ", 1)
					switch msg.MsgType {
					case 1:

						if msg.FromUserName[:2] == "@@" {
							//群消息，暂时不处理
							if msg.FromUserNickName == "" {
								contentSlice := strings.Split(msg.Content, ":<br/>")
								msg.Content = contentSlice[1]

							}
						} else {
							if w.AutoReply {
								w.SendMsg(msg.FromUserName, w.AutoReplyMsg(), false)
							}
						}
						if msg.ToUserNickName == "" {
							if user, ok := w.MemberMap[msg.ToUserName]; ok {
								msg.ToUserNickName = user.NickName
							}

						}
						if msg.FromUserNickName == "" {
							if user, ok := w.MemberMap[msg.FromUserNickName]; ok {
								msg.FromUserNickName = user.NickName
							}
						}
						msgIn <- msg
					case 3:
						//图片
					case 34:
						//语音
					case 47:
						//动画表情
					case 49:
						//链接
					case 51:
						//获取联系人信息成功
					case 62:
						//获得一段小视频
					case 10002:
						//撤回一条消息
					default:
						msg := Message{}
						msg.Content = fmt.Sprintf("未知消息：%s", m)
						msgIn <- msg
					}

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
			time.Sleep(time.Second * 4)

			continue
		}

		if time.Now().Sub(w.lastCheckTs).Seconds() <= 20 {
			time.Sleep(time.Second * time.Duration(time.Now().Sub(w.lastCheckTs).Seconds()))
		}

	}
}

func (w *Wechat) MsgDaemon(msgOut chan MessageOut, autoReply chan int) {
	msg := MessageOut{}
	var autoMode int
	for {
		select {
		case msg = <-msgOut:
			w.Log.Printf("the msg to send %+v", msg)
			w.SendMsg(msg.ToUserName, msg.Content, false)
		case autoMode = <-autoReply:
			w.Log.Println("the autoreply mode:", autoMode)
			if autoMode == 1 {
				w.AutoReply = true
			} else if autoMode == 0 {
				w.AutoReply = false
			}
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

func (w *Wechat) SendMsg(toUserName, message string, isFile bool) (err error) {
	resp := new(MsgResp)

	apiUrl := fmt.Sprintf("%s/webwxsendmsg?pass_ticket=%s", w.BaseUri, w.Request.PassTicket)
	clientMsgId := strconv.Itoa(w.GetUnixTime()) + "0" + strconv.Itoa(rand.Int())[3:6]
	params := make(map[string]interface{})
	params["BaseRequest"] = w.BaseRequest
	msg := make(map[string]interface{})
	msg["Type"] = 1
	msg["Content"] = message
	msg["FromUserName"] = w.User.UserName
	msg["LocalID"] = clientMsgId
	msg["ClientMsgId"] = clientMsgId
	msg["ToUserName"] = toUserName
	params["Msg"] = msg
	data, err := json.Marshal(params)
	if err != nil {
		w.Log.Printf("json.Marshal(%v):%v\n", params, err)
	}
	if err := w.Send(apiUrl, bytes.NewReader(data), resp); err != nil {
		w.Log.Print("w.Send(%s,%v,%v):%v", apiUrl, string(data), err)
	}

	return
}

func (w *Wechat) GetGroupName(id string) (name string) {
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

func (w *Wechat) SendTest(apiURI string, body io.Reader, call Caller) (err error) {
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

	respBody, err := ioutil.ReadAll(reader)
	w.Log.Printf("the respBody:%s", string(respBody))

	if err = json.NewDecoder(reader).Decode(call); err != nil {
		w.Log.Printf("the error:%+v", err)
		return
	}
	if !call.IsSuccess() {
		return call.Error()
	}
	return
}

func (w *Wechat) GetTuringReply(msg string) (retMsg string, err error) {
	params := url.Values{}
	params.Add("key", TUringUserId)
	params.Add("info", msg)
	params.Add("userid", TUringUserId)
	data, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", TuringUrl, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	dt, _ := ioutil.ReadAll(resp.Body)
	return string(dt), nil
}

func (w *Wechat) SetCookies() {
	//w.Client.Jar.SetCookies(u, cookies)

}

func (w *Wechat) GetUnixTime() int {
	return int(time.Now().Unix())
}
