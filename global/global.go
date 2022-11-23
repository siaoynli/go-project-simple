package global

import (
	"github.com/siaoynli/pkg/cache"
	"github.com/siaoynli/pkg/es"
	"github.com/siaoynli/pkg/nosql"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ES    *es.Client
	LOG   *zap.Logger
	DB    *gorm.DB
	CACHE *cache.Redis
	Mongo *nosql.MgClient
)
