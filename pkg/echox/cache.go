package echox

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

type cache struct {
	maxAgeMs int32
	cache    map[string]userState
}

func newCache(maxAgeMs int32) *cache {
	return &cache{
		maxAgeMs: maxAgeMs,
		cache:    make(map[string]userState),
	}
}

type userState struct {
	stored time.Time
	User   UserInfo
}

func (c *cache) Get(token *jwt.Token) (cached UserInfo, ok bool) {
	now := time.Now()
	if hit, ok := c.cache[token.Raw]; ok {
		if now.Sub(hit.stored) < (time.Duration(c.maxAgeMs) * time.Millisecond) {
			return hit.User, true
		} else {
			delete(c.cache, token.Raw)
		}
	}
	return UserInfo{}, false
}

func (c *cache) Put(token *jwt.Token, u UserInfo) {
	now := time.Now()
	c.cache[token.Raw] = userState{
		stored: now,
		User:   u,
	}
}
