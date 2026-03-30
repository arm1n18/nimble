package commands

import (
	"path"
	"strconv"
	"time"

	"github.com/arm1n18/nimble/parser"
	"github.com/arm1n18/nimble/protocol"
	"github.com/arm1n18/nimble/storage"
)

func getKeysByPattern(m map[string]*storage.CacheData, pattern string, lim int) []string {
	var args []string

	symbol, ok := parser.GetPatternSymbol(pattern)
	if !ok {
		return args
	}

	switch symbol {
	case "*":
		{
			for k := range m {
				match, err := path.Match(pattern, k)
				if err != nil {
					continue
				}

				if match {
					args = append(args, k)
					if lim > 0 && len(args) >= lim {
						break
					}
				}
			}
		}
	case "?":
		{
			for k := range m {
				match, err := path.Match(pattern, k)
				if err != nil {
					continue
				}

				if match && len(pattern) == len(k) {
					args = append(args, k)
					if lim > 0 && len(args) >= lim {
						break
					}
				}
			}
		}
	}

	return args
}

/*
Set the value of a key with optional expiration.

Description:

	Sets the value of KEY to VALUE with optional expiration in seconds.

Example:
  - Pattern: SET KEY "VALUE" 60

Notes:
  - Returns OK on success.
*/
func SET(c *storage.Cache, k, v, t string) protocol.Response {
	pT, err := strconv.Atoi(t)
	if err != nil || pT < -1 {
		return protocol.Response{
			Success: false,
			Output:  protocol.ErrInvalidTTL.Error(),
		}
	}

	if k == "" || k == "*" {
		return protocol.Response{
			Success: false,
			Output:  protocol.ErrWrongKey.Error(),
		}
	}

	var expiresAt *time.Time
	if pT != -1 {
		et := time.Now().Add(time.Duration(pT) * time.Second)
		expiresAt = &et
	}

	c.WithLock(func() {
		c.SetUnsafe(k, &storage.CacheData{
			Value:     v,
			Type:      storage.String,
			Requests:  1,
			CreatedAt: time.Now(),
			ExpiresAt: expiresAt,
		})
	})

	return protocol.Response{
		Success: true,
		Output:  protocol.Ok(),
	}
}

/*
Set the values of a key with optional expiration.

Description:

	Sets the values of specified KEYS to their VALUES.
	All keys are set WITHOUT expiration.

Example:
  - Pattern: MSET KEY_1 "VALUE_1" KEY_2 "VALUE_2"

Notes:
  - By default MSET doesn`t support TTK.
    Use SET with expiration options or apply TTL separately after MSET.
  - Returns OK on success.
*/
func MSET(c *storage.Cache, args ...string) protocol.Response {
	if len(args) == 0 || len(args)%2 != 0 {
		return protocol.Response{
			Success: false,
			Output:  protocol.ErrNotEnoughValues.Error(),
		}
	}

	for i := 0; i < len(args); i += 2 {
		k := args[i]

		if k == "" || k == "*" {
			return protocol.Response{
				Success: false,
				Output:  protocol.ErrWrongKey.Error(),
			}
		}
	}

	c.WithLock(func() {
		now := time.Now()

		for i := 0; i < len(args); i += 2 {
			k, v := args[i], args[i+1]

			c.SetUnsafe(k, &storage.CacheData{
				Value:     v,
				Type:      storage.String,
				Requests:  1,
				CreatedAt: now,
			})
		}
	})

	return protocol.Response{
		Success: true,
		Output:  protocol.Ok(),
	}
}

/*
Get the values of a key.

Description:

	Returns the value associated with KEY.

Example:
  - Pattern: GET KEY

Notes:
  - Returns the value of the key as a string, or null if the key does not exist.
*/
func GET(c *storage.Cache, k string) protocol.Response {
	var res protocol.Response

	c.WithRWLock(func() {
		if cd, exists := c.GetUnsafe(k); exists {
			cd.Requests++
			res = protocol.Response{
				Success: true,
				Output:  protocol.String(cd.Value),
			}
			return
		}

		res = protocol.Response{
			Success: true,
			Output:  protocol.Nil(),
		}
	})

	return res
}

/*
Get the values of multiple keys.

Description:

	Returns the values associated with the specified KEYS.

Example:
  - Pattern: MGET KEY_1 KEY_2

Notes:
  - Returns an array of values or nulls.
*/
func MGET(c *storage.Cache, args ...string) protocol.Response {
	var res protocol.Response

	c.WithRWLock(func() {
		arr := make([]string, 0, len(args))

		for _, k := range args {
			if cd, exists := c.GetUnsafe(k); exists {
				cd.Requests++
				arr = append(arr, cd.Value)
			} else {
				arr = append(arr, protocol.Nil())
			}
		}

		res = protocol.Response{
			Success: true,
			Output:  protocol.Array(serializeList(arr)),
		}
	})

	return res
}
