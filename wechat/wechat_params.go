package wechat

import (
	"encoding/xml"
	"fmt"
	"os"
)

type GetUUIDParams struct {
	AppId    string  `json:"appid"`
	Fun      string  `json:"fun"`
	Lang     string  `json:"lang"`
	UnixTime float64 `json:"_"`
}

type Response struct {
	BaseResponse *BaseResponse `json:"BaseResponse"`
}

type Request struct {
	BaseRequest *BaseRequest

	MemberCount int    `json:",omitempty"`
	MemberList  []User `json:",omitempty"`
	Topic       string `json:",omitempty"`

	ChatRoomName  string `json:",omitempty"`
	DelMemberList string `json:",omitempty"`
	AddMemberList string `json:",omitempty"`
}
type Caller interface {
	IsSuccess() bool
	Error() error
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

type BaseResponse struct {
	Ret    int
	ErrMsg string
}

type InitResp struct {
	Response
	User                User    `json:"User"`
	Count               int     `json:"Count"`
	ContactList         []User  `json:"ContactList"`
	SyncKey             SyncKey `json:"SyncKey"`
	ChatSet             string  `json:"ChatSet"`
	SKey                string  `json:"SKey"`
	ClientVersion       int     `json:"ClientVersion"`
	SystemTime          int     `json:"SystemTime"`
	GrayScale           int     `json:"GrayScale"`
	InviteStartCount    int     `json:"InviteStartCount"`
	MPSubscribeMsgCount int     `json:"MPSubscribeMsgCount"`
	MPSubscribeMsgList  string  `json:"MPSubscribeMsgList"`
	ClickReportInterval int     `json:"ClickReportInterval"`
}

type SyncKey struct {
	Count int      `json:"Count"`
	List  []KeyVal `json:"List"`
}

type KeyVal struct {
	Key int `json:"Key"`
	Val int `json:"Val"`
}

func (r *Response) IsSuccess() bool {
	return r.BaseResponse.Ret == 0
}

func (r *Response) Error() error {
	return fmt.Errorf("message:[%s]", r.BaseResponse.ErrMsg)
}

type User struct {
	UserName          string `json:"UserName"`
	Uin               string `json:"Uin"`
	NickName          string `json:"NickName"`
	HeadImgUrl        string `json:"HeadImgUrl" xml:""`
	RemarkName        string `json:"RemarkName" xml:""`
	PYInitial         string `json:"PYInitial" xml:""`
	PYQuanPin         string `json:"PYQuanPin" xml:""`
	RemarkPYInitial   string `json:"RemarkPYInitial" xml:""`
	RemarkPYQuanPin   string `json:"RemarkPYQuanPin" xml:""`
	HideInputBarFlag  int    `json:"HideInputBarFlag" xml:""`
	StarFriend        int    `json:"StarFriend" xml:""`
	Sex               int    `json:"Sex" xml:""`
	Signature         string `json:"Signature" xml:""`
	AppAccountFlag    int    `json:"AppAccountFlag" xml:""`
	VerifyFlag        int    `json:"VerifyFlag" xml:""`
	ContactFlag       int    `json:"ContactFlag" xml:""`
	WebWxPluginSwitch int    `json:"WebWxPluginSwitch" xml:""`
	HeadImgFlag       int    `json:"HeadImgFlag" xml:""`
	SnsFlag           int    `json:"SnsFlag" xml:""`
}

type MemberResp struct {
	Response
	MemberCount  int
	ChatRoomName string
	MemberList   []*Member
}

type Member struct {
	UserName     string
	NickName     string
	RemarkName   string
	VerifyFlag   int
	MemberStatus int
}

func NewGetUUIDParams(appid, fun, lang string, times float64) *GetUUIDParams {
	return &GetUUIDParams{
		AppId:    appid,
		Fun:      fun,
		Lang:     lang,
		UnixTime: times,
	}
}
func createFile(name string, data []byte, isAppend bool) (err error) {
	oflag := os.O_CREATE | os.O_WRONLY
	if isAppend {
		oflag |= os.O_APPEND
	} else {
		oflag |= os.O_TRUNC
	}

	file, err := os.OpenFile(name, oflag, 0600)
	if err != nil {
		return
	}
	defer file.Close()

	_, err = file.Write(data)
	return
}
