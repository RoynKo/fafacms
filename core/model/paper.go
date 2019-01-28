package model

type Content struct {
	Id         int    `json:"id" xorm:"bigint pk autoincr"`
	Title      string `json:"name,omitempty" xorm:"varchar(200) notnull"`
	UserId     string `json:"user_id"`
	Describe   string `json:"describe" xorm:"TEXT`
	CreateTime int    `json:"create_time,omitempty"`
	UpdateTime int    `json:"update_time,omitempty"`
	DeleteTime int    `json:"delete_time,omitempty"`
	Status     int    `json:"status,omitempty xorm:"not null comment('1 normal，2 deleted') TINYINT(1)"`
	ImagePath  string `json:"image_path" xorm:"TEXT`
	Views      int    `json:"views"`
	Node       int    `json:"node`
	Password   string `json:"password,omitempty"`
	Good       int    `json:"good"`
	Bad        int    `json:"bad"`
	IsPhoto    int    `json:"status,omitempty xorm:"not null comment('1 paper，2 album') TINYINT(1)"`
}

type ContentNode struct {
	Id         int    `json:"id" xorm:"bigint pk autoincr"`
	Name       string `json:"name,omitempty" xorm:"varchar(100) notnull"`
	Describe   string `json:"describe" xorm:"TEXT`
	CreateTime int    `json:"create_time,omitempty"`
	UpdateTime int    `json:"update_time,omitempty"`
	ImagePath  string `json:"image_path" xorm:"TEXT`
}
