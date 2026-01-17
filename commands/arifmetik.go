package commands

import (
	"cache/logger"
	"cache/storage"
	"strconv"
)

func parseFloat(v any) (float64, bool) {
	f, err := strconv.ParseFloat(v.(string), 64)
	if err != nil {
		return -1, false
	}

	return f, true
}

func serializeFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// Increase the number by 1
func INCR(c *storage.Cache, k string) {
	c.WithLock(func() {
		cd, exists := c.GetUnsafe(k)
		if !exists {
			logger.Error("Can`t find %v in memory", k)
			return
		}

		cv, ok := parseFloat(cd.Value)
		if !ok {
			logger.Error("Mismatch type")
			return
		}

		cd.Value = serializeFloat(cv + 1)
		cd.Requests++

		logger.Success("OK")
	})
}

// Decrease the number by 1
func DECR(c *storage.Cache, k string) {
	c.WithLock(func() {
		cd, exists := c.GetUnsafe(k)
		if !exists {
			logger.Error("Can`t find %v in memory", k)
			return
		}

		cv, ok := parseFloat(cd.Value)
		if !ok {
			logger.Error("Mismatch type")
			return
		}

		cd.Value = serializeFloat(cv - 1)
		cd.Requests++

		logger.Success("OK")
	})
}

// Increase the number by n
func INCRBY(c *storage.Cache, k, v string) {
	f, ok := parseFloat(v)
	if !ok {
		logger.Error("Can`t parse number")
		return
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(k)
		if !exists {
			logger.Error("Can`t find %v in memory", k)
			return
		}

		cv, ok := parseFloat(cd.Value)
		if !ok {
			logger.Error("Mismatch type")
			return
		}

		cd.Value = serializeFloat(cv + f)
		cd.Requests++

		logger.Success("OK")
	})
}

// Decrease the number by n
func DECRBY(c *storage.Cache, k, v string) {
	f, ok := parseFloat(v)
	if !ok {
		logger.Error("Can`t parse number")
		return
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(k)
		if !exists {
			logger.Error("Can`t find %v in memory", k)
			return
		}

		cv, ok := parseFloat(cd.Value)
		if !ok {
			logger.Error("Mismatch type")
			return
		}

		cd.Value = serializeFloat(cv - f)
		cd.Requests++

		logger.Success("OK")
	})
}

// Multiply the number by n
func MULL(c *storage.Cache, k, v string) {
	f, ok := parseFloat(v)
	if !ok {
		logger.Error("Can`t parse number")
		return
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(k)
		if !exists {
			logger.Error("Can`t find %v in memory", k)
			return
		}

		cv, ok := parseFloat(cd.Value)
		if !ok {
			logger.Error("Mismatch type")
			return
		}

		cd.Value = serializeFloat(cv * f)
		cd.Requests++

		logger.Success("OK")
	})
}

// Divide the number by n
func DIV(c *storage.Cache, k, v string) {
	f, ok := parseFloat(v)
	if !ok {
		logger.Error("Can`t parse number")
		return
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(k)
		if !exists {
			logger.Error("Can`t find %v in memory", k)
			return
		}

		cv, ok := parseFloat(cd.Value)
		if !ok {
			logger.Error("Mismatch type")
			return
		}

		cd.Value = serializeFloat(cv / f)
		cd.Requests++

		logger.Success("OK")
	})
}
