package wechat

import (
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"time"
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
	AppId       = "wx782c26e4c19acffb"
	Lang        = "zh_CN"
	LastCheckTs = time.Now()
	LoginUrl    = "https://login.weixin.qq.com/jslogin"
)

type Wechat struct {
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
	SyncKeys        []string
	Users           []string
	MemberList      []string //
	ContactList     []string //好友
	GroupList       []string //群
	GroupMemberList []string //群友
	PublicUserList  []string //公众号
	SpecialUserList []string //特殊账号
	AutoReplyMode   bool     //default false
	AutoOpen        bool
	SyncHost        string
	Interactive     bool
	TimeOut         int // 同步时间间隔   default:20
	MediaCount      int // -1
	SaveFolder      string
	Client          *http.Client
	Request         *BaseRequest
}

func NewWechat() *Wechat {
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
		DeviceId:      "e123456789001234",
		AutoReplyMode: false,
		Interactive:   false,
		AutoOpen:      false,
		MediaCount:    -1,
		Client: &http.Client{
			Transport: &transport,
			Jar:       jar,
			Timeout:   1 * time.Minute,
		},
		Request:    new(BaseRequest),
		Root:       root,
		SaveFolder: path.Join(root, "saved"),
	}

}

type BaseRequest struct {
	XMLName xml.Name `xml:"error",json:"-"`

	Ret        int    `xml:"ret",json:"-"`
	Message    string `xml:"message",json:"-"`
	Skey       string `xml:"skey"`
	Wxsid      string `xml:"wxsid",json:"Sid"`
	Wxuin      int    `xml:"wxuin",json:"Uin"`
	PassTicket string `xml:"pass_ticket",json:"-"`

	DeviceID string `xml:"-"`
}

func (w *Wechat) GetUUID() (err error) {
	params := url.Values{}
	params.Set("appid", AppId)
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

func (w *Wechat) Post(url string, data url.Values, jsonFmt bool) (result string) {
	//req.Header.Set("User-agent", UserAgent)

	resp, err := w.Client.PostForm(url, data)

	//req.Header.Set("ContentType", "application/json; charset=UTF-8")

	fmt.Printf("resp%#v", resp)
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

func (w *Wechat) SetCookies() {
	//w.Client.Jar.SetCookies(u, cookies)

}

func main() {

}
