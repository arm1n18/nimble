package commands

import (
	"fmt"
	"nimble/protocol"
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

/*
Increase the number by 1.

Description:

	Increments the integer value by 1.

Example:

  - Pattern: INCR "KEY"

  - Result: 1

Notes:
  - Returns the new value after incrementing.
  - If the key does not exist, it is set to 0 before incrementing.
*/
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
			result = protocol.Number(1)
			return
		}

		cv, ok := parseFloat(cd.Value)
		if !ok {
			result = protocol.ErrNotANumber.Error()
			return
		}

		calc := cv + 1
		cd.Value = serializeFloat(calc)
		cd.Requests++

		result = protocol.Number(calc)
	})

	return result
}

/*
Decrease the number by 1.

Description:

	Decrements the integer value by 1.

Example:

  - Pattern: DECR "KEY"

  - Result: -1

Notes:
  - Returns the new value after decrementing.
  - If the key does not exist, it is set to 0 before decrementing.
*/
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
			result = protocol.Number(-1)
			return
		}

		cv, ok := parseFloat(cd.Value)
		if !ok {
			result = protocol.ErrNotANumber.Error()
			return
		}

		calc := cv - 1
		cd.Value = serializeFloat(calc)
		cd.Requests++

		result = protocol.Number(calc)
	})

	return result
}

/*
Increase the number by N.

Description:

	Increments the integer value by N.

Example:

  - Pattern: INCRBY "KEY" 10

  - Result: 10

Notes:
  - Returns the new value after incrementing.
  - If the key does not exist, it is set to 0 before incrementing.
*/
func INCRBY(c *storage.Cache, k, v string) string {
	var result string

	f, ok := parseFloat(v)
	if !ok {
		return protocol.ErrMismatchType.Error()
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
			result = protocol.Number(f)
			return
		}

		cv, ok := parseFloat(cd.Value)
		if !ok {
			result = protocol.ErrNotANumber.Error()
			return
		}

		calc := cv + f
		cd.Value = serializeFloat(calc)
		cd.Requests++

		result = protocol.Number(calc)
	})

	return result
}

/*
Decrease the number by N.

Description:

	Decrements the integer value by N.

Example:

  - Pattern: DECRBY "KEY" 10

  - Result: -10

Notes:
  - Returns the new value after decrementing.
  - If the key does not exist, it is set to 0 before decrementing.
*/
func DECRBY(c *storage.Cache, k, v string) string {
	var result string

	f, ok := parseFloat(v)
	if !ok {
		return protocol.ErrMismatchType.Error()
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
			result = protocol.Number(f * -1)
			return
		}

		cv, ok := parseFloat(cd.Value)
		if !ok {
			result = protocol.ErrNotANumber.Error()
			return
		}

		calc := cv - f
		cd.Value = serializeFloat(calc)
		cd.Requests++

		result = protocol.Number(calc)
	})

	return result
}

/*
Multiply the number by N.

Description:

	Multiplies the integer value by N.

Example:

  - Pattern: MUL "KEY" 2

  - Result: 10

Notes:
  - Returns the new value after multiplication.
  - If the key does not exist, it is set to 0.
*/
func MUL(c *storage.Cache, k, v string) string {
	var result string

	f, ok := parseFloat(v)
	if !ok {
		return protocol.ErrMismatchType.Error()
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
			result = protocol.Number(0)
			return
		}

		cv, ok := parseFloat(cd.Value)
		if !ok {
			result = protocol.ErrNotANumber.Error()
			return
		}

		calc := cv * f
		cd.Value = serializeFloat(calc)
		cd.Requests++

		result = protocol.Number(calc)
	})

	return result
}

/*
Divide the number by N.

Description:

	Divides the integer value by N.

Example:

  - Pattern: DIV "KEY" 2

  - Result: 5

Notes:
  - Returns the new value after division.
  - If the key does not exist, it is set to 0.
*/
func DIV(c *storage.Cache, k, v string) string {
	var result string

	f, ok := parseFloat(v)
	if !ok {
		return protocol.ErrMismatchType.Error()
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
			result = protocol.Number(0)
			return
		}

		cv, ok := parseFloat(cd.Value)
		if !ok {
			result = protocol.ErrNotANumber.Error()
			return
		}

		calc := cv / f
		cd.Value = serializeFloat(calc)
		cd.Requests++

		result = protocol.Number(calc)
	})

	return result
}
