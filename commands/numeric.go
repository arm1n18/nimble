package commands

import (
	"fmt"
	"strconv"
	"time"

	"github.com/arm1n18/nimble/protocol"
	"github.com/arm1n18/nimble/storage"
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
func INCR(c *storage.Cache, k string) protocol.Response {
	var res protocol.Response

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(k)
		if !exists {
			c.SetUnsafe(k, &storage.CacheData{
				Value:     "1",
				Requests:  1,
				Type:      storage.String,
				CreatedAt: time.Now(),
			})
			res = protocol.Response{
				Success: true,
				Output:  protocol.Number(1),
			}
			return
		}

		cv, ok := parseFloat(cd.Value)
		if !ok {
			res = protocol.Response{
				Success: false,
				Output:  protocol.ErrNotANumber.Error(),
			}
			return
		}

		calc := cv + 1
		cd.Value = serializeFloat(calc)
		cd.Requests++

		res = protocol.Response{
			Success: true,
			Output:  protocol.Number(calc),
		}
	})

	return res
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
func DECR(c *storage.Cache, k string) protocol.Response {
	var res protocol.Response

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(k)
		if !exists {
			c.SetUnsafe(k, &storage.CacheData{
				Value:     "-1",
				Requests:  1,
				Type:      storage.String,
				CreatedAt: time.Now(),
			})
			res = protocol.Response{
				Success: true,
				Output:  protocol.Number(-1),
			}
			return
		}

		cv, ok := parseFloat(cd.Value)
		if !ok {
			res = protocol.Response{
				Success: false,
				Output:  protocol.ErrNotANumber.Error(),
			}
			return
		}

		calc := cv - 1
		cd.Value = serializeFloat(calc)
		cd.Requests++

		res = protocol.Response{
			Success: true,
			Output:  protocol.Number(calc),
		}
	})

	return res
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
func INCRBY(c *storage.Cache, k, v string) protocol.Response {
	var res protocol.Response

	f, ok := parseFloat(v)
	if !ok {
		res = protocol.Response{
			Success: false,
			Output:  protocol.ErrMismatchType.Error(),
		}
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
			res = protocol.Response{
				Success: true,
				Output:  protocol.Number(f),
			}
			return
		}

		cv, ok := parseFloat(cd.Value)
		if !ok {
			res = protocol.Response{
				Success: false,
				Output:  protocol.ErrNotANumber.Error(),
			}
			return
		}

		calc := cv + f
		cd.Value = serializeFloat(calc)
		cd.Requests++

		res = protocol.Response{
			Success: true,
			Output:  protocol.Number(calc),
		}
	})

	return res
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
func DECRBY(c *storage.Cache, k, v string) protocol.Response {
	var res protocol.Response

	f, ok := parseFloat(v)
	if !ok {
		res = protocol.Response{
			Success: false,
			Output:  protocol.ErrMismatchType.Error(),
		}
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
			res = protocol.Response{
				Success: true,
				Output:  protocol.Number(f * -1),
			}
			return
		}

		cv, ok := parseFloat(cd.Value)
		if !ok {
			res = protocol.Response{
				Success: false,
				Output:  protocol.ErrNotANumber.Error(),
			}
			return
		}

		calc := cv - f
		cd.Value = serializeFloat(calc)
		cd.Requests++

		res = protocol.Response{
			Success: true,
			Output:  protocol.Number(calc),
		}
	})

	return res
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
func MUL(c *storage.Cache, k, v string) protocol.Response {
	var res protocol.Response

	f, ok := parseFloat(v)
	if !ok {
		res = protocol.Response{
			Success: false,
			Output:  protocol.ErrMismatchType.Error(),
		}
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
			res = protocol.Response{
				Success: true,
				Output:  protocol.Number(0),
			}
			return
		}

		cv, ok := parseFloat(cd.Value)
		if !ok {
			res = protocol.Response{
				Success: false,
				Output:  protocol.ErrNotANumber.Error(),
			}
			return
		}

		calc := cv * f
		cd.Value = serializeFloat(calc)
		cd.Requests++

		res = protocol.Response{
			Success: true,
			Output:  protocol.Number(calc),
		}
	})

	return res
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
func DIV(c *storage.Cache, k, v string) protocol.Response {
	var res protocol.Response

	f, ok := parseFloat(v)
	if !ok {
		res = protocol.Response{
			Success: false,
			Output:  protocol.ErrMismatchType.Error(),
		}
	}

	if f == 0 {
		return protocol.Response{
			Success: false,
			Output:  protocol.ErrorMessage("Division by zero is not allowed"),
		}
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
			res = protocol.Response{
				Success: true,
				Output:  protocol.Number(0),
			}
			return
		}

		cv, ok := parseFloat(cd.Value)
		if !ok {
			res = protocol.Response{
				Success: false,
				Output:  protocol.ErrNotANumber.Error(),
			}
			return
		}

		calc := cv / f
		cd.Value = serializeFloat(calc)
		cd.Requests++

		res = protocol.Response{
			Success: true,
			Output:  protocol.Number(calc),
		}
	})

	return res
}
