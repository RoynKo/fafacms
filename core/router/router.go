package router

import (
	"github.com/gin-gonic/gin"
	"github.com/hunterhug/fafacms/core/controllers"
)

type HttpHandle struct {
	Name   string
	Func   gin.HandlerFunc
	Method []string
	Admin  bool
}

var (
	POST = []string{"POST"}
	GET  = []string{"GET"}
	GP   = []string{"POST", "GET"}
)

// 路由，最后一个参数表示是否需要管理权限
var (
	HomeRouter = map[string]HttpHandle{
		// 前端路由
		"/":                       {"Home", controllers.Home, GP, false},
		"/u/:name":                {"user home page", controllers.Home, GP, false},
		"/u/:name/:node":          {"user node page", controllers.Home, GP, false},
		"/u/:name/:node/:content": {"user content page", controllers.Home, GP, false},

		// 前端的用户授权路由，不需要登录即可操作
		"/login":           {"User Login", controllers.Login, GP, false},
		"/logout":          {"User Logout", controllers.Logout, GP, false},
		"/register":        {"User Register", controllers.RegisterUser, GP, false},
		"/activate":        {"User Verify Email To Activate", controllers.ActivateUser, GP, false},               // 用户自己激活
		"/activate/code":   {"User Resend Email Activate Code", controllers.ResendActivateCodeToUser, GP, false}, // 激活码过期重新获取
		"/password/forget": {"User Forget Password Gen Code", controllers.ForgetPasswordOfUser, GP, false},       // 忘记密码，验证码发往邮箱
		"/password/change": {"User Change Password", controllers.ChangePasswordOfUser, GP, false},                // 根据邮箱验证码修改密码
	}

	// /v1/user/create
	// need login group auth
	V1Router = map[string]HttpHandle{
		// 用户组操作
		"/group/create": {"Create Group", controllers.CreateGroup, POST, true},
		"/group/update": {"Update Group", controllers.UpdateGroup, POST, true},
		"/group/delete": {"Delete Group", controllers.DeleteGroup, POST, true},
		"/group/take":   {"Take Group", controllers.TakeGroup, GP, true},
		"/group/list":   {"List Group", controllers.ListGroup, GP, true},

		// 用户操作
		"/user/list":         {"User List All", controllers.ListUser, GP, true},              // 超级管理员列出用户列表
		"/user/create":       {"User Create", controllers.CreateUser, GP, true},              // 超级管理员创建用户，默认激活
		"/user/assign":       {"User Assign Group", controllers.AssignGroupToUser, GP, true}, // 超级管理员给用户分配用户组
		"/user/info":         {"User Info Self", controllers.TakeUser, GP, false},            // 获取自己的信息
		"/user/update":       {"User Update Self", controllers.UpdateUser, GP, false},        // 更新自己的信息
		"/user/admin/update": {"User Update Admin", controllers.UpdateUserAdmin, GP, true},   // 管理员修改其他用户信息

		// 资源操作
		"/resource/list":   {"Resource List All", controllers.ListResource, GP, true},               // 列出资源
		"/resource/assign": {"Resource Assign Group", controllers.AssignGroupAndResource, GP, true}, // 资源分配给组

		// 文件操作
		"/file/upload":       {"File Upload", controllers.UploadFile, POST, false},
		"/file/list":         {"File List Self", controllers.ListFile, POST, false},
		"/file/update":       {"File Update Self", controllers.UpdateFile, POST, false},
		"/file/admin/list":   {"File List All", controllers.ListFileAdmin, POST, true},     // 管理员查看所有文件
		"/file/admin/update": {"File Update All", controllers.UpdateFileAdmin, POST, true}, // 管理员修改文件

		// 内容节点操作，管理员不关心，所有没有全局接口
		"/node/create": {"Create Node Self", controllers.CreateNode, POST, false},
		"/node/update": {"Update Node Self", controllers.UpdateNode, POST, false},
		"/node/delete": {"Delete Node Self", controllers.DeleteNode, POST, false},
		"/node/take":   {"Take Node Self", controllers.TakeNode, GP, false},
		"/node/list":   {"List Node Self", controllers.ListNode, GP, false},

		// 内容
		//"/content/create": {controllers.CreateContent, POST},
		//"/content/update": {controllers.UpdateContent, POST},
		//"/content/delete": {controllers.DeleteContent, POST},
		//"/content/take":   {controllers.TakeContent, GP},
		//"/content/list":   {controllers.ListContent, GP},
		//
		//"/comment/create": {controllers.CreateComment, POST},
		//"/comment/update": {controllers.UpdateComment, POST},
		//"/comment/delete": {controllers.DeleteComment, POST},
		//"/comment/take":   {controllers.TakeComment, GP},
		//"/comment/list":   {controllers.ListComment, GP},
	}
)

// home end.
func SetRouter(router *gin.Engine) {
	for url, app := range HomeRouter {
		for _, method := range app.Method {
			router.Handle(method, url, app.Func)
		}
	}
}

func SetAPIRouter(router *gin.RouterGroup, handles map[string]HttpHandle) {
	for url, app := range handles {
		for _, method := range app.Method {
			router.Handle(method, url, app.Func)
		}
	}
}
