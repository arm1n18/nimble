package commands

import (
	"strconv"
	"strings"

	"github.com/arm1n18/nimble/parser"
	"github.com/arm1n18/nimble/protocol"
	"github.com/arm1n18/nimble/storage"
)

/*
Delete values by key.

Description:

	Deletes values by the provided key and returns the number of deleted items.

Example:

  - Pattern: DEL KEY_1 KEY_2

  - res: 2

Notes:
  - Returns the number of keys that were deleted.
*/
func DEL(c *storage.Cache, args ...string) protocol.Response {
	var res protocol.Response
	var q int

	if len(args) == 0 {
		return protocol.Response{
			Success: false,
			Output:  protocol.ErrNotEnoughValues.Error(),
		}
	}

	if len(args) == 1 && args[0] == "*" {
		c.WithRWLock(func() {
			cd := c.GetUnsafeData()
			q = len(cd)
		})

		c.ResetCache()
		return protocol.Response{
			Success: true,
			Output:  protocol.Number(q),
		}
	}

	c.WithLock(func() {
		for _, k := range args {
			if _, exists := c.GetUnsafe(k); exists {
				delete(c.GetUnsafeData(), k)
				q++
			}
		}

		res = protocol.Response{
			Success: true,
			Output:  protocol.Number(q),
		}
	})

	return res
}

/*
Copy value from one key to another.

Description:

	Copies the value from source_key to target_key.

Example:
  - Pattern: COPY SOURCE_KEY TARGET_KEY

Notes:
  - If target_key already exists, it will be overwritten.
  - If target_key doesn't exist, it will be created.
  - Returns OK on success.
*/
func COPY(c *storage.Cache, f, t string) protocol.Response {
	var res protocol.Response

	if len(f) == 0 || len(t) == 0 || f == "*" || t == "*" {
		return protocol.Response{
			Success: false,
			Output:  protocol.ErrWrongKey.Error(),
		}
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(f)
		if !exists {
			res = protocol.Response{
				Success: false,
				Output:  protocol.Failure(),
			}
			return
		}

		c.SetPartialUnsafe(t, storage.CacheDataUpdate{Value: &cd.Value, Type: &cd.Type})

		res = protocol.Response{
			Success: true,
			Output:  protocol.Success(),
		}
	})

	return res
}

/*
List stored keys.

Description:

	Returns a list of all stored keys.

Example:

  - Pattern: LIST

  - res: [KEY_1, KEY_2]

Notes:
  - Returns an array of all keys that exist.
*/
func LIST(c *storage.Cache) protocol.Response {
	var res protocol.Response

	c.WithRWLock(func() {
		cd := c.GetUnsafeData()
		arr := make([]string, 0, len(cd))

		for k := range cd {
			arr = append(arr, k)
		}

		res = protocol.Response{
			Success: true,
			Output:  protocol.Array(serializeList(arr)),
		}
	})

	return res
}

/*
Show the number of stored keys.

Description:

	Returns a number of all stored keys.

Example:

  - Pattern: LISTLEN

  - res: 2

Notes:
  - Returns the number of all keys that exist.
*/
func LISTLEN(c *storage.Cache) protocol.Response {
	var res protocol.Response

	c.WithRWLock(func() {
		res = protocol.Response{
			Success: true,
			Output:  protocol.Number(len(c.GetUnsafeData())),
		}
	})

	return res
}

/*
Count how many keys exist in the cache.

Description:

	Checargs one or more keys in the cache and returns the total number of kesy that exist.

Example:

  - Pattern: EXISTS KEY_1 KEY_2 KEY_0

  - res: 2

  - Explanation: (KEY_1 exists, KEY_2 exists, KEY_0 does not exist)

Notes:
  - Returns the number of specified keys that exist.
*/
func EXISTS(c *storage.Cache, args ...string) protocol.Response {
	var res protocol.Response

	c.WithRWLock(func() {
		var q int
		cd := c.GetUnsafeData()

		for _, k := range args {
			if _, exists := cd[k]; exists {
				q++
			}
		}

		res = protocol.Response{
			Success: true,
			Output:  protocol.Number(q),
		}
	})

	return res
}

