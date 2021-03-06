package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/hunterhug/fafacms/core/flog"
	"github.com/hunterhug/fafacms/core/model"
	"github.com/hunterhug/parrot/util"
)

type LoginRequest struct {
	UserName string `json:"user_name"`
	PassWd   string `json:"pass_wd"`
	Remember bool   `json:"remember"`
}

func Login(c *gin.Context) {
	resp := new(Resp)
	req := new(LoginRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	// paras not empty
	if req.UserName == "" || req.PassWd == "" {
		flog.Log.Errorf("login err:%s", "paras wrong")
		resp.Error = Error(ParasError, "field username or pass_wd")
		return
	}
	// check session
	userInfo, _ := GetUserSession(c)
	if userInfo != nil {
		//c.Set("skipLog", true)
		c.Set("uid", userInfo.Id)
		resp.Flag = true
		return
	}

	// check cookie
	success, userInfo := CheckCookie(c)
	if success {
		err := SetUserSession(c, userInfo)
		if err != nil {
			flog.Log.Errorf("login err:%s", err.Error())
			resp.Error = Error(SetUserSessionError, err.Error())
			return
		}

		c.Set("uid", userInfo.Id)
		resp.Flag = true
		return
	}

	// super root user login
	if req.UserName == "hunterhug" && req.PassWd == "hunterhug" {
		u := new(model.User)
		u.Id = -1
		err := SetUserSession(c, u)
		if err != nil {
			flog.Log.Errorf("login err:%s", err.Error())
		}

		c.Set("uid", u.Id)
		resp.Flag = true
		return
	}

	// common people login
	uu := new(model.User)
	uu.Name = req.UserName
	uu.Password = req.PassWd
	ok, err := uu.GetRaw()
	if err != nil {
		flog.Log.Errorf("login err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !ok {
		flog.Log.Errorf("login err:%s", "user or password wrong")
		resp.Error = Error(LoginWrong, "")
		return
	}

	c.Set("uid", uu.Id)

	err = SetUserSession(c, uu)
	if err != nil {
		flog.Log.Errorf("login err:%s", err.Error())
		resp.Error = Error(SetUserSessionError, err.Error())
		return
	}

	if req.Remember {
		authKey := util.Md5(c.ClientIP() + "|" + uu.Password)
		secretKey := util.IS(uu.Id) + "|" + authKey
		c.SetCookie("auth", secretKey, 3600*24*7, "/", "", false, true)
	} else {
		// cookie clean
		c.SetCookie("auth", "", -1, "/", "", false, true)
	}

	resp.Flag = true
}

func Logout(c *gin.Context) {
	resp := new(Resp)
	defer func() {
		JSON(c, 200, resp)
	}()
	user, _ := GetUserSession(c)
	if user != nil {
		DeleteUserSession(c)
	}
	c.SetCookie("auth", "", 0, "", "", false, false)
	resp.Flag = true
}
