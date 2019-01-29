package main

import (
	"flag"
	"github.com/hunterhug/fafacms/core/config"
	"github.com/hunterhug/fafacms/core/controllers"
	"github.com/hunterhug/fafacms/core/flog"
	"github.com/hunterhug/fafacms/core/model"
	"github.com/hunterhug/fafacms/core/router"
	"github.com/hunterhug/fafacms/core/server"
)

var (
	configFile  string
	createTable bool
)

func init() {
	flag.StringVar(&configFile, "config", "./config.json", "config file")
	flag.BoolVar(&createTable, "db", true, "create db table")
	flag.Parse()
}

func main() {
	var err error

	// Init Config
	err = server.InitConfig(configFile)
	if err != nil {
		panic(err)
	}

	// Init Log
	flog.InitLog(config.FafaConfig.DefaultConfig.LogPath)
	if config.FafaConfig.DefaultConfig.Debug {
		flog.SetLogLevel("DEBUG")
	}

	flog.Log.Notice("Hi! FaFa CMS!")
	flog.Log.Debugf("Hi! Config is %#v", config.FafaConfig)

	// Init Db
	err = server.InitRdb(config.FafaConfig.DbConfig)
	if err != nil {
		panic(err)
	}

	err = server.InitSession(config.FafaConfig.SessionConfig)
	if err != nil {
		panic(err)
	}

	// Create Table, Here to init db
	if createTable {
		server.CreateTable([]interface{}{
			model.User{},
			model.Group{},
			model.Resource{},
			model.GroupResource{},
			model.Content{},
			model.ContentNode{},
			model.Comment{},
			model.Log{},
			model.Picture{},
		})
	}

	// Server Run
	engine := server.Server()

	// Storage static API
	engine.Static("/storage", config.FafaConfig.DefaultConfig.StoragePath)

	// Web welcome home!
	router.SetRouter(engine)

	// Auth API load
	controllers.InitAuthResource()

	// V1 API, will be change to V2...
	v1 := engine.Group("/v1")
	v1.Use(controllers.AuthFilter)

	// Base API no version
	base := engine.Group("/b")
	base.Use(controllers.AuthFilter)

	// Router Set
	router.SetAPIRouter(v1, router.V1Router)
	router.SetAPIRouter(base, router.BaseRouter)

	config.Log.Noticef("Server run in %s", config.FafaConfig.DefaultConfig.WebPort)
	err = engine.Run(config.FafaConfig.DefaultConfig.WebPort)
	if err != nil {
		panic(err)
	}
}