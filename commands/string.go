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

// Store data of string type in the cache
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

// Store data of string type in the cache
func MSET(c *storage.Cache, args ...string) string {
	var result string

	if len(args)%2 == 0 && len(args) != 0 {
		// if ok := removeQuotes(&s, 1, 1); !ok {
		// 	return
		// }

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

// Get string type of data from the cache
func GET(c *storage.Cache, k string) string {
	var result string

	c.WithRWLock(func() {
		if cd, exists := c.GetUnsafe(k); exists {
			cd.Requests++

			// m, ok := cd.Value.(map[string]struct{})
			// if ok {
			// 	tS := make([]string, 0, len(m))

			// 	for k := range m {
			// 		tS = append(tS, k)
			// 	}

			// 	cV = tS
			// } else {
			// 	cV = cd.Value
			// }

			result = protocol.String(cd.Value)
			return
		} else {
			result = protocol.Nil()
			return
		}
	})

	return result
}

// Get string type of data from the cache
func MGET(c *storage.Cache, args ...string) string {
	var result string

	c.WithRWLock(func() {
		var arr []string

		for _, k := range args {
			if cd, exists := c.GetUnsafe(k); exists {
				cd.Requests++

				// m, ok := cd.Value.(map[string]string{})
				// if ok {
				// 	tS := make([]string, 0, len(m))

				// 	for k := range m {
				// 		tS = append(tS, k)
				// 	}

				// 	cV = tS
				// } else {
				// 	cV = cd.Value
				// }

				arr = append(arr, cd.Value)
			} else {
				arr = append(arr, protocol.Nil())
			}
		}

		result = protocol.Array(serializeList(arr))
	})

	return result
}
