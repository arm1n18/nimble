package commands

import (
	"encoding/json"
	"time"

	"github.com/arm1n18/nimble/protocol"
	"github.com/arm1n18/nimble/storage"
)

func parseHash(s string) (map[string]string, bool) {
	var m map[string]string
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return nil, false
	}
	if m == nil {
		m = make(map[string]string)
	}

	return m, true
}

func serializeHash(m map[string]string) string {
	b, _ := json.Marshal(m)
	return string(b)
}

/*
Store a key-value pair in a hash.

Description:

	Adds or updates a key-value pair in the hash.

Example:
  - Pattern: HSET HASH_NAME KEY VALUE

Notes:
  - Creates the hash if it doesn't exist.
  - Returns OK on success.
*/
func HSET(c *storage.Cache, h string, args ...string) protocol.Response {
	var res string

	if len(args) == 0 || len(args)%2 != 0 {
		return protocol.Response{
			Success: false,
			Output:  protocol.ErrNotEnoughValues.Error(),
		}
	}

	c.WithLock(func() {
		hash := make(map[string]string)

		if cd, exists := c.GetUnsafe(h); exists {
			cd.Requests++

			if m, ok := parseHash(cd.Value); ok {
				for i := 0; i < len(args); i += 2 {
					m[args[i]] = args[i+1]
				}

				c.SetUnsafe(h, &storage.CacheData{
					Value:    serializeHash(m),
					Type:     storage.Hash,
					Requests: cd.Requests + 1,
				})

				res = protocol.Ok()
				return
			}
		}

		for i := 0; i < len(args); i += 2 {
			hash[args[i]] = args[i+1]
		}

		c.SetUnsafe(h, &storage.CacheData{
			Value:     serializeHash(hash),
			Type:      storage.Hash,
			Requests:  1,
			CreatedAt: time.Now(),
		})

		res = protocol.Ok()
	})

	return protocol.Response{
		Success: true,
		Output:  res,
	}
}

