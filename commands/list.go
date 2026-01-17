package commands

import (
	"cache/logger"
	"cache/storage"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

func parseList(s string) ([]string, bool) {
	var l []string
	if err := json.Unmarshal([]byte(s), &l); err != nil {
		return nil, false
	}

	return l, true
}

func serializeList(l []string) string {
	b, _ := json.Marshal(l)
	return string(b)
}

// Create empty list
func ESET(c *storage.Cache, l string) {
	c.WithLock(func() {
		if _, exists := c.GetUnsafe(l); exists {
			logger.Error("%s already exists", l)
			return
		}

		slice := make([]string, 0)

		c.SetUnsafe(l, &storage.CacheData{
			Value:     serializeList(slice),
			Type:      storage.List,
			Requests:  1,
			CreatedAt: time.Now(),
		})

		logger.Success("OK")
	})
}

// Set value at index in the list
func LSET(c *storage.Cache, l string, s ...string) {
	if len(s) == 0 || len(s)%2 != 0 {
		logger.Error("Not enough values")
		return
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(l)

		var slice []string

		if !exists {
			slice = []string{}
		} else {
			var ok bool
			slice, ok = parseList(cd.Value)
			if !ok {
				logger.Error("%s isn`t list", l)
				return
			}
		}

		for i := 0; i < len(s); i += 2 {
			index, err := strconv.Atoi(s[i])
			if err != nil {
				logger.Error("Index must be a number: %s", s[i])
				return
			}

			v := s[i+1]

			if index < 0 {
				logger.Error("Index out of range: %v", index)
				return
			}

			if index >= len(slice) {
				nS := make([]string, index+1)
				copy(nS, slice)
				slice = nS
			}

			slice[index] = v
		}

		if !exists {
			c.SetUnsafe(l, &storage.CacheData{
				Value:     serializeList(slice),
				Type:      storage.List,
				Requests:  1,
				CreatedAt: time.Now(),
			})
		} else {
			cd.Value = serializeList(slice)
			cd.Requests++
		}

		logger.Success("OK")
	})
}

// Get value at index in the list
func LGET(c *storage.Cache, l string, s ...string) {
	if len(s) == 0 {
		logger.Error("Not enough values")
		return
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(l)

		var slice []string

		if !exists {
			logger.Error("Can`t find %v in memory", l)
			return
		} else {
			cd, ok := parseList(cd.Value)
			if !ok {
				logger.Error("%s isn`t list", l)
				return
			}

			for i := 0; i < len(s); i++ {
				index, err := strconv.Atoi(s[i])
				if err != nil {
					logger.Error("Index must be a number: %s", s[i])
					return
				}

				slice = append(slice, cd[index])
			}
		}

		fmt.Println(slice)
	})
}

// Push to the start of the list
func SPUSH(c *storage.Cache, l string, s ...string) {
	if len(s) == 0 {
		logger.Error("Not enough values")
		return
	}

	// if ok := removeQuotes(&s, 0, 1); !ok {
	// 	return
	// }

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(l)

		if !exists {
			c.SetUnsafe(l, &storage.CacheData{
				Value:     serializeList(append([]string{}, s...)),
				Type:      storage.List,
				Requests:  1,
				CreatedAt: time.Now(),
			})
		} else {
			slice, ok := parseList(cd.Value)
			if !ok {
				logger.Error("%s isn`t list", l)
				return
			}

			slice = append(s, slice...)
			cd.Value = serializeList(slice)
			cd.Requests++
		}

		logger.Success("OK")
	})
}

// Push to the end of the list
func EPUSH(c *storage.Cache, l string, s ...string) {
	if len(s) == 0 {
		logger.Error("Not enough values")
		return
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(l)

		if !exists {
			c.SetUnsafe(l, &storage.CacheData{
				Value:     serializeList(append([]string{}, s...)),
				Type:      storage.List,
				Requests:  1,
				CreatedAt: time.Now(),
			})
		} else {
			slice, ok := parseList(cd.Value)
			if !ok {
				logger.Error("%s isn`t list", l)
				return
			}

			slice = append(slice, s...)
			cd.Value = serializeList(slice)
			cd.Requests++
		}

		logger.Success("OK")
	})
}

// Remove from the start of the list
func SPOP(c *storage.Cache, l, s string) {
	q := 1

	if s != "" {
		var err error
		q, err = strconv.Atoi(s)
		if err != nil {
			logger.Error("Can`t parse number")
			return
		}
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(l)
		if !exists {
			logger.Error("Can`t find %v in memory", l)
			return
		}

		slice, ok := parseList(cd.Value)
		if !ok {
			logger.Error("%s isn`t list", l)
			return
		}

		if len(slice) < q {
			q = len(slice)
		}

		slice = slice[q:]

		cd.Value = serializeList(slice)

		cd.Requests++

		logger.Success("OK")
	})
}

