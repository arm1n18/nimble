package commands

import (
	"nimble/parser"
	"nimble/protocol"
	"nimble/storage"
	"strconv"
	"strings"
)

// Remove any type of data from the cache
func DEL(c *storage.Cache, args ...string) string {
	var result string
	var q int

	if len(args) == 0 {
		return protocol.ErrNotEnoughValues.Error()
	}

	if len(args) == 1 && args[0] == "*" {
		c.WithRWLock(func() {
			cd := c.GetUnsafeData()
			q = len(cd)
		})

		c.ResetCache()
		return protocol.Number(q)
	}

	c.WithLock(func() {
		for _, k := range args {
			if _, exists := c.GetUnsafe(k); exists {
				delete(c.GetUnsafeData(), k)
				q++
			}
		}

		result = protocol.Number(q)
	})

	return result
}

// Copy data from one structure to another
func COPY(c *storage.Cache, f, t string) string {
	var result string

	if len(f) == 0 || len(t) == 0 || f == "*" || t == "*" {
		return protocol.ErrWrongKey.Error()
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(f)
		if !exists {
			// result = protocol.ErrorMessage("Can`t find %v in memory", f)
			result = protocol.Failure()
			return
		}

		c.SetPartialUnsafe(t, storage.CacheDataUpdate{Value: &cd.Value, Type: &cd.Type})

		result = protocol.Success()
	})

	return result
}

// Show all the keys
func LIST(c *storage.Cache) string {
	var result string

	c.WithRWLock(func() {
		cd := c.GetUnsafeData()
		arr := make([]string, 0, len(cd))

		for k := range cd {
			arr = append(arr, k)
		}

		result = protocol.Array(serializeList(arr))
	})

	return result
}

// Show the number of keys
func LISTLEN(c *storage.Cache) string {
	var result string

	c.WithRWLock(func() {
		result = protocol.Number(len(c.GetUnsafeData()))
	})

	return result
}

/*
Count how many keys exist in the cache

Description:
Checargs one or more keys in the cache and returns the total number of kesy that exist.

Example:

  - Pattern: EXISTS KEY_1 KEY_2 KEY_0

  - Result: 2

  - Explanation: (KEY_1 exists, KEY_2 exists, KEY_0 does not exist)
*/
func EXISTS(c *storage.Cache, args ...string) string {
	var result string

	c.WithRWLock(func() {
		var q int
		cd := c.GetUnsafeData()

		for _, k := range args {
			if _, exists := cd[k]; exists {
				q++
			}
		}

		result = protocol.Number(q)
	})

	return result
}

/*
Check if the keys exist and return array.

Description:
Checargs one or more keys in the cache and returns an array of integers.

Behavior:
  - For each key provided:
    1 if the key exists or
    0 if the key does not exist

Example:

  - Pattern: LEXISTS KEY_1 KEY_2 KEY_0

  - Result: [ 1, 1, 0 ]

  - Explanation: (KEY_1 exists, KEY_2 exists, KEY_0 does not exist)
*/
func LEXISTS(c *storage.Cache, args ...string) string {
	var result string

	c.WithRWLock(func() {
		arr := make([]string, len(args))
		cd := c.GetUnsafeData()

		for i, k := range args {
			if _, exists := cd[k]; exists {
				arr[i] = "1"
			} else {
				arr[i] = "0"
			}
		}

		result = protocol.Array(serializeList(arr))
	})

	return result
}

/*
Get keys from cache by pattern

1. Pattern '*'

  - Return all keys that start with prefix before '*'

  - Pattern: KEYS user:*

  - Result: [ user:1, user:123, user:ABC ]

2. Pattern '?'

  - Return all keys that start with prefix before '?' and have a length equal to the number of '?' after prefix

  - Pattern: KEYS user:???

  - Result: [ user:123, user:256, user:ABC ]
*/
func KEYS(c *storage.Cache, args ...string) string {
	var result string

	if len(args) == 0 || len(args) == 2 || len(args) > 3 {
		return protocol.ErrInvalidSyntax.Error()
	}

	d := -1

	if len(args) > 2 {
		if strings.ToLower(args[1]) == "count" {
			var err error
			d, err = strconv.Atoi(args[2])
			if err != nil {
				return protocol.ErrNotANumber.Error()
			}
		} else {
			return protocol.ErrInvalidSyntax.Error()
		}
	}

	if parser.IsPatternCmd(args[0]) {
		c.WithRWLock(func() {
			result = protocol.Array(serializeList(getKeysByPattern(c.GetUnsafeData(), args[0], d)))
		})
	}

	return result
}

func RENAME(c *storage.Cache, f, t string) string {
	var result string

	c.WithLock(func() {
		sCd, exists := c.GetUnsafe(f)
		if !exists {
			// result = protocol.ErrorMessage("can`t find %v in memory", f)
			result = protocol.Failure()
			return
		}

		_, exists = c.GetUnsafe(t)
		if exists {
			// result = protocol.ErrorMessage("%v already exists", t)
			result = protocol.Failure()
			return
		}

		d := c.GetUnsafeData()

		delete(d, f)
		c.SetUnsafe(t, sCd)

		result = protocol.Success()
	})

	return result
}

// Show the type of data stored in the cache
func TYPE(c *storage.Cache, k string) string {
	var result string

	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(k)
		if !exists {
			result = protocol.ErrorMessage("Can`t find %v in memory", k)
			return
		}

		result = string(cd.Type)
	})

	return result
}
