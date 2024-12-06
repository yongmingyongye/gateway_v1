package model

type Respond struct {
	Success bool    `json:"success"`
	Status  string  `json:"status"`
	Data    []Proxy `json:"data"`
}

type Proxy struct {
	ID            string `json:"id"`
	Remark        string `json:"remark"`        // 描述
	Prefix        string `json:"prefix"`        // 转发的前缀判断
	Upstream      string `json:"upstream"`      // 后端nginx或ip地址
	RewritePrefix string `json:"rewritePrefix"` // 重写
}
