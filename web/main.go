package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/siaoynli/go-project-simple/config"
	"github.com/siaoynli/go-project-simple/global"
	"github.com/siaoynli/go-project-simple/internal/server/api"
	_ "github.com/siaoynli/go-project-simple/web/bootstrap"
	"github.com/siaoynli/pkg/cache"
	"github.com/siaoynli/pkg/db"
	"github.com/siaoynli/pkg/es"
	"github.com/siaoynli/pkg/shutdown"
	"go.uber.org/zap"
)

func main() {
	router := api.InitRouter()
	listenAddr := fmt.Sprintf(":%d", config.Cfg.App.HttpPort)
	global.LOG.Warn("start http server", zap.String("listenAddr", listenAddr))
	server := &http.Server{
		Addr:           listenAddr,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			global.LOG.Error("http server start error", zap.Error(err))
		}
	}()

	//优雅关闭
	shutdown.NewHook().Close(
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			if err := server.Shutdown(ctx); err != nil {
				global.LOG.Error("http server shutdown err", zap.Error(err))
			}
		},

		func() {
			es.CloseAll()
		},
		func() {
			//关闭mysql
			if err := db.CloseMysqlClient(db.DefaultClient); err != nil {
				global.LOG.Error("mysql shutdown err", zap.Error(err), zap.String("client", db.DefaultClient))
			}
		},

		func() {
			err := global.CACHE.Close()
			if err != nil {
				global.LOG.Error("redis close error", zap.Error(err), zap.String("client", cache.DefaultRedisClient))
			}
		},
		func() {
			if global.Mongo != nil {
				global.Mongo.Close()
			}
		},
		func() {
			err := global.LOG.Sync()
			if err != nil {
				fmt.Println(err)
			}
		},
	)
}
