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

type History struct {
	Command   string
	Timestamp time.Time
}

type Cache struct {
	data map[string]*CacheData
	mu   sync.RWMutex

	maxHistory   int
	history      []History
	historyIndex map[string][]int
}

func CreateCache(mH int) *Cache {
	return &Cache{
		data:         make(map[string]*CacheData),
		history:      make([]History, 0),
		historyIndex: make(map[string][]int),
		maxHistory:   mH,
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

func (c *Cache) AddToHistory(k, cmd string) {
	c.WithLock(func() {
		c.history = append(c.history, History{
			Command:   cmd,
			Timestamp: time.Now(),
		})

		c.historyIndex[k] = append(c.historyIndex[k], len(c.history)-1)

		if len(c.history) > c.maxHistory {
			dif := len(c.history) - c.maxHistory
			c.history = c.history[dif:]

			for k, indexes := range c.historyIndex {
				newIndexes := make([]int, 0, len(indexes))
				for _, i := range indexes {
					i -= dif
					if i >= 0 {
						newIndexes = append(newIndexes, i)
					}
				}

				if len(newIndexes) == 0 {
					delete(c.historyIndex, k)
				} else {
					c.historyIndex[k] = newIndexes
				}
			}
		}
	})
}

func (c *Cache) GetKeyHistory(k string) []History {
	var res []History

	c.WithRWLock(func() {
		indexes, exists := c.historyIndex[k]
		if !exists || len(indexes) == 0 {
			res = nil
			return
		}

		history := make([]History, len(indexes))
		for i, index := range indexes {
			history[i] = c.history[index]
		}

		res = history
	})

	return res
}

func (c *Cache) GetHistory() []History {
	var res []History

	c.WithRWLock(func() {
		for _, h := range c.history {
			res = append(res, h)
		}
	})

	return res
}

func (c *Cache) GetHistoryUntil(t time.Time) []History {
	var res []History

	c.WithRWLock(func() {
		for _, h := range c.history {
			if !h.Timestamp.After(t) {
				res = append(res, h)
			} else {
				break
			}
		}
	})

	return res
}
