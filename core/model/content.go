package model

import (
	"errors"
	"github.com/hunterhug/fafacms/core/config"
	"time"
)

// 内容表
type Content struct {
	Id           int    `json:"id" xorm:"bigint pk autoincr"`
	Seo          string `json:"seo" xorm:"index"`
	Title        string `json:"title" xorm:"varchar(200) notnull"`
	PreTitle     string `json:"pre_title" xorm:"varchar(200) notnull"`
	UserId       int    `json:"user_id" xorm:"bigint index"` // 内容所属用户
	UserName     string `json:"user_name" xorm:"index"`
	NodeId       int    `json:"node_id" xorm:"bigint index"`                                                          // 节点ID
	NodeSeo      string `json:"node_seo" xorm:"index"`                                                                // 节点ID SEO
	Status       int    `json:"status" xorm:"not null comment('0 normal, 1 hide，2 ban, 3 rubbish') TINYINT(1) index"` // 0-1-2-3为正常
	Top          int    `json:"top" xorm:"not null comment('0 normal, 1 top') TINYINT(1) index"`                      // 置顶
	Describe     string `json:"describe" xorm:"TEXT"`
	PreDescribe  string `json:"pre_describe" xorm:"TEXT"`                                                           // 预览内容，临时保存，当修改后调用发布接口，会刷新到Describe，每次这个字段刷新都会记录进历史表
	PreFlush     int    `json:"pre_flush" xorm:"not null comment('1 flush') TINYINT(1)"`                            // 是否预览内容已经被刷新
	CloseComment int    `json:"close_comment" xorm:"not null comment('0 close, 1 open, 2 direct open') TINYINT(1)"` // 关闭评论开关，默认关闭
	Version      int    `json:"version"`                                                                            // 0表示什么都没发布  发布了多少次版本
	CreateTime   int64  `json:"create_time"`
	UpdateTime   int64  `json:"update_time,omitempty"`
	PublishTime  int64  `json:"publish_time,omitempty"`
	ImagePath    string `json:"image_path" xorm:"varchar(700)"`
	Views        int    `json:"views"` // 被点击多少次，弱化
	Password     string `json:"password,omitempty"`
	SortNum      int64  `json:"sort_num"`
}

var ContentSortName = []string{"=id", "-user_id", "-top", `-sort_num`, "-create_time", "-update_time", "-views", "=version", "+status", "=seo"}

// 内容历史表
type ContentHistory struct {
	Id         int    `json:"id" xorm:"bigint pk autoincr"`
	ContentId  int    `json:"content_id" xorm:"bigint index"` // 内容ID
	Title      string `json:"title" xorm:"varchar(200) notnull"`
	UserId     int    `json:"user_id" xorm:"bigint index"` // 内容所属的用户ID
	NodeId     int    `json:"node_id" xorm:"bigint index"` // 内容所属的节点
	Describe   string `json:"describe" xorm:"TEXT"`
	Types      int    `json:"types" xorm:"not null comment('0 auto save, 1 publish, 3 cancel') TINYINT(1)"` // 0表示是自动刷新的，1表示发布，2表示是从历史版本恢复的
	CreateTime int64  `json:"create_time"`
}

var ContentHistorySortName = []string{"=id", "-user_id", "-create_time", "-content_id"}

// 统计节点下的内容数量
func (c *Content) CountNumUnderNode() (int64, error) {
	if c.UserId == 0 || c.NodeId == 0 {
		return 0, errors.New("where is empty")
	}

	allNum, err := config.FafaRdb.Client.Table(c).Where("user_id=?", c.UserId).And("node_id=?", c.NodeId).Count()
	if err != nil {
		return 0, err
	}
	return allNum, nil
}

func (c *Content) CheckSeoValid() (bool, error) {
	// 用户ID和SEO必须同时存在
	if c.UserId == 0 || c.Seo == "" {
		return false, errors.New("where is empty")
	}

	// 常规统计
	num, err := config.FafaRdb.Client.Table(c).Where("user_id=?", c.UserId).And("seo=?", c.Seo).Count()

	// 如果大于1表示存在
	if num >= 1 {
		return true, nil
	}
	return false, err
}

