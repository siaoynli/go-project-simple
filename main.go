package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/siaoynli/go-project-simple/config"
	"github.com/siaoynli/go-project-simple/global"
	"github.com/siaoynli/go-project-simple/internal/server/api"
	"github.com/siaoynli/go-project-simple/metric"
	"github.com/siaoynli/pkg/cache"
	"github.com/siaoynli/pkg/db"
	"github.com/siaoynli/pkg/es"
	"github.com/siaoynli/pkg/httpclient"
	"github.com/siaoynli/pkg/logger"
	"github.com/siaoynli/pkg/nosql"
	"github.com/siaoynli/pkg/prome"
	"github.com/siaoynli/pkg/shutdown"
	"github.com/siaoynli/pkg/timeutil"
	"github.com/siaoynli/pkg/trace"
	"go.uber.org/zap"
)

func init() {
	config.LoadConfig()
	InitLog()
	initMysqlClient()
	initRedisClient()
	initMongoClient()
	initESClient()
	// initProme()
}
func InitLog() {
	// 初始化 logger
	global.LOG = logger.InitLogger(
		//logger.WithDisableConsole(),
		logger.WithTimeLayout(timeutil.CSTLayout),
		logger.WithFileRotationP(config.Cfg.App.AppLogPath),
	)
}
func initMysqlClient() {
	mysqlCfg := config.Cfg.Mysql
	logger.Warn("mysqlCfg", zap.Any("", mysqlCfg))
	err := db.InitMysqlClient(db.DefaultClient, mysqlCfg.User, mysqlCfg.Password, mysqlCfg.Host, mysqlCfg.DBName)
	if err != nil {
		global.LOG.Error("mysql init error", zap.Error(err))
		panic("initMysqlClient error")
	}
	global.DB = db.GetMysqlClient(db.DefaultClient).DB
}
func initRedisClient() {
	redisCfg := config.Cfg.Redis
	opt := redis.Options{
		Addr:         redisCfg.Host,
		Password:     redisCfg.Password,
		DB:           redisCfg.DB,
		MaxRetries:   redisCfg.MaxRetries,
		PoolSize:     redisCfg.PoolSize,
		MinIdleConns: redisCfg.MinIdleConn,
	}
	redisTrace := trace.Cache{
		Name:                  "redis",
		SlowLoggerMillisecond: 500,
		Logger:                logger.GetLogger(),
		AlwaysTrace:           config.Cfg.App.RunMode == config.RunModeDev,
	}
	err := cache.InitRedis(cache.DefaultRedisClient, &opt, &redisTrace)
	if err != nil {
		global.LOG.Error("redis init error", zap.Error(err))
		panic("initRedisClient error")
	}
	global.CACHE = cache.GetRedisClient(cache.DefaultRedisClient)
}

func initESClient() {
	ESCfg := config.Cfg.Elasticsearch
	err := es.InitClientWithOptions(es.DefaultClient, ESCfg.Host,
		ESCfg.User,
		ESCfg.Password,
		es.WithScheme("https"))
	if err != nil {
		logger.Error("InitClientWithOptions error", zap.Error(err), zap.String("client", es.DefaultClient))
		panic(err)
	}
	global.ES = es.GetClient(es.DefaultClient)
}

func initMongoClient() {
	err := nosql.InitMongoClient(nosql.DefaultMongoClient, config.Cfg.MongoDB.User,
		config.Cfg.MongoDB.Password, config.Cfg.MongoDB.Host, 200)
	if err != nil {
		logger.Error("InitMongoClient error", zap.Error(err), zap.String("client", nosql.DefaultMongoClient))
		//panic(err)
	} else {
		global.Mongo = nosql.GetMongoClient(nosql.DefaultMongoClient)
	}

}

func initProme() {
	prome.InitPromethues(config.Cfg.Prome.Host, time.Second*60, config.AppName, httpclient.DefaultClient, metric.ProductSearch)
}
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
