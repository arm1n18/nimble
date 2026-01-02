package commands

import (
	"cache/logger"
	"cache/storage"
	"strconv"
)

// Increase the number by 1
func INCR(c *storage.Cache, k string) {
	c.WithLock(func() {
		cachedData, exists := c.GetUnsafe(k)
		if !exists {
			logger.Error("Can`t find %v in memory", k)
			return
		}

		currentValue, ok := cachedData.Value.(float64)
		if !ok {
			logger.Error("Mismatch type")
			return
		}

		cachedData.Value = float64(currentValue) + 1
		cachedData.Requests++

		logger.Success("OK")
	})
}

// Decrease the number by 1
func DECR(c *storage.Cache, k string) {
	c.WithLock(func() {
		cachedData, exists := c.GetUnsafe(k)
		if !exists {
			logger.Error("Can`t find %v in memory", k)
			return
		}

		currentValue, ok := cachedData.Value.(float64)
		if !ok {
			logger.Error("Mismatch type")
			return
		}

		cachedData.Value = float64(currentValue) - 1
		cachedData.Requests++

		logger.Success("OK")
	})
}

// Increase the number by n
func INCRBY(c *storage.Cache, k, d string) {
	digit, err := strconv.ParseFloat(d, 64)
	if err != nil {
		logger.Error("Can`t parse number")
		return
	}

	c.WithLock(func() {
		cachedData, exists := c.GetUnsafe(k)
		if !exists {
			logger.Error("Can`t find %v in memory", k)
			return
		}

		currentValue, ok := cachedData.Value.(float64)
		if !ok {
			logger.Error("Mismatch type")
			return
		}

		cachedData.Value = float64(currentValue) + digit
		cachedData.Requests++

		logger.Success("OK")
	})
}

// Decrease the number by n
func DECRBY(c *storage.Cache, k, d string) {
	digit, err := strconv.ParseFloat(d, 64)
	if err != nil {
		logger.Error("Can`t parse number")
		return
	}

	c.WithLock(func() {
		cachedData, exists := c.GetUnsafe(k)
		if !exists {
			logger.Error("Can`t find %v in memory", k)
			return
		}

		currentValue, ok := cachedData.Value.(float64)
		if !ok {
			logger.Error("Mismatch type")
			return
		}

		cachedData.Value = float64(currentValue) - digit
		cachedData.Requests++

		logger.Success("OK")
	})
}

// Multiply the number by n
func MULL(c *storage.Cache, k, d string) {
	digit, err := strconv.ParseFloat(d, 64)
	if err != nil {
		logger.Error("Can`t parse number")
		return
	}

	c.WithLock(func() {
		cachedData, exists := c.GetUnsafe(k)
		if !exists {
			logger.Error("Can`t find %v in memory", k)
			return
		}

		currentValue, ok := cachedData.Value.(float64)
		if !ok {
			logger.Error("Mismatch type")
			return
		}

		cachedData.Value = float64(currentValue) * digit
		cachedData.Requests++

		logger.Success("OK")
	})
}

// Divide the number by n
func DIV(c *storage.Cache, k, d string) {
	digit, err := strconv.ParseFloat(d, 64)
	if err != nil {
		logger.Error("Can`t parse number")
		return
	}

	c.WithLock(func() {
		cachedData, exists := c.GetUnsafe(k)
		if !exists {
			logger.Error("Can`t find %v in memory", k)
			return
		}

		currentValue, ok := cachedData.Value.(float64)
		if !ok {
			logger.Error("Mismatch type")
			return
		}

		cachedData.Value = float64(currentValue) / digit
		cachedData.Requests++

		logger.Success("OK")
	})
}
