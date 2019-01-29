package model

type Group struct {
	Id         int    `json:"id" xorm:"bigint pk autoincr"`
	Name       string `json:"name,omitempty" xorm:"varchar(100) notnull"`
	Describe   string `json:"describe,omitempty" xorm:"TEXT`
	CreateTime int    `json:"create_time,omitempty"`
	UpdateTime int    `json:"update_time,omitempty"`
	ImagePath  string `json:"image_path" xorm:"TEXT`

	// Future...
	Aa string `json:"aa,omitempty"`
	Ab string `json:"ab,omitempty"`
	Ac string `json:"ac,omitempty"`
	Ad string `json:"ad,omitempty"`
}