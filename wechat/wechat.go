package wechat

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/skratchdot/open-golang/open"
)

const (
	UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.109 Safari/537.36"
)

var (
	SaveSubFolders = map[string]string{"webwxgeticon": "icons",
		"webwxgetheadimg": "headimgs",
		"webwxgetmsgimg":  "msgimgs",
		"webwxgetvideo":   "videos",
		"webwxgetvoice":   "voices",
		"_showQRCodeImg":  "qrcodes",
	}
	AppID        = "wx782c26e4c19acffb"
	Lang         = "zh_CN"
	LastCheckTs  = time.Now()
	LoginUrl     = "https://login.weixin.qq.com/jslogin"
	QrUrl        = "https://login.weixin.qq.com/qrcode/"
	TuringUrl    = "" //"http://www.tuling123.com/openapi/api"
	APIKEY       = "" //"391ad66ebad2477b908dce8e79f101e7"
	TUringUserId = "" //"abc123"
)

type Wechat struct {
	User            User
	Root            string
	Debug           bool
	Uuid            string
	BaseUri         string
	RedirectedUri   string
	Uin             string
	Sid             string
	Skey            string
	PassTicket      string
	DeviceId        string
	BaseRequest     map[string]string
	LowSyncKey      string
	SyncKeyStr      string
	SyncHost        string
	SyncKey         SyncKey
	Users           []string
	InitContactList []User   //谈话的人
	MemberList      []Member //
	ContactList     []Member //好友
	GroupList       []string //群
	GroupMemberList []Member //群友
	PublicUserList  []Member //公众号
	SpecialUserList []Member //特殊账号

	AutoReplyMode bool //default false
	AutoOpen      bool
	Interactive   bool
	TotalMember   int
	TimeOut       int // 同步时间间隔   default:20
	MediaCount    int // -1
	SaveFolder    string
	QrImagePath   string
	Client        *http.Client
	Request       *BaseRequest
	Log           *log.Logger
	MemberMap     map[string]Member
	ChatSet       []string

	AutoReply    bool     //是否自动回复
	ReplyMsgs    []string // 回复的消息列表
	AutoReplySrc bool     //默认false,自动回复，列表。true调用AI机器人。
	lastCheckTs  time.Time
	SetCookie    []string
}

func NewWechat(logger *log.Logger) *Wechat {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil
	}

	root, err := os.Getwd()
	transport := *(http.DefaultTransport.(*http.Transport))
	transport.ResponseHeaderTimeout = 1 * time.Minute
	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	return &Wechat{
		Debug:         true,
		DeviceId:      "e123456789002237",
		AutoReplyMode: false,
		Interactive:   false,
		AutoOpen:      false,
		MediaCount:    -1,
		Client: &http.Client{
			Transport: &transport,
			Jar:       jar,
			Timeout:   1 * time.Minute,
		},
		Request:     new(BaseRequest),
		Root:        root,
		SaveFolder:  path.Join(root, "saved"),
		QrImagePath: filepath.Join(root, "qr.jpg"),
		Log:         logger,
		MemberMap:   make(map[string]Member),
		SetCookie:   []string{},
	}

}

func (w *Wechat) WaitForLogin() (err error) {

	err = w.GetUUID()
	if err != nil {
		err = fmt.Errorf("get the uuid failed with error:%v", err)
	}
	err = w.GetQR()
	if err != nil {
		err = fmt.Errorf("创建二维码失败:%s", err.Error())
	}
	defer os.Remove(w.QrImagePath)
	w.Log.Println("扫描二维码登陆....")
	code, tip := "", 1
	for code != "200" {
		w.RedirectedUri, code, tip, err = w.waitToLogin(w.Uuid, tip)
		if err != nil {
			err = fmt.Errorf("二维码登陆失败：%s", err.Error())
			return
		}
	}
	return
}