// Remove from the end of the list
func EPOP(c *storage.Cache, l, s string) {
	q := 1

	if s != "" {
		var err error
		q, err = strconv.Atoi(s)
		if err != nil {
			logger.Error("Can`t parse number")
			return
		}
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(l)
		if !exists {
			logger.Error("Can`t find %v in memory", l)
			return
		}

		slice, ok := parseList(cd.Value)
		if !ok {
			logger.Error("%s isn`t list", l)
			return
		}

		if q >= len(slice) {
			slice = []string{}
		} else {
			slice = slice[q:]
		}

		cd.Value = serializeList(slice)

		cd.Requests++

		logger.Success("OK")
	})
}

// Get list of l list in a given range
func SRANGE(c *storage.Cache, l string, s []string) {
	if len(s) != 2 {
		logger.Error("Not enough values")
		return
	}

	start, err := strconv.Atoi(s[0])
	if err != nil {
		logger.Error("Can`t parse number")
		return
	}

	end, err := strconv.Atoi(s[1])
	if err != nil {
		logger.Error("Can`t parse number")
		return
	}

	cd, exists := c.GetUnsafe(l)

	if !exists {
		logger.Error("Can`t find %v in memory", l)
		return
	}

	v, ok := parseList(cd.Value)
	if !ok {
		logger.Error("Value is not a slice")
		return
	}
	length := len(v)

	if end == -1 || end > length {
		end = length
	}

	if start < 0 {
		start = 0
	}

	if start > end {
		logger.Error("Invalid range")
	}

	list := make([]string, 0, end-start)

	for i := start; i < end; i++ {
		// fmt.Printf("%v) %v\n", i+1, v[i])
		list = append(list, v[i])
	}

	fmt.Println(list)
}

// Check if the values exist in the list and return their quantity
func CONTAINS(c *storage.Cache, l string, ks []string) {
	var res int

	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(l)
		if !exists {
			logger.Error("can`t find %v in memory", l)
			return
		}

		cd.Requests++

		s, ok := parseList(cd.Value)
		if !ok {
			logger.Error("%s isn`t list", l)
			return
		}

		for _, v := range s {
			for _, k := range ks {
				if k == v {
					res++
				}
			}
		}
	})

	fmt.Println(res)
}

// Check if the values exist in the list and return list
func LCONTAINS(c *storage.Cache, l string, ks []string) {
	res := make([]int, len(ks))
	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(l)
		if !exists {
			logger.Error("can`t find %v in memory", l)
			return
		}

		cd.Requests++

		s, ok := parseList(cd.Value)
		if !ok {
			logger.Error("%s isn`t list", l)
			return
		}

		tM := make(map[string]struct{})
		for _, v := range s {
			tM[v] = struct{}{}
		}

		for i, k := range ks {
			if _, exists := tM[k]; exists {
				res[i] = 1
			} else {
				res[i] = 0
			}
		}
	})

	fmt.Println(res)
}

// Return index of target value in list
func INDEXOF(c *storage.Cache, l, k string) {
	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(l)
		if !exists {
			logger.Error("can`t find %v in memory", l)
			return
		}

		cd.Requests++

		s, ok := parseList(cd.Value)
		if !ok {
			logger.Error("%s isn`t list", l)
			return
		}

		tI := -1

		for i, v := range s {
			if v == k {
				tI = i
				break
			}
		}

		fmt.Println(tI)
	})
}

// Get length of the list
func LLEN(c *storage.Cache, l string) {
	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(l)

		if !exists {
			logger.Error("Can`t find %v in memory", l)
			return
		} else {
			cd, ok := parseList(cd.Value)
			if !ok {
				logger.Error("%s isn`t list", l)
				return
			}
			fmt.Println(len(cd))
		}

	})
}

// Clear the list
func LCLEAR(c *storage.Cache, l string) {
	c.WithLock(func() {
		cd, exists := c.GetUnsafe(l)

		if !exists {
			logger.Error("Can`t find %v in memory", l)
			return
		} else {
			_, ok := parseList(cd.Value)
			if !ok {
				logger.Error("%s isn`t list", l)
				return
			}

			cd.Value = serializeList([]string{})
			cd.Requests++

			logger.Success("OK")
		}

	})
}