/*
Check if the keys exist and return array.

Description:

	Checks one or more keys in the cache and returns an array of integers.

Behavior:
  - For each key provided:
    1 if the key exists or
    0 if the key does not exist

Example:

  - Pattern: LEXISTS KEY_1 KEY_2 KEY_0

  - res: [1, 1, 0]

  - Explanation: (KEY_1 exists, KEY_2 exists, KEY_0 does not exist)

Notes:
  - Returns an array of 1s and 0s, where 1 indicates the key exists and 0 indicates it does not.
*/
func LEXISTS(c *storage.Cache, args ...string) protocol.Response {
	var res protocol.Response

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

		res = protocol.Response{
			Success: true,
			Output:  protocol.Array(serializeList(arr)),
		}
	})

	return res
}

/*
Get keys from cache by pattern.

1. Pattern '*'

  - Return all keys that start with prefix before '*'

  - Pattern: KEYS user:*

  - res: [user:1, user:123, user:ABC]

2. Pattern '?'

  - Return all keys that start with prefix before '?' and have a length equal to the number of '?' after prefix

  - Pattern: KEYS user:???

  - res: [user:123, user:256, user:ABC]

Notes:
  - Returns an array of keys matching the pattern.
*/
func KEYS(c *storage.Cache, args ...string) protocol.Response {
	var res protocol.Response

	if len(args) == 0 || len(args) == 2 || len(args) > 3 {
		return protocol.Response{
			Success: false,
			Output:  protocol.ErrInvalidSyntax.Error(),
		}
	}

	d := -1

	if len(args) > 2 {
		if strings.ToLower(args[1]) == "count" {
			var err error
			d, err = strconv.Atoi(args[2])
			if err != nil {
				return protocol.Response{
					Success: false,
					Output:  protocol.ErrNotANumber.Error(),
				}
			}
		} else {
			return protocol.Response{
				Success: false,
				Output:  protocol.ErrInvalidSyntax.Error(),
			}
		}
	}

	if parser.IsPatternCmd(args[0]) {
		c.WithRWLock(func() {
			res = protocol.Response{
				Success: true,
				Output:  protocol.Array(serializeList(getKeysByPattern(c.GetUnsafeData(), args[0], d))),
			}
		})
	}

	return res
}

/*
Rename a key.

Description:

	Renames source_key to target_key.

Example:
  - Pattern: RENAME SOURCE_KEY TARGET_KEY

Notes:
  - Returns 1 on success.
*/
func RENAME(c *storage.Cache, f, t string) protocol.Response {
	var res protocol.Response

	c.WithLock(func() {
		sCd, exists := c.GetUnsafe(f)
		if !exists {
			res = protocol.Response{
				Success: false,
				Output:  protocol.Failure(),
			}
			return
		}

		_, exists = c.GetUnsafe(t)
		if exists {
			res = protocol.Response{
				Success: false,
				Output:  protocol.Failure(),
			}
			return
		}

		d := c.GetUnsafeData()

		delete(d, f)
		c.SetUnsafe(t, sCd)

		res = protocol.Response{
			Success: true,
			Output:  protocol.Success(),
		}
	})

	return res
}

/*
Get the type of a key.

Description:

	Returns the type of data stored for the specified key.

Example:
  - Pattern: TYPE KEY

Notes:
  - Returns the data type as a string.
*/
func TYPE(c *storage.Cache, k string) protocol.Response {
	var res protocol.Response

	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(k)
		if !exists {
			res = protocol.Response{
				Success: false,
				Output:  protocol.ErrorMessage("Can`t find %v in memory", k),
			}
			return
		}

		res = protocol.Response{
			Success: true,
			Output:  string(cd.Type),
		}
	})

	return res
}