/*
Get a key value from a hash.

Description:

	Returns the value associated with the specified key in the hash.

Example:
  - Pattern: HGET HASH_NAME KEY_1 KEY_2

Notes:
  - Returns an array of values stored by key.
*/
func HGET(c *storage.Cache, h string, ks ...string) protocol.Response {
	var res protocol.Response

	c.WithRWLock(func() {
		arr := make([]string, len(ks))

		cd, exists := c.GetUnsafe(h)
		if !exists {
			res = protocol.Response{
				Success: true,
				Output:  protocol.Array("[]"),
			}
			return
		}

		cd.Requests++

		m, ok := parseHash(cd.Value)
		if !ok {
			res = protocol.Response{
				Success: false,
				Output:  protocol.ErrMismatchType.Error(),
			}
			return
		}

		for i, k := range ks {
			if v, exists := m[k]; exists {
				arr[i] = v
			} else {
				arr[i] = protocol.Nil()
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
Get all key names from a hash.

Description:

	Returns a list of all keys stored in the hash.

Example:
  - Pattern: HKEYS HASH_NAME

Notes:
  - Returns an array of all keys.
*/
func HKEYS(c *storage.Cache, h string) protocol.Response {
	var res protocol.Response

	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(h)
		if !exists {
			res = protocol.Response{
				Success: true,
				Output:  protocol.Array("[]"),
			}
			return
		}

		cd.Requests++

		m, ok := parseHash(cd.Value)
		if !ok {
			res = protocol.Response{
				Success: false,
				Output:  protocol.ErrMismatchType.Error(),
			}
			return
		}

		s := make([]string, 0, len(m))
		for k := range m {
			s = append(s, k)
		}

		res = protocol.Response{
			Success: true,
			Output:  protocol.Array(serializeList(s)),
		}
	})

	return res
}

/*
Get all values from a hash.

Description:

	Returns a list of all values stored in the hash.

Example:
  - Pattern: HVALUES HASH_NAME

Notes:
  - Returns an array of all values.
*/
func HVALUES(c *storage.Cache, h string) protocol.Response {
	var res protocol.Response

	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(h)
		if !exists {
			res = protocol.Response{
				Success: true,
				Output:  protocol.Array("[]"),
			}
			return
		}

		cd.Requests++

		m, ok := parseHash(cd.Value)
		if !ok {
			res = protocol.Response{
				Success: false,
				Output:  protocol.ErrMismatchType.Error(),
			}
			return
		}

		s := make([]string, 0, len(m))
		for _, v := range m {
			s = append(s, v)
		}

		res = protocol.Response{
			Success: true,
			Output:  protocol.Array(serializeList(s)),
		}
	})

	return res
}

/*
Delete one or more keys from a hash.

Description:

	Removes the specified keys from the hash.

Example:
  - Pattern: HDEL HASH_NAME KEY_1 KEY_2

Notes:
  - Returns the number of deleted keys.
*/
func HDEL(c *storage.Cache, h string, args ...string) protocol.Response {
	var res protocol.Response

	if len(args) == 0 {
		return protocol.Response{
			Success: false,
			Output:  protocol.ErrNotEnoughValues.Error(),
		}
	}

	c.WithLock(func() {
		var q int

		cd, exists := c.GetUnsafe(h)
		if !exists {
			res = protocol.Response{
				Success: true,
				Output:  protocol.Number(-1),
			}
			return
		}

		cd.Requests++

		m, ok := parseHash(cd.Value)
		if !ok {
			res = protocol.Response{
				Success: false,
				Output:  protocol.ErrMismatchType.Error(),
			}
			return
		}

		for _, k := range args {
			if _, exists := m[k]; exists {
				delete(m, k)
				q++
			}
		}

		c.SetUnsafe(h, &storage.CacheData{
			Value:    serializeHash(m),
			Requests: cd.Requests + 1,
		})

		res = protocol.Response{
			Success: true,
			Output:  protocol.Number(q),
		}
	})

	return res
}

/*
Check if keys exist in a hash.

Description:

	Checks if the specified keys exist in the hash.

Example:
  - Pattern: HCONTAINS HASH_NAME KEY_1 KEY_2

Notes:
  - Returns the number of specified keys that exist in the hash.
*/
func HCONTAINS(c *storage.Cache, h string, args ...string) protocol.Response {
	var res protocol.Response

	c.WithRWLock(func() {
		var q int

		cd, exists := c.GetUnsafe(h)
		if !exists {
			res = protocol.Response{
				Success: false,
				Output:  protocol.Number(-1),
			}
			return
		}

		cd.Requests++

		m, ok := parseHash(cd.Value)
		if !ok {
			res = protocol.Response{
				Success: false,
				Output:  protocol.ErrMismatchType.Error(),
			}
			return
		}

		for _, k := range args {
			if _, exists := m[k]; exists {
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
Check if keys exist in a hash.

Description:

	Checks if the specified keys exist in the hash.

Example:

  - Pattern: LHCONTAINS HASH_NAME KEY_1 KEY_2

  - res: [1, 0]

Notes:
  - Returns an array of 1s and 0s, where 1 indicates the key exists and 0 indicates it does not.
*/
func LHCONTAINS(c *storage.Cache, h string, args ...string) protocol.Response {
	var res protocol.Response

	c.WithRWLock(func() {
		arr := make([]string, len(args))

		cd, exists := c.GetUnsafe(h)
		if !exists {
			res = protocol.Response{
				Success: true,
				Output:  protocol.Array("[]"),
			}
			return
		}

		cd.Requests++

		m, ok := parseHash(cd.Value)
		if !ok {
			res = protocol.Response{
				Success: false,
				Output:  protocol.ErrMismatchType.Error(),
			}
			return
		}

		for i, k := range args {
			if _, exists := m[k]; exists {
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
Show the number of stored keys in hash.

Description:

	Returns a number of all stored keys in hash.

Example:

  - Pattern: HLEN HASH_NAME

  - res: 2

Notes:
  - Returns the number of all keys that exist in the hash.
*/
func HLEN(c *storage.Cache, h string) protocol.Response {
	var res protocol.Response

	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(h)
		if !exists {
			res = protocol.Response{
				Success: true,
				Output:  protocol.Number(-1),
			}
			return
		}

		cd.Requests++

		m, ok := parseHash(cd.Value)
		if !ok {
			res = protocol.Response{
				Success: false,
				Output:  protocol.ErrMismatchType.Error(),
			}
			return
		}

		res = protocol.Response{
			Success: true,
			Output:  protocol.Number(len(m)),
		}
	})

	return res
}
