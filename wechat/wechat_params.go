package wechat

type GetUUIDParams struct {
	AppId    string  `json:"appid"`
	Fun      string  `json:"fun"`
	Lang     string  `json:"lang"`
	UnixTime float64 `json:"_"`
}

func NewGetUUIDParams(appid, fun, lang string, times float64) *GetUUIDParams {
	return &GetUUIDParams{
		AppId:    appid,
		Fun:      fun,
		Lang:     lang,
		UnixTime: times,
	}
}
