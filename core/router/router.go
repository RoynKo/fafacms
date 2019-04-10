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

var (
	HomeRouter = map[string]HttpHandle{
		"/":                {"Home", controllers.Home, GP, false},
		"/login":           {"User Login", controllers.Login, GP, false},
		"/logout":          {"User Logout", controllers.Logout, GP, false},
		"/register":        {"User Register", controllers.RegisterUser, GP, false},
		"/activate":        {"User Verify Email To Activate", controllers.ActivateUser, GP, false},
		"/activate/code":   {"User Resend Email Activate Code", controllers.ResendActivateCodeToUser, GP, false},
		"/password/forget": {"User Forget Password Gen Code", controllers.ForgetPassword, GP, false},
		"/password/change": {"User Change Password", controllers.ChangePassword, GP, false},
	}

	// /v1/user/create
	// need login group auth
	V1Router = map[string]HttpHandle{

		"/group/create": {"Create Group", controllers.CreateGroup, POST, true},
		"/group/update": {"Update Group", controllers.UpdateGroup, POST, true},
		"/group/delete": {"Delete Group", controllers.DeleteGroup, POST, true},
		"/group/take":   {"Take Group", controllers.TakeGroup, GP, true},
		"/group/list":   {"List Group", controllers.ListGroup, GP, true},

		"/user/list":   {"User List All", controllers.ListUser, GP, true},
		"/user/info":   {"User Info Self", controllers.TakeUser, GP, false},
		"/user/update": {"User Update Self", controllers.UpdateUser, GP, false},


		// set/update User groupid
		// update group resource
		//"/resource/create": {controllers.CreateResource, POST},
		//"/resource/update": {controllers.UpdateResource, POST},
		//"/resource/delete": {controllers.DeleteResource, POST},
		//"/resource/take":   {controllers.TakeResource, GP},
		//"/resource/list":   {controllers.ListResource, GP},
		//
		//"/auth/update": {controllers.UpdateAuth, GP},
		//

		// here important
		//"/node/create": {controllers.CreateNode, POST},
		//"/node/update": {controllers.UpdateNode, POST},
		//"/node/delete": {controllers.DeleteNode, POST},
		//"/node/take":   {controllers.TakeNode, GP},
		//"/node/list":   {controllers.ListNode, GP},
		//
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

	// /b/upload
	// need login group auth
	BaseRouter = map[string]HttpHandle{
		"/upload": {"File Upload", controllers.Upload, POST, false},
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