func (w *Wechat) waitToLogin(uuid string, tip int) (redirectUri, code string, rt int, err error) {
	loginUri := fmt.Sprintf("https://login.weixin.qq.com/cgi-bin/mmwebwx-bin/login?tip=%d&uuid=%s&_=%s", tip, uuid, time.Now().Unix())
	rt = tip
	resp, err := w.Client.Get(loginUri)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	re := regexp.MustCompile(`window.code=(\d+);`)
	pm := re.FindStringSubmatch(string(data))

	if len(pm) != 0 {
		code = pm[1]

	} else {
		err = errors.New("can't find the code")
		return
	}
	rt = 0
	switch code {
	case "201":
		w.Log.Println("扫描成功，请在手机上点击确认登陆")
	case "200":
		reRedirect := regexp.MustCompile(`window.redirect_uri="(\S+?)"`)
		pmSub := reRedirect.FindStringSubmatch(string(data))

		if len(pmSub) != 0 {
			redirectUri = pmSub[1]
		} else {
			err = errors.New("regex error in window.redirect_uri")
			return
		}
		redirectUri += "&fun=new"
	case "408":
	case "0":
		err = errors.New("超时了，请重启程序")
	default:
		err = errors.New("其它错误，请重启")

	}
	return
}

func (w *Wechat) GetQR() (err error) {
	if w.Uuid == "" {
		err = errors.New("no this uuid")
		return
	}
	params := url.Values{}
	params.Set("t", "webwx")
	params.Set("_", strconv.FormatInt(time.Now().Unix(), 10))
	req, err := http.NewRequest("POST", QrUrl+w.Uuid, strings.NewReader(params.Encode()))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cache-Control", "no-cache")
	resp, err := w.Client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if err = createFile(w.QrImagePath, data, false); err != nil {
		return
	}

	return open.Start(w.QrImagePath)

}

func (w *Wechat) SetSynKey() {

}

func (w *Wechat) AutoReplyMsg() string {
	if w.AutoReplySrc {
		return "" //not enabled
	} else {
		if len(w.ReplyMsgs) == 0 {
			return "未设置"
		}
		return w.ReplyMsgs[0]
	}

}

func (w *Wechat) GetUUID() (err error) {
	params := url.Values{}
	params.Set("appid", AppID)
	params.Set("fun", "new")
	params.Set("lang", "zh_CN")
	params.Set("_", strconv.FormatInt(time.Now().Unix(), 10))
	datas := w.Post(LoginUrl, params, false)

	re := regexp.MustCompile(`window.QRLogin.code = (\d+); window.QRLogin.uuid = "(\S+?)"`)
	pm := re.FindStringSubmatch(string(datas))

	fmt.Printf("%v", pm)

	if len(pm) > 0 {
		code := pm[1]
		if code != "200" {
			err = errors.New("the status error")
		} else {
			w.Uuid = pm[2]
		}
		return
	} else {
		err = errors.New("get uuid failed")
		return
	}

}

func (w *Wechat) Login() (err error) {
	w.Log.Printf("the redirectedUri:%v", w.RedirectedUri)

	resp, err := w.Client.Get(w.RedirectedUri)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	w.SetCookie = resp.Header["Set-Cookie"]
	reader := resp.Body.(io.Reader)
	if err = xml.NewDecoder(reader).Decode(w.Request); err != nil {
		return
	}

	w.Request.DeviceID = w.DeviceId

	data, err := json.Marshal(Request{
		BaseRequest: w.Request,
	})
	if err != nil {
		return
	}

	name := "webwxinit"
	newResp := new(InitResp)

	index := strings.LastIndex(w.RedirectedUri, "/")
	if index == -1 {
		index = len(w.RedirectedUri)
	}
	w.BaseUri = w.RedirectedUri[:index]

	apiUri := fmt.Sprintf("%s/%s?pass_ticket=%s&skey=%s&r=%d", w.BaseUri, name, w.Request.PassTicket, w.Request.Skey, int(time.Now().Unix()))
	if err = w.Send(apiUri, bytes.NewReader(data), newResp); err != nil {
		return
	}
	w.Log.Printf("the newResp:%#v", newResp)
	for _, contact := range newResp.ContactList {
		w.InitContactList = append(w.InitContactList, contact)
	}

	w.ChatSet = strings.Split(newResp.ChatSet, ",")
	w.User = newResp.User
	w.SyncKey = newResp.SyncKey
	w.SyncKeyStr = ""
	for i, item := range w.SyncKey.List {

		if i == 0 {
			w.SyncKeyStr = strconv.Itoa(item.Key) + "_" + strconv.Itoa(item.Val)
			continue
		}

		w.SyncKeyStr += "|" + strconv.Itoa(item.Key) + "_" + strconv.Itoa(item.Val)

	}
	w.Log.Printf("the response:%+v\n", newResp)
	w.Log.Printf("the sync key is %s\n", w.SyncKeyStr)
	return
}