// 硬核插入
func (c *Content) Insert() (int64, error) {
	c.CreateTime = time.Now().Unix()
	return config.FafaRdb.InsertOne(c)
}

// 一般的获取，放松，需要内容ID
func (c *Content) Get() (bool, error) {
	if c.Id == 0 {
		return false, errors.New("where is empty")
	}

	return config.FafaRdb.Client.Get(c)
}

// 硬一点
func (c *Content) GetByRaw() (bool, error) {
	return config.FafaRdb.Client.Get(c)
}

// 更新内容，会写历史表
func (c *Content) UpdateDescribeAndHistory() error {
	if c.UserId == 0 || c.Id == 0 {
		return errors.New("where is empty")
	}

	session := config.FafaRdb.Client.NewSession()
	defer session.Close()

	history := new(ContentHistory)
	history.NodeId = c.NodeId
	history.CreateTime = time.Now().Unix()
	// 之前的内容要刷进历史表
	history.Title = c.PreTitle
	history.Describe = c.PreDescribe
	history.ContentId = c.Id

	// 一般类型
	history.Types = 0
	_, err := session.InsertOne(history)
	if err != nil {
		session.Rollback()
		return err
	}

	// 版本要+1
	c.Version = c.Version + 1
	c.UpdateTime = time.Now().Unix()

	// 把目前的内容写进去
	c.PreDescribe = c.Describe
	c.PreTitle = c.Title
	c.PreFlush = 0
	_, err = session.Cols("update_time", "version", "pre_title", "pre_describe", "pre_flush").Where("id=?", c.Id).And("user_id=?", c.UserId).Update(c)
	if err != nil {
		session.Rollback()
		return err
	}

	err = session.Commit()
	if err != nil {
		session.Rollback()
	}

	return err
}

// 更新SEO，不需要更新时间，在内容变化才需要
func (c *Content) UpdateSeo() (int64, error) {
	if c.UserId == 0 || c.Id == 0 {
		return 0, errors.New("where is empty")
	}
	return config.FafaRdb.Client.Cols("seo").Where("id=?", c.Id).And("user_id=?", c.UserId).Update(c)
}

// 更新图片
func (c *Content) UpdateImage() (int64, error) {
	if c.UserId == 0 || c.Id == 0 {
		return 0, errors.New("where is empty")
	}
	return config.FafaRdb.Client.Cols("image_path").Where("id=?", c.Id).And("user_id=?", c.UserId).Update(c)
}

// 更新状态
func (c *Content) UpdateStatus() (int64, error) {
	if c.UserId == 0 || c.Id == 0 {
		return 0, errors.New("where is empty")
	}
	return config.FafaRdb.Client.Cols("status").Where("id=?", c.Id).And("user_id=?", c.UserId).Update(c)
}

// 更新Top
func (c *Content) UpdateTop() (int64, error) {
	if c.UserId == 0 || c.Id == 0 {
		return 0, errors.New("where is empty")
	}
	return config.FafaRdb.Client.Cols("top").Where("id=?", c.Id).And("user_id=?", c.UserId).Update(c)
}

// 更新密码
func (c *Content) UpdatePassword() (int64, error) {
	if c.UserId == 0 || c.Id == 0 {
		return 0, errors.New("where is empty")
	}
	return config.FafaRdb.Client.Cols("password").Where("id=?", c.Id).And("user_id=?", c.UserId).Update(c)
}

// 更新内容的节点
func (n *Content) UpdateNode(beforeNodeId int) error {
	if n.UserId == 0 || n.Id == 0 {
		return errors.New("where is empty")
	}

	session := config.FafaRdb.Client.NewSession()
	defer session.Close()
	err := session.Begin()
	if err != nil {
		return err
	}

	// 先把这个内容顶出去
	_, err = session.Exec("update fafacms_content SET sort_num=sort_num-1 where sort_num > ? and user_id = ? and node_id = ?", n.SortNum, n.UserId, beforeNodeId)
	if err != nil {
		session.Rollback()
		return err
	}

	// 统计目前节点的数量
	c, err := session.Table(n).Where("user_id=?", n.UserId).And("node_id", n.NodeId).Count()
	if err != nil {
		session.Rollback()
		return err
	}

	// 好，这个内容顶上
	n.SortNum = c

	// 每次更改节点，他都会成为这一层最靓丽排得最前面的仔
	_, err = session.Exec("update fafacms_content SET sort_num=?, node_id=?, node_seo=? where id = ? and user_id = ?", n.SortNum, n.NodeId, n.NodeSeo, n.Id, n.UserId)
	if err != nil {
		session.Rollback()
		return err
	}

	err = session.Commit()
	if err != nil {
		session.Rollback()
		return err
	}

	return nil
}

