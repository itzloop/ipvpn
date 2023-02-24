package cache

import (
	"github.com/itzloop/ipvpn/pkg/config"
	"github.com/patrickmn/go-cache"
	"time"
)

func NewCache(cfg *config.Config) *cache.Cache {
	// TODO configure
	return cache.New(30*time.Minute, 15*time.Minute)
}
