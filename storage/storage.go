package storage

import (
	"sync"
	"time"
)

type CacheDataType string

const (
	String CacheDataType = "string"
	List   CacheDataType = "list"
	Hash   CacheDataType = "hash"
	Set    CacheDataType = "set"
	ZSet   CacheDataType = "zset"
)

type CacheDataUpdate struct {
	Value     *string
	Requests  *int
	Type      *CacheDataType
	CreatedAt *time.Time
	ExpiresAt *time.Time
}

type CacheData struct {
	Value     string
	Requests  int
	Type      CacheDataType
	CreatedAt time.Time
	ExpiresAt *time.Time
}

type Cache struct {
	data map[string]*CacheData
	mu   sync.RWMutex
}

func CreateCache() *Cache {
	return &Cache{
		data: make(map[string]*CacheData),
	}
}

func (c *Cache) ResetCache() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*CacheData)
}

func (c *Cache) WithLock(fn func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	fn()
}

func (c *Cache) WithRWLock(fn func()) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	fn()
}

func (c *Cache) GetUnsafe(k string) (*CacheData, bool) {
	cd, exists := c.data[k]

	if exists && cd.ExpiresAt != nil && cd.ExpiresAt.Before(time.Now()) {
		delete(c.data, k)
		return nil, false
	}

	return cd, exists
}

func (c *Cache) GetSafe(k string) (*CacheData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.GetUnsafe(k)
}

func (c *Cache) SetUnsafe(k string, v *CacheData) {
	c.data[k] = v
}

func (c *Cache) SetPartialUnsafe(k string, d CacheDataUpdate) {
	cd, exists := c.GetUnsafe(k)
	if !exists {
		return
	}

	if !exists {
		cd = &CacheData{}
		c.data[k] = cd
	}

	if d.Value != nil {
		cd.Value = *d.Value
	}
	if d.Requests != nil {
		cd.Requests = *d.Requests
	}
	if d.CreatedAt != nil {
		cd.CreatedAt = *d.CreatedAt
	}

	cd.ExpiresAt = d.ExpiresAt

	if d.Type != nil {
		cd.Type = *d.Type
	}
}

func (c *Cache) GetUnsafeData() map[string]*CacheData {
	return c.data
}

func (c *Cache) BGGC(interval time.Duration) {
	func() {
		t := time.NewTicker(interval)
		defer t.Stop()

		for range t.C {
			now := time.Now()

			c.WithLock(func() {
				for k, v := range c.data {
					if v.ExpiresAt != nil && now.After(*v.ExpiresAt) {
						delete(c.data, k)
					}
				}
			})
		}
	}()
}
