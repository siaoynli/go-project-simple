package bootstrap

import (
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/siaoynli/go-project-simple/config"
	"github.com/siaoynli/go-project-simple/global"
	"github.com/siaoynli/go-project-simple/metric"

	"github.com/siaoynli/pkg/cache"
	"github.com/siaoynli/pkg/db"
	"github.com/siaoynli/pkg/es"
	"github.com/siaoynli/pkg/httpclient"
	"github.com/siaoynli/pkg/logger"
	"github.com/siaoynli/pkg/nosql"
	"github.com/siaoynli/pkg/prome"
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
		logger.WithDisableConsole(),
		logger.WithTimeLayout(timeutil.CSTLayout),
		logger.WithFileRotationP(&logger.LogOptions{}),
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
