package storage

import (
	"log"
	"sync"
	"time"
)

// Defined memory usage limit
var MAX_MEMO int

type Mode string

const (
	ReadOnly  Mode = "read-only"
	ReadWrite Mode = "read-write"
)

type CacheDataUpdate struct {
	Value     interface{}
	Requests  *int
	TimeStamp *time.Time
	TTK       *time.Timer
	TTL       *time.Time
}

type CacheData struct {
	Value     interface{}
	Requests  int
	TimeStamp time.Time
	TTK       *time.Timer
	TTL       *time.Time
}

type Cache struct {
	data map[string]*CacheData
	mode Mode
	mu   sync.RWMutex
}

func CreateCache() *Cache {
	return &Cache{
		data: make(map[string]*CacheData),
		mode: ReadWrite,
	}
}

func (c *Cache) ResetCache() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*CacheData)
	c.mode = ReadWrite
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
	cacheData, exists := c.data[k]
	return cacheData, exists
}

func (c *Cache) GetSafe(k string) (*CacheData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.GetUnsafe(k)
}

func (c *Cache) GetData() map[string]*CacheData {
	return c.data
}

func (c *Cache) GetMode() Mode {
	return c.mode
}

func (c *Cache) SetUnsafe(k string, v *CacheData) {
	c.data[k] = v
}

func (c *Cache) SetPartialUnsafe(k string, nD CacheDataUpdate) {
	cd, exists := c.GetUnsafe(k)
	if !exists {
		log.Printf("Can`t find %v in memory", k)
		return
	}

	if !exists {
		cd = &CacheData{}
		c.data[k] = cd
	}

	if nD.Value != nil {
		cd.Value = nD.Value
	}
	if nD.Requests != nil {
		cd.Requests = *nD.Requests
	}
	if nD.TimeStamp != nil {
		cd.TimeStamp = *nD.TimeStamp
	}
	if nD.TTK != nil {
		cd.TTK = nD.TTK
	}
	if nD.TTL != nil {
		cd.TTL = nD.TTL
	}
}

func (c *Cache) SetMode(m Mode) {
	c.mode = m
}
