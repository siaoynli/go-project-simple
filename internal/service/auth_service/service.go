package auth_service

import (
	"github.com/siaoynli/go-project-simple/internal/server/api/api_response"
	"github.com/siaoynli/pkg/cache"
	"gorm.io/gorm"
)

type Service interface {
	DetailByKey(ctx *api_response.Gin, key string) (data *CacheAuthorizedData, err error)
}

type service struct {
	db    *gorm.DB
	cache *cache.Redis
}

func New(db *gorm.DB, cache *cache.Redis) Service {
	return &service{
		db:    db,
		cache: cache,
	}
}
