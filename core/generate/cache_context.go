package generate

import "github.com/railwayapp/railpack-go/core/plan"

const (
	APT_CACHE_KEY  = "apt"
	MISE_CACHE_KEY = "mise"
)

type CacheContext struct {
	Caches map[string]*plan.Cache
}

func NewCacheContext() *CacheContext {
	return &CacheContext{
		Caches: make(map[string]*plan.Cache),
	}
}

func (c *CacheContext) AddCache(name string, directory string) string {
	c.Caches[name] = plan.NewCache(directory)
	return name
}

func (c *CacheContext) GetCache(name string) *plan.Cache {
	return c.Caches[name]
}

func (c *CacheContext) GetAptCache() string {
	if _, ok := c.Caches[APT_CACHE_KEY]; !ok {
		aptCache := plan.NewCache("/var/cache/apt")
		aptCache.Type = plan.CacheTypeLocked
		c.Caches[APT_CACHE_KEY] = aptCache
	}

	return APT_CACHE_KEY
}