// 更新前都会调用， 不需要处理错误，不考虑互斥
func (c *Content) UpdateView() {
	config.FafaRdb.Client.Where("id=?", c.Id).Incr("views").Cols("views").Update(c)
}

// 发布更新内容
func (c *Content) PublishDescribe() error {
	if c.UserId == 0 || c.Id == 0 {
		return errors.New("where is empty")
	}

	session := config.FafaRdb.Client.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return err
	}

	history := new(ContentHistory)
	history.NodeId = c.NodeId
	history.CreateTime = time.Now().Unix()
	// 之前的内容要刷进历史表
	history.Title = c.PreTitle
	history.Describe = c.PreDescribe
	history.ContentId = c.Id

	// 发布类型
	history.Types = 1
	_, err := session.InsertOne(history)
	if err != nil {
		session.Rollback()
		return err
	}

	// 版本要+1
	c.Version = c.Version + 1
	c.UpdateTime = time.Now().Unix()
	c.PreFlush = 1
	c.Title = c.PreTitle
	c.Describe = c.PreDescribe
	c.PublishTime = c.UpdateTime
	_, err = session.Cols("title", "describe", "pre_flush", "update_time", "publish_time", "version").Where("id=?", c.Id).And("user_id=?", c.UserId).Update(c)
	if err != nil {
		session.Rollback()
		return err
	}

	if err := session.Commit(); err != nil {
		session.Rollback()
		return err
	}
	return nil
}

func (c *Content) ResetDescribe() error {
	if c.UserId == 0 || c.Id == 0 {
		return errors.New("where is empty")
	}

	session := config.FafaRdb.Client.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return err
	}

	history := new(ContentHistory)
	history.NodeId = c.NodeId
	history.CreateTime = time.Now().Unix()
	// 之前的内容要刷进历史表
	history.Title = c.PreTitle
	history.Describe = c.PreDescribe
	history.ContentId = c.Id

	// 恢复类型
	history.Types = 2
	_, err := session.InsertOne(history)
	if err != nil {
		session.Rollback()
		return err
	}

	// 版本要+1
	c.Version = c.Version + 1
	c.UpdateTime = time.Now().Unix()
	c.PreFlush = 0
	c.PreTitle = c.Title
	c.PreDescribe = c.Describe
	_, err = session.Cols("pre_title", "pre_describe", "pre_flush", "update_time", "version").Where("id=?", c.Id).And("user_id=?", c.UserId).Update(c)
	if err != nil {
		session.Rollback()
		return err
	}

	if err := session.Commit(); err != nil {
		session.Rollback()
		return err
	}
	return nil
}

// 级联删除
func (c *Content) Delete() error {
	if c.UserId == 0 || c.Id == 0 {
		return errors.New("where is empty")
	}
	//c.UpdateTime = time.Now().Unix()
	//c.Status = 4
	//return config.FafaRdb.Client.Cols("status", "update_time").Where("status>=?", 2).Where("id=?", c.Id).And("user_id=?", c.UserId).Update(c)
	session := config.FafaRdb.Client.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return err
	}

	if _, err := session.Where("id=?", c.Id).And("user_id=?", c.UserId).Delete(new(Content)); err != nil {
		session.Rollback()
		return err
	}

	if _, err := session.Where("content_id=?", c.Id).And("user_id=?", c.UserId).Delete(new(ContentHistory)); err != nil {
		session.Rollback()
		return err
	}

	if err := session.Commit(); err != nil {
		return err
	}

	return nil
}

func (c *ContentHistory) GetRaw() (bool, error) {
	return config.FafaRdb.Client.Get(c)
}
