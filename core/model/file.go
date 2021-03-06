package model

import (
	"errors"
	"github.com/hunterhug/fafacms/core/config"
	"github.com/hunterhug/fafacms/core/util"
	"time"
)

type File struct {
	Id             int    `json:"id" xorm:"bigint pk autoincr"`
	Type           string `json:"type" xorm:"index"` // 这个字段可以作为索引，因为类型不长
	Tag            string `json:"tag" xorm:"index"`  // 这个用户打标签就不能太长了
	UserId         int    `json:"user_id" xorm:"bigint index"`
	UserName       string `json:"user_name" xorm:"index"`
	FileName       string `json:"file_name"`
	ReallyFileName string `json:"really_file_name"`
	HashCode       string `json:"hash_code" xorm:"unique"`
	Url            string `json:"url" xorm:"varchar(700)"` // index too long  布隆过滤器/字符串型的字段最好不要直接建索引
	UrlHashCode    string `json:"url_hash_code" xorm:"unique"`
	Describe       string `json:"describe" xorm:"TEXT"`
	CreateTime     int64  `json:"create_time"`
	UpdateTime     int64  `json:"update_time,omitempty"`
	Status         int    `json:"status" xorm:"not null comment('0 normal，1 hide but can use') TINYINT(1)"` // 逻辑隐藏为1，HashCode仍有效
	StoreType      int    `json:"store_type" xorm:"not null comment('0 local，1 oss') TINYINT(1)"`
	IsPicture      int    `json:"is_picture"`
	Size           int64  `json:"size"`
}

var FileSortName = []string{"=id", "-create_time", "-update_time", "=user_id", "=type", "=tag", "=store_type", "=status", "=size"}

// 判断文件是否存在，被隐藏的文件也可以找到
// 必须是图片，毕竟我们的系统目前只会上传图片
func (f *File) Exist() (bool, error) {
	if f.Id == 0 && f.Url == "" {
		return false, errors.New("where is empty")
	}
	s := config.FafaRdb.Client.Table(f)
	s.Where("1=1")

	if f.Id != 0 {
		s.And("id=?", f.Id)
	}
	if f.Url != "" {
		h, err := util.Sha256([]byte(f.Url))
		if err != nil {
			return false, err
		}
		s.And("url_hash_code=?", h)
	}

	c, err := s.Where("is_picture=?", 1).Count()

	if c >= 1 {
		return true, nil
	}

	return false, err
}

// 获取文件信息
func (f *File) Get() (bool, error) {
	if f.Id == 0 && f.Url == "" && f.UrlHashCode == "" && f.HashCode == "" {
		return false, errors.New("where is empty")
	}

	if f.Url != "" {
		h, err := util.Sha256([]byte(f.Url))
		if err != nil {
			return false, err
		}
		f.UrlHashCode = h
		f.Url = ""
	}

	return config.FafaRdb.Client.Get(f)
}

// 可以隐藏文件的更新操作
func (f *File) Update(hide bool) (bool, error) {
	if f.Id == 0 {
		return false, errors.New("where is empty")
	}

	s := config.FafaRdb.Client.NewSession()
	defer s.Close()

	s.Where("id=?", f.Id)

	// 如果要隐藏，那么这样：
	if hide {
		f.Status = 1
		s.Cols("status")
	}

	if f.UserId != 0 {
		s.And("user_id=?", f.UserId)
	}

	if f.Describe != "" {
		s.Cols("describe")
	}

	if f.Tag != "" {
		s.Cols("tag")
	}

	f.UpdateTime = time.Now().Unix()
	s.Cols("update_time")
	_, err := s.Update(f)
	if err != nil {
		return false, err
	}

	return true, nil
}

// 更新状态
func (f *File) UpdateStatus() (bool, error) {
	if f.Id == 0 {
		return false, errors.New("where is empty")
	}

	s := config.FafaRdb.Client.NewSession()
	defer s.Close()

	s.Where("id=?", f.Id)

	s.Cols("status")

	if f.UserId != 0 {
		s.And("user_id=?", f.UserId)
	}

	f.UpdateTime = time.Now().Unix()
	s.Cols("update_time")

	_, err := s.Update(f)
	if err != nil {
		return false, err
	}

	return true, nil
}
