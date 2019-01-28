package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/hunterhug/fafacms/core/config"
)

func Login(c *gin.Context) {
	resp := new(config.Resp)
	defer func() {
		c.JSON(200, resp)
	}()
}
