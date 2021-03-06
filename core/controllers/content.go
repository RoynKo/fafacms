package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/hunterhug/fafacms/core/config"
	"github.com/hunterhug/fafacms/core/flog"
	"github.com/hunterhug/fafacms/core/model"
	"math"
)

// 创建内容
type CreateContentRequest struct {
	Seo          string `json:"seo" validate:"omitempty,alphanumunicode,gt=3,lt=30"` // 内容应该有个好听的标志
	Title        string `json:"title" validate:"required,lt=100"`                    // 必须有标题吧
	Status       int    `json:"status" validate:"oneof=0 1"`                         // 隐藏内容，1就是隐藏
	Top          int    `json:"top" validate:"oneof=0 1"`                            // 置顶，1就是置顶
	Describe     string `json:"describe" validate:"omitempty"`                       // 正文
	ImagePath    string `json:"image_path" validate:"omitempty,lt=100"`              // 内容背景图
	NodeId       int    `json:"node_id"`                                             // 内容所属节点，可以没有节点
	Password     string `json:"password"`                                            // 如果非空表示需要密码
	CloseComment int    `json:"close_comment" validate:"oneof=0 1 2"`                // 评论设置，0关闭评论，1打开评论需要审核，2打开评论不需要审核
}

func CreateContent(c *gin.Context) {
	resp := new(Resp)
	req := new(CreateContentRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("CreateContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		flog.Log.Errorf("CreateContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	content := new(model.Content)
	content.UserId = uu.Id
	if req.Seo != "" {
		content.Seo = req.Seo
		exist, err := content.CheckSeoValid()
		if err != nil {
			flog.Log.Errorf("CreateContent err: %s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
		if exist {
			flog.Log.Errorf("CreateContent err: %s", "seo repeat")
			resp.Error = Error(ContentSeoAlreadyBeUsed, "")
			return
		}
	}

	if req.NodeId == 0 {
		flog.Log.Errorf("CreateContent err: %s", "node_id can not empty")
		resp.Error = Error(ParasError, "node_id can not empty")
		return
	}

	content.NodeId = req.NodeId
	contentNode := new(model.ContentNode)
	contentNode.Id = req.NodeId
	contentNode.UserId = uu.Id
	exist, err := contentNode.Get()
	if err != nil {
		flog.Log.Errorf("CreateContent err: %s", err.Error())
		resp.Error = Error(DBError, "")
		return
	}

	if !exist {
		flog.Log.Errorf("CreateContent err: %s", "node not found")
		resp.Error = Error(ContentNodeNotFound, "")
		return
	}

	content.NodeSeo = contentNode.Seo

	if req.ImagePath != "" {
		content.ImagePath = req.ImagePath
		p := new(model.File)
		p.Url = req.ImagePath
		ok, err := p.Exist()
		if err != nil {
			flog.Log.Errorf("CreateContent err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}

		if !ok {
			flog.Log.Errorf("CreateContent err: image not exist")
			resp.Error = Error(FileCanNotBeFound, "")
			return
		}
	}

	content.Status = req.Status
	content.PreDescribe = req.Describe
	content.PreTitle = req.Title
	content.Password = req.Password
	content.CloseComment = req.CloseComment
	content.Top = req.Top
	content.UserName = uu.Name
	content.SortNum, _ = content.CountNumUnderNode()
	_, err = content.Insert()
	if err != nil {
		flog.Log.Errorf("CreateContent err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	resp.Data = content
	resp.Flag = true
}

// 更新内容SEO
type UpdateSeoOfContentRequest struct {
	Id  int    `json:"id" validate:"required"`
	Seo string `json:"seo" validate:"required,alphanumunicode,gt=3,lt=30"`
}

func UpdateSeoOfContent(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdateSeoOfContentRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("UpdateSeoOfContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		flog.Log.Errorf("UpdateSeoOfContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	contentBefore := new(model.Content)
	contentBefore.Id = req.Id
	contentBefore.UserId = uu.Id
	exist, err := contentBefore.Get()
	if err != nil {
		flog.Log.Errorf("UpdateSeoOfContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		flog.Log.Errorf("UpdateSeoOfContent err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	content := new(model.Content)
	content.Id = req.Id
	content.UserId = uu.Id
	if req.Seo != contentBefore.Seo {
		content.Seo = req.Seo
		exist, err := content.CheckSeoValid()
		if err != nil {
			flog.Log.Errorf("UpdateSeoOfContent err: %s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
		if exist {
			flog.Log.Errorf("UpdateSeoOfContent err: %s", "seo repeat")
			resp.Error = Error(ContentSeoAlreadyBeUsed, "")
			return
		}

		_, err = content.UpdateSeo()
		if err != nil {
			flog.Log.Errorf("UpdateSeoOfContent err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}
	resp.Flag = true
}

// 更新内容图片
type UpdateImageOfContentRequest struct {
	Id        int    `json:"id" validate:"required"`
	ImagePath string `json:"image_path" validate:"required,lt=100"`
}

func UpdateImageOfContent(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdateImageOfContentRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("UpdateImageOfContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		flog.Log.Errorf("UpdateImageOfContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	contentBefore := new(model.Content)
	contentBefore.Id = req.Id
	contentBefore.UserId = uu.Id
	exist, err := contentBefore.Get()
	if err != nil {
		flog.Log.Errorf("UpdateImageOfContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		flog.Log.Errorf("UpdateImageOfContent err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	content := new(model.Content)
	content.Id = req.Id
	content.UserId = uu.Id
	if req.ImagePath != contentBefore.ImagePath {
		p := new(model.File)
		p.Url = req.ImagePath
		ok, err := p.Exist()
		if err != nil {
			flog.Log.Errorf("UpdateImageOfContent err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}

		if !ok {
			flog.Log.Errorf("UpdateImageOfContent err: image not exist")
			resp.Error = Error(FileCanNotBeFound, "")
			return
		}

		content.ImagePath = req.ImagePath
		_, err = content.UpdateImage()
		if err != nil {
			flog.Log.Errorf("UpdateImageOfContent err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}
	resp.Flag = true
}

// 管理员更新内容状态
type UpdateStatusOfContentAdminRequest struct {
	Id     int `json:"id" validate:"required"`
	Status int `json:"status" validate:"oneof=0 1 2 3"`
}

func UpdateStatusOfContentAdmin(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdateStatusOfContentAdminRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("UpdateStatusOfContentAdmin err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	contentBefore := new(model.Content)
	contentBefore.Id = req.Id
	exist, err := contentBefore.GetByRaw()
	if err != nil {
		flog.Log.Errorf("UpdateStatusOfContentAdmin err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		flog.Log.Errorf("UpdateStatusOfContentAdmin err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	content := new(model.Content)
	content.Id = req.Id
	if req.Status != contentBefore.Status {
		content.Status = req.Status
		_, err = content.UpdateStatus()
		if err != nil {
			flog.Log.Errorf("UpdateStatusOfContentAdmin err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}
	resp.Flag = true
}

// 更新内容状态
type UpdateStatusOfContentRequest struct {
	Id     int `json:"id" validate:"required"`
	Status int `json:"status" validate:"oneof=0 1"`
}

func UpdateStatusOfContent(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdateStatusOfContentRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("UpdateStatusOfContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		flog.Log.Errorf("UpdateStatusOfContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	contentBefore := new(model.Content)
	contentBefore.Id = req.Id
	contentBefore.UserId = uu.Id
	exist, err := contentBefore.Get()
	if err != nil {
		flog.Log.Errorf("UpdateStatusOfContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		flog.Log.Errorf("UpdateStatusOfContent err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	if contentBefore.Status == 2 {
		flog.Log.Errorf("UpdateStatusOfContent err: %s", "content ban")
		resp.Error = Error(ContentBanPermit, "")
		return
	}

	if contentBefore.Status == 3 {
		flog.Log.Errorf("UpdateStatusOfContent err: %s", "content rubbish")
		resp.Error = Error(ContentInRubbish, "")
		return
	}

	content := new(model.Content)
	content.Id = req.Id
	content.UserId = uu.Id
	if req.Status != contentBefore.Status {
		content.Status = req.Status
		_, err = content.UpdateStatus()
		if err != nil {
			flog.Log.Errorf("UpdateStatusOfContent err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}
	resp.Flag = true
}

// 更新内容的节点
type UpdateNodesOfContentRequest struct {
	Id     int `json:"id" validate:"required"`
	NodeId int `json:"node_id" validate:"required"`
}

func UpdateNodeOfContent(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdateNodesOfContentRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("UpdateNodeOfContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		flog.Log.Errorf("UpdateNodeOfContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	contentBefore := new(model.Content)
	contentBefore.Id = req.Id
	contentBefore.UserId = uu.Id
	exist, err := contentBefore.Get()
	if err != nil {
		flog.Log.Errorf("UpdateNodeOfContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		flog.Log.Errorf("UpdateNodeOfContent err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	content := new(model.Content)
	content.Id = req.Id
	content.UserId = uu.Id
	if req.NodeId != contentBefore.NodeId {
		content.NodeId = req.NodeId

		contentNode := new(model.ContentNode)
		contentNode.Id = req.NodeId
		contentNode.UserId = uu.Id
		exist, err := contentNode.Get()
		if err != nil {
			flog.Log.Errorf("UpdateNodeOfContent err: %s", err.Error())
			resp.Error = Error(DBError, "")
			return
		}
		if !exist {
			flog.Log.Errorf("UpdateNodeOfContent err: %s", "node not found")
			resp.Error = Error(ContentNodeNotFound, "")
			return
		}

		// SEO变了，也要带上
		content.NodeSeo = contentNode.Seo
		content.SortNum = contentBefore.SortNum
		err = content.UpdateNode(contentBefore.NodeId)
		if err != nil {
			flog.Log.Errorf("UpdateNodeOfContent err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}
	resp.Flag = true
}

// 更新内容置顶
type UpdateTopOfContentRequest struct {
	Id  int `json:"id" validate:"required"`
	Top int `json:"top" validate:"oneof=0 1"`
}

func UpdateTopOfContent(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdateTopOfContentRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("UpdateTopOfContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		flog.Log.Errorf("UpdateTopOfContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	contentBefore := new(model.Content)
	contentBefore.Id = req.Id
	contentBefore.UserId = uu.Id
	exist, err := contentBefore.Get()
	if err != nil {
		flog.Log.Errorf("UpdateTopOfContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		flog.Log.Errorf("UpdateTopOfContent err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	content := new(model.Content)
	content.Id = req.Id
	content.UserId = uu.Id
	if req.Top != contentBefore.Top {
		content.Top = req.Top
		_, err = content.UpdateTop()
		if err != nil {
			flog.Log.Errorf("UpdateTopOfContent err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}
	resp.Flag = true
}

// 更新内容置顶
type UpdatePasswordOfContentRequest struct {
	Id       int    `json:"id" validate:"required"`
	Password string `json:"password"`
}

func UpdatePasswordOfContent(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdatePasswordOfContentRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("UpdatePasswordOfContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		flog.Log.Errorf("UpdatePasswordOfContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	contentBefore := new(model.Content)
	contentBefore.Id = req.Id
	contentBefore.UserId = uu.Id
	exist, err := contentBefore.Get()
	if err != nil {
		flog.Log.Errorf("UpdatePasswordOfContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		flog.Log.Errorf("UpdatePasswordOfContent err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	content := new(model.Content)
	content.Id = req.Id
	content.UserId = uu.Id
	if req.Password != contentBefore.Password {
		content.Password = req.Password
		_, err = content.UpdatePassword()
		if err != nil {
			flog.Log.Errorf("UpdatePasswordOfContent err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}
	resp.Flag = true
}

// 更新内容标题和具体内容
type UpdateInfoOfContentRequest struct {
	Id       int    `json:"id" validate:"required"`
	Title    string `json:"title" validate:"required,lt=100"`
	Describe string `json:"describe" validate:"omitempty"`
}

func UpdateInfoOfContent(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdateInfoOfContentRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("UpdateInfoOfContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		flog.Log.Errorf("UpdateInfoOfContent err: %s", err.Error())
		resp.Error = Error(I500, "")
		return
	}

	contentBefore := new(model.Content)
	contentBefore.Id = req.Id
	contentBefore.UserId = uu.Id
	exist, err := contentBefore.Get()
	if err != nil {
		flog.Log.Errorf("UpdateInfoOfContent err: %s", err.Error())
		resp.Error = Error(DBError, "")
		return
	}

	if !exist {
		flog.Log.Errorf("UpdateInfoOfContent err: %s", "content not found")
		resp.Error = Error(DbNotFound, "content not found")
		return
	}

	content := new(model.Content)
	content.Id = req.Id
	content.UserId = uu.Id

	//  如果内容更新
	if contentBefore.PreDescribe != req.Describe || contentBefore.PreTitle != req.Title {
		// 一旦更新就这样
		content.Describe = req.Describe
		content.Title = req.Title
		err = content.UpdateDescribeAndHistory()
		if err != nil {
			flog.Log.Errorf("UpdateInfoOfContent err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}
	resp.Flag = true
}

// 将Y放在X的上面
// 内容相对节点简单点，没有层次
type SortContentRequest struct {
	XID int `json:"xid" validate:"required"`
	YID int `json:"yid"`
}

//  拖曳排序超级函数
func SortContent(c *gin.Context) {
	resp := new(Resp)
	req := new(SortContentRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("SortContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	if req.XID == req.YID {
		flog.Log.Errorf("SortContent err: %s", "xid=yid not right")
		resp.Error = Error(ParasError, "xid=yid not right")
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		flog.Log.Errorf("SortContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	x := new(model.Content)
	x.Id = req.XID
	x.UserId = uu.Id
	exist, err := x.Get()
	if err != nil {
		flog.Log.Errorf("SortContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		flog.Log.Errorf("SortContent err: %s", "x node not found")
		resp.Error = Error(ContentNotFound, "x node not found")
		return
	}

	// x节点要拉到最下面
	if req.YID == 0 {
		session := config.FafaRdb.Client.NewSession()
		defer session.Close()

		err = session.Begin()
		if err != nil {
			flog.Log.Errorf("SortContent err: %s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}

		// 比x小的都往上走，因为x要做垫底小弟
		_, err = session.Exec("update fafacms_content SET sort_num=sort_num+1 where sort_num < ? and user_id = ? and node_id = ?", x.SortNum, uu.Id, x.NodeId)
		if err != nil {
			session.Rollback()
			flog.Log.Errorf("SortContent err: %s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}

		// x做小弟
		_, err = session.Exec("update fafacms_content SET sort_num=0 where user_id = ? and node_id = ? and id = ?", uu.Id, x.NodeId, x.Id)
		if err != nil {
			session.Rollback()
			flog.Log.Errorf("SortContent err: %s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
		resp.Flag = true
		return
	}

	y := new(model.Content)
	y.Id = req.YID
	y.UserId = uu.Id
	exist, err = y.Get()
	if err != nil {
		flog.Log.Errorf("SortContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		flog.Log.Errorf("SortContent err: %s", "y node not found")
		resp.Error = Error(ContentNotFound, "y node not found")
		return
	}

	if x.NodeId != y.NodeId {
		flog.Log.Errorf("SortContent err: %s", "x y node are different")
		resp.Error = Error(ContentsAreInDifferentNode, "")
		return
	}

	session := config.FafaRdb.Client.NewSession()
	defer session.Close()

	err = session.Begin()
	if err != nil {
		flog.Log.Errorf("SortContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	// 先把x假装删掉，比x大的都-1，依次顶上x的位置
	_, err = session.Exec("update fafacms_content SET sort_num=sort_num-1 where sort_num > ? and user_id = ? and node_id = ?", x.SortNum, uu.Id, x.NodeId)
	if err != nil {
		session.Rollback()
		flog.Log.Errorf("SortContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	// 把大于y排序的节点都+1，腾出位置给x
	_, err = session.Exec("update fafacms_content SET sort_num=sort_num+1 where sort_num > ? and user_id = ? and node_id = ?", y.SortNum, uu.Id, y.NodeId)
	if err != nil {
		session.Rollback()
		flog.Log.Errorf("SortContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	// x顶上, 且y>x
	if y.SortNum > x.SortNum {
		_, err = session.Exec("update fafacms_content SET sort_num=? where user_id = ? and id = ?", y.SortNum, uu.Id, x.Id)
	} else {
		// 否则
		_, err = session.Exec("update fafacms_content SET sort_num=? where user_id = ? and id = ?", y.SortNum+1, uu.Id, x.Id)
	}
	if err != nil {
		session.Rollback()
		flog.Log.Errorf("SortContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	err = session.Commit()
	if err != nil {
		session.Rollback()
		flog.Log.Errorf("SortContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	resp.Flag = true
	return
}

// 发布内容
type PublishContentRequest struct {
	Id int `json:"id" validate:"required"`
}

func PublishContent(c *gin.Context) {
	resp := new(Resp)
	req := new(PublishContentRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("PublishContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		flog.Log.Errorf("PublishContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	content := new(model.Content)
	content.Id = req.Id
	content.UserId = uu.Id
	exist, err := content.Get()
	if err != nil {
		flog.Log.Errorf("PublishContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		flog.Log.Errorf("PublishContent err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	if content.PreFlush == 1 {
		resp.Flag = true
		return
	}

	err = content.PublishDescribe()
	if err != nil {
		flog.Log.Errorf("PublishContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	resp.Flag = true
}

// 从历史版本恢复，只需要历史ID
type RestoreContentRequest struct {
	HistoryId int `json:"history_id" validate:"required"`
}

func RestoreContent(c *gin.Context) {
	resp := new(Resp)
	req := new(RestoreContentRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("RestoreContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		flog.Log.Errorf("RestoreContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	contentH := new(model.ContentHistory)
	contentH.Id = req.HistoryId
	contentH.UserId = uu.Id
	exist, err := contentH.GetRaw()
	if err != nil {
		flog.Log.Errorf("RestoreContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		flog.Log.Errorf("RestoreContent err: %s", "content history not found")
		resp.Error = Error(ContentHistoryNotFound, "")
		return
	}

	content := new(model.Content)
	content.Id = contentH.ContentId
	content.UserId = uu.Id
	exist, err = content.Get()
	if err != nil {
		flog.Log.Errorf("RestoreContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		flog.Log.Errorf("RestoreContent err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	content.Title = contentH.Title
	content.Describe = contentH.Describe
	err = content.ResetDescribe()
	if err != nil {
		flog.Log.Errorf("RestoreContent err: %s", err.Error())
		resp.Error = Error(DBError, "")
		return
	}
	resp.Flag = true
}

type ListContentRequest struct {
	Id               int      `json:"id"`
	Seo              string   `json:"seo" validate:"omitempty,alphanumunicode,gt=3,lt=30"`
	NodeId           int      `json:"node_id"`
	NodeSeo          string   `json:"node_seo"`
	Top              int      `json:"top" validate:"oneof=-1 0 1"`
	Status           int      `json:"status" validate:"oneof=-1 0 1 2 3"`
	CloseComment     int      `json:"close_comment" validate:"oneof=-1 0 1 2"`
	UserId           int      `json:"user_id"`
	UserName         string   `json:"user_name"`
	CreateTimeBegin  int64    `json:"create_time_begin"`
	CreateTimeEnd    int64    `json:"create_time_end"`
	UpdateTimeBegin  int64    `json:"update_time_begin"`
	UpdateTimeEnd    int64    `json:"update_time_end"`
	PublishTimeBegin int64    `json:"publish_time_begin"`
	PublishTimeEnd   int64    `json:"publish_time_end"`
	Sort             []string `json:"sort" validate:"dive,lt=100"`
	PageHelp
}

type ListContentResponse struct {
	Contents []model.Content `json:"contents"`
	PageHelp
}

func ListContent(c *gin.Context) {
	resp := new(Resp)
	uu, err := GetUserSession(c)
	if err != nil {
		flog.Log.Errorf("ListContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		JSONL(c, 200, nil, resp)
		return
	}

	uid := uu.Id
	ListContentHelper(c, uid)
}

func ListContentAdmin(c *gin.Context) {
	ListContentHelper(c, 0)
}

func ListContentHelper(c *gin.Context, userId int) {
	resp := new(Resp)

	respResult := new(ListContentResponse)
	req := new(ListContentRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("ListContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	// new query list session
	session := config.FafaRdb.Client.NewSession()
	defer session.Close()

	// group list where prepare
	session.Table(new(model.Content)).Where("1=1")

	// query prepare
	if req.Id != 0 {
		session.And("id=?", req.Id)
	}

	if userId != 0 {
		session.And("user_id=?", userId)

		// 不设置条件，只列出非垃圾
		if req.Status == -1 {
			session.And("status!=?", 3)
		} else {
			session.And("status=?", req.Status)
		}

	} else {
		if req.Status != -1 {
			session.And("status=?", req.Status)
		}
		if req.UserName != "" {
			session.And("user_name=?", req.UserName)
		}
		if req.UserId != 0 {
			session.And("user_id=?", req.UserId)
		}
	}

	if req.Top != -1 {
		session.And("top=?", req.Top)
	}

	if req.Seo != "" {
		session.And("seo=?", req.Seo)
	}

	if req.CloseComment != -1 {
		session.And("close_comment=?", req.CloseComment)
	}

	if req.NodeId != 0 {
		session.And("node_id=?", req.NodeId)
	}

	if req.NodeSeo != "" {
		session.And("node_seo=?", req.NodeSeo)
	}

	if req.CreateTimeBegin > 0 {
		session.And("create_time>=?", req.CreateTimeBegin)
	}

	if req.CreateTimeEnd > 0 {
		session.And("create_time<?", req.CreateTimeBegin)
	}

	if req.UpdateTimeBegin > 0 {
		session.And("update_time>=?", req.UpdateTimeBegin)
	}

	if req.UpdateTimeEnd > 0 {
		session.And("update_time<?", req.UpdateTimeEnd)
	}

	if req.PublishTimeBegin > 0 {
		session.And("publish_time>=?", req.PublishTimeBegin)
	}

	if req.PublishTimeEnd > 0 {
		session.And("publish_time<?", req.PublishTimeEnd)
	}

	// count num
	countSession := session.Clone()
	defer countSession.Close()
	total, err := countSession.Count()
	if err != nil {
		flog.Log.Errorf("ListContent err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	// if count>0 start list
	cs := make([]model.Content, 0)
	p := &req.PageHelp
	if total == 0 {
	} else {
		// sql build
		p.build(session, req.Sort, model.ContentSortName)
		// do query
		err = session.Omit("describe", "pre_describe").Find(&cs)
		if err != nil {
			flog.Log.Errorf("ListContent err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}

	// result
	respResult.Contents = cs
	p.Pages = int(math.Ceil(float64(total) / float64(p.Limit)))
	respResult.PageHelp = *p
	resp.Data = respResult
	resp.Flag = true
}

type ListContentHistoryRequest struct {
	Id     int      `json:"id" validate:"required"`
	UserId int      `json:"user_id"`
	Sort   []string `json:"sort" validate:"dive,lt=100"`
	PageHelp
}

type ListContentHistoryResponse struct {
	Contents []model.ContentHistory `json:"contents"`
	PageHelp
}

func ListContentHistory(c *gin.Context) {
	resp := new(Resp)
	uu, err := GetUserSession(c)
	if err != nil {
		flog.Log.Errorf("ListContentHistory err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		JSONL(c, 200, nil, resp)
		return
	}

	uid := uu.Id
	ListContentHistoryHelper(c, uid)
}

func ListContentHistoryAdmin(c *gin.Context) {
	ListContentHistoryHelper(c, 0)
}

func ListContentHistoryHelper(c *gin.Context, userId int) {
	resp := new(Resp)

	respResult := new(ListContentHistoryResponse)
	req := new(ListContentHistoryRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("ListContentHistory err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	// new query list session
	session := config.FafaRdb.Client.NewSession()
	defer session.Close()

	// group list where prepare
	session.Table(new(model.ContentHistory)).Where("1=1")

	session.And("content_id=?", req.Id)

	if userId != 0 {
		session.And("user_id=?", userId)
	} else {
		if req.UserId != 0 {
			session.And("user_id=?", req.UserId)
		}
	}

	// count num
	countSession := session.Clone()
	defer countSession.Close()
	total, err := countSession.Count()
	if err != nil {
		flog.Log.Errorf("ListContentHistory err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	// if count>0 start list
	cs := make([]model.ContentHistory, 0)
	p := &req.PageHelp
	if total == 0 {
	} else {
		// sql build
		p.build(session, req.Sort, model.ContentHistorySortName)
		// do query
		err = session.Omit("describe").Find(&cs)
		if err != nil {
			flog.Log.Errorf("ListContentHistory err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}

	// result
	respResult.Contents = cs
	p.Pages = int(math.Ceil(float64(total) / float64(p.Limit)))
	respResult.PageHelp = *p
	resp.Data = respResult
	resp.Flag = true
}

type TakeContentRequest struct {
	Id int `json:"id" validate:"required"`
}

// 获取内容
func TakeContentHelper(c *gin.Context, userId int) {
	resp := new(Resp)
	req := new(TakeContentRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("TakeContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	content := new(model.Content)
	content.Id = req.Id
	content.UserId = userId
	exist, err := content.Get()
	if err != nil {
		flog.Log.Errorf("TakeContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		flog.Log.Errorf("TakeContent err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	resp.Data = content
	resp.Flag = true
}

func TakeContent(c *gin.Context) {
	resp := new(Resp)
	uu, err := GetUserSession(c)
	if err != nil {
		flog.Log.Errorf("TakeContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		JSONL(c, 200, nil, resp)
		return
	}

	uid := uu.Id
	TakeContentHelper(c, uid)
}

func TakeContentAdmin(c *gin.Context) {
	TakeContentHelper(c, 0)
}

type TakeContentHistoryRequest struct {
	Id int `json:"id" validate:"required"`
}

func TakeContentHistoryHelper(c *gin.Context, userId int) {
	resp := new(Resp)
	req := new(TakeContentHistoryRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("TakeContentHistory err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	content := new(model.ContentHistory)
	content.Id = req.Id
	content.UserId = userId
	exist, err := content.GetRaw()
	if err != nil {
		flog.Log.Errorf("TakeContentHistory err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		flog.Log.Errorf("TakeContentHistory err: %s", "content history not found")
		resp.Error = Error(ContentHistoryNotFound, "")
		return
	}

	resp.Data = content
	resp.Flag = true
}

func TakeContentHistory(c *gin.Context) {
	resp := new(Resp)
	uu, err := GetUserSession(c)
	if err != nil {
		flog.Log.Errorf("TakeContentHistory err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		JSONL(c, 200, nil, resp)
		return
	}

	uid := uu.Id
	TakeContentHistoryHelper(c, uid)
}

func TakeContentHistoryAdmin(c *gin.Context) {
	TakeContentHistoryHelper(c, 0)
}

type SentContentToRubbishRequest struct {
	Id int `json:"id" validate:"required"`
}

// 垃圾回收，假删除(容忍禁止的内容丢出去.这样是个漏洞，允许他 !!!)
func SentContentToRubbish(c *gin.Context) {
	resp := new(Resp)
	req := new(SentContentToRubbishRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("SentContentToRubbish err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		flog.Log.Errorf("SentContentToRubbish err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	contentBefore := new(model.Content)
	contentBefore.Id = req.Id
	contentBefore.UserId = uu.Id
	exist, err := contentBefore.Get()
	if err != nil {
		flog.Log.Errorf("SentContentToRubbish err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		flog.Log.Errorf("SentContentToRubbish err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	if contentBefore.Status == 3 {
		resp.Flag = true
		return
	}

	content := new(model.Content)
	content.Id = req.Id
	content.UserId = uu.Id
	content.Status = 3
	_, err = content.UpdateStatus()
	if err != nil {
		flog.Log.Errorf("SentContentToRubbish err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	resp.Flag = true
}

type ReCycleOfContentInRubbishRequest struct {
	Id int `json:"id" validate:"required"`
}

// 垃圾恢复
func ReCycleOfContentInRubbish(c *gin.Context) {
	resp := new(Resp)
	req := new(SentContentToRubbishRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("ReCycleOfContentInRubbish err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		flog.Log.Errorf("ReCycleOfContentInRubbish err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	contentBefore := new(model.Content)
	contentBefore.Id = req.Id
	contentBefore.UserId = uu.Id
	exist, err := contentBefore.Get()
	if err != nil {
		flog.Log.Errorf("ReCycleOfContentInRubbish err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		flog.Log.Errorf("ReCycleOfContentInRubbish err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	if contentBefore.Status == 3 {
		content := new(model.Content)
		content.Id = req.Id
		content.UserId = uu.Id
		content.Status = 0
		_, err = content.UpdateStatus()
		if err != nil {
			flog.Log.Errorf("ReCycleOfContentInRubbish err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}

	resp.Flag = true
}

type ReallyDeleteContentRequest struct {
	Id int `json:"id" validate:"required"`
}

func ReallyDeleteContent(c *gin.Context) {
	resp := new(Resp)
	req := new(ReallyDeleteContentRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("ReallyDeleteContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		flog.Log.Errorf("ReallyDeleteContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	contentBefore := new(model.Content)
	contentBefore.Id = req.Id
	contentBefore.UserId = uu.Id
	exist, err := contentBefore.Get()
	if err != nil {
		flog.Log.Errorf("ReallyDeleteContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		flog.Log.Errorf("ReallyDeleteContent err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	// 只有回收站的才能删除
	if contentBefore.Status == 3 {
		content := new(model.Content)
		content.Id = req.Id
		content.UserId = uu.Id
		err = content.Delete()
		if err != nil {
			flog.Log.Errorf("ReallyDeleteContent err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}

	resp.Flag = true
}
