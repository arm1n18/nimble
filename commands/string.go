package commands

import (
	"nimble/parser"
	"nimble/protocol"
	"nimble/storage"
	"path"
	"strconv"
	"time"
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
func SET(c *storage.Cache, k, v, t string) string {
	var result string

	pT, err := strconv.Atoi(t)
	if err != nil {
		return protocol.ErrInvalidTTL.Error()
	}

	if pT < -1 {
		return protocol.ErrInvalidTTL.Error()
	}

	var expiresAt *time.Time
	if pT != -1 {
		et := time.Now().Add(time.Duration(pT) * time.Second)
		expiresAt = &et
	}

	c.WithLock(func() {
		if k == "" || k == "*" {
			result = protocol.ErrWrongKey.Error()
			return
		}

		c.SetUnsafe(k, &storage.CacheData{
			Value:     v,
			Type:      storage.String,
			Requests:  1,
			CreatedAt: time.Now(),
			ExpiresAt: expiresAt,
		})

		result = protocol.Ok()
	})

	return result
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
func MSET(c *storage.Cache, args ...string) string {
	var result string

	if len(args)%2 == 0 && len(args) != 0 {
		for i := 0; i < len(args); i += 2 {
			k, _ := args[i], args[i+1]

			if k == "" || k == "*" {
				return protocol.ErrWrongKey.Error()
			}
		}

		c.WithLock(func() {

			for i := 0; i < len(args); i += 2 {
				k, v := args[i], args[i+1]

				c.SetUnsafe(k, &storage.CacheData{
					Value:     v,
					Type:      storage.String,
					Requests:  1,
					CreatedAt: time.Now(),
				})
			}

			result = protocol.Ok()
		})
	} else {
		result = protocol.ErrNotEnoughValues.Error()
		return result
	}

	return result
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
func GET(c *storage.Cache, k string) string {
	var result string

	c.WithRWLock(func() {
		if cd, exists := c.GetUnsafe(k); exists {
			cd.Requests++
			result = protocol.String(cd.Value)
			return
		} else {
			result = protocol.Nil()
			return
		}
	})

	return result
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
func MGET(c *storage.Cache, args ...string) string {
	var result string

	c.WithRWLock(func() {
		var arr []string

		for _, k := range args {
			if cd, exists := c.GetUnsafe(k); exists {
				cd.Requests++
				arr = append(arr, cd.Value)
			} else {
				arr = append(arr, protocol.Nil())
			}
		}

		result = protocol.Array(serializeList(arr))
	})

	return result
}
