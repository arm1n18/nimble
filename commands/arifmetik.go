package commands

import (
	"fmt"
	"nimble/formatter"
	"nimble/storage"
	"strconv"
	"time"
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
func INCR(c *storage.Cache, k string) string {
	var result string

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(k)
		if !exists {
			c.SetUnsafe(k, &storage.CacheData{
				Value:     "1",
				Requests:  1,
				Type:      storage.String,
				CreatedAt: time.Now(),
			})
			result = formatter.Number(1)
			return
		}

		cv, ok := parseFloat(cd.Value)
		if !ok {
			result = formatter.ErrNotANumber.Error()
			return
		}

		calc := cv + 1
		cd.Value = serializeFloat(calc)
		cd.Requests++

		result = formatter.Number(calc)
	})

	return result
}

// Decrease the number by 1
func DECR(c *storage.Cache, k string) string {
	var result string

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(k)
		if !exists {
			c.SetUnsafe(k, &storage.CacheData{
				Value:     "-1",
				Requests:  1,
				Type:      storage.String,
				CreatedAt: time.Now(),
			})
			result = formatter.Number(-1)
			return
		}

		cv, ok := parseFloat(cd.Value)
		if !ok {
			result = formatter.ErrNotANumber.Error()
			return
		}

		calc := cv - 1
		cd.Value = serializeFloat(calc)
		cd.Requests++

		result = formatter.Number(calc)
	})

	return result
}

// Increase the number by n
func INCRBY(c *storage.Cache, k, v string) string {
	var result string

	f, ok := parseFloat(v)
	if !ok {
		return formatter.ErrMismatchType.Error()
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(k)
		if !exists {
			c.SetUnsafe(k, &storage.CacheData{
				Value:     fmt.Sprint(f),
				Requests:  1,
				Type:      storage.String,
				CreatedAt: time.Now(),
			})
			result = formatter.Number(f)
			return
		}

		cv, ok := parseFloat(cd.Value)
		if !ok {
			result = formatter.ErrNotANumber.Error()
			return
		}

		calc := cv + f
		cd.Value = serializeFloat(calc)
		cd.Requests++

		result = formatter.Number(calc)
	})

	return result
}

// Decrease the number by n
func DECRBY(c *storage.Cache, k, v string) string {
	var result string

	f, ok := parseFloat(v)
	if !ok {
		return formatter.ErrMismatchType.Error()
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(k)
		if !exists {
			c.SetUnsafe(k, &storage.CacheData{
				Value:     fmt.Sprint(f * -1),
				Requests:  1,
				Type:      storage.String,
				CreatedAt: time.Now(),
			})
			result = formatter.Number(f * -1)
			return
		}

		cv, ok := parseFloat(cd.Value)
		if !ok {
			result = formatter.ErrNotANumber.Error()
			return
		}

		calc := cv - f
		cd.Value = serializeFloat(calc)
		cd.Requests++

		result = formatter.Number(calc)
	})

	return result
}

// Multiply the number by n
func MULL(c *storage.Cache, k, v string) string {
	var result string

	f, ok := parseFloat(v)
	if !ok {
		return formatter.ErrMismatchType.Error()
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(k)
		if !exists {
			c.SetUnsafe(k, &storage.CacheData{
				Value:     "0",
				Requests:  1,
				Type:      storage.String,
				CreatedAt: time.Now(),
			})
			result = formatter.Number(0)
			return
		}

		cv, ok := parseFloat(cd.Value)
		if !ok {
			result = formatter.ErrNotANumber.Error()
			return
		}

		calc := cv * f
		cd.Value = serializeFloat(calc)
		cd.Requests++

		result = formatter.Number(calc)
	})

	return result
}

// Divide the number by n
func DIV(c *storage.Cache, k, v string) string {
	var result string

	f, ok := parseFloat(v)
	if !ok {
		return formatter.ErrMismatchType.Error()
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(k)
		if !exists {
			c.SetUnsafe(k, &storage.CacheData{
				Value:     "0",
				Requests:  1,
				Type:      storage.String,
				CreatedAt: time.Now(),
			})
			result = formatter.Number(0)
			return
		}

		cv, ok := parseFloat(cd.Value)
		if !ok {
			result = formatter.ErrNotANumber.Error()
			return
		}

		calc := cv / f
		cd.Value = serializeFloat(calc)
		cd.Requests++

		result = formatter.Number(calc)
	})

	return result
}
