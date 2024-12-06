package model

import (
	"time"
)

type Proxy struct {
	ID            string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Remark        string    `json:"remark"`
	Prefix        string    `gorm:"uniqueIndex;not null" json:"prefix"`
	Upstream      string    `json:"upstream"`
	RewritePrefix string    `json:"rewritePrefix"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Respond struct {
	Data []Proxy `json:"data"`
}

func (p *Proxy) TableName() string {
	return "proxies"
}
