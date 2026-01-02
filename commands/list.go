package commands

import (
	"cache/logger"
	"cache/storage"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// Set value at index in the array
func LSET(c *storage.Cache, aN string, s []string) {
	if len(s) == 0 || len(s)%2 != 0 {
		logger.Error("Not enough values")
		return
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(aN)

		var slice []string

		if !exists {
			slice = []string{}
		} else {
			var ok bool
			slice, ok = cd.Value.([]string)
			if !ok {
				logger.Error("%s isn`t array", aN)
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
			c.SetUnsafe(aN, &storage.CacheData{
				Value:     slice,
				Requests:  1,
				TimeStamp: time.Now(),
			})
		} else {
			cd.Value = slice
			cd.Requests++
		}

		logger.Success("OK")
	})
}

// Get value at index in the array
func LGET(c *storage.Cache, aN string, s []string) {
	if len(s) == 0 {
		logger.Error("Not enough values")
		return
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(aN)

		var slice []string

		if !exists {
			logger.Error("Can`t find %v in memory", aN)
			return
		} else {
			cd, ok := cd.Value.([]string)
			if !ok {
				logger.Error("%s isn`t array", aN)
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

// Push to the start of the array
func SPUSH(c *storage.Cache, aN string, s []string) {
	if len(s) == 0 {
		logger.Error("Not enough values")
		return
	}

	// if ok := removeQuotes(&s, 0, 1); !ok {
	// 	return
	// }

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(aN)

		if !exists {
			c.SetUnsafe(aN, &storage.CacheData{
				Value:     append([]string{}, s...),
				Requests:  1,
				TimeStamp: time.Now(),
			})
		} else {
			slice, ok := cd.Value.([]string)
			if !ok {
				logger.Error("%s isn`t array", aN)
				return
			}

			cd.Value = append(s, slice...)
			cd.Requests++
		}

		logger.Success("OK")
	})
}

// Push to the end of the array
func EPUSH(c *storage.Cache, aN string, s []string) {
	if len(s) == 0 {
		logger.Error("Not enough values")
		return
	}

	// if ok := removeQuotes(&s, 0, 1); !ok {
	// 	return
	// }

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(aN)

		if !exists {
			c.SetUnsafe(aN, &storage.CacheData{
				Value:     append([]string{}, s...),
				Requests:  1,
				TimeStamp: time.Now(),
			})
		} else {
			slice, ok := cd.Value.([]string)
			if !ok {
				logger.Error("%s isn`t array", aN)
				return
			}

			cd.Value = append(slice, s...)
			cd.Requests++
		}

		logger.Success("OK")
	})
}

// Remove from the start of the array
func SPOP(c *storage.Cache, aN, s string) {
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
		cd, exists := c.GetUnsafe(aN)
		if !exists {
			logger.Error("Can`t find %v in memory", aN)
			return
		}

		slice, ok := cd.Value.([]string)
		if !ok {
			logger.Error("%s isn`t array", aN)
			return
		}

		if len(slice) < q {
			logger.Error("Current q is bigger then len")
			q = len(slice)
		}

		slice = slice[q:]

		if len(slice) == 0 {
			delete(c.GetData(), aN)
		} else {
			cd.Value = slice
		}

		cd.Requests++

		logger.Success("OK")
	})
}

// Remove from the end of the array
func EPOP(c *storage.Cache, aN, s string) {
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
		cd, exists := c.GetUnsafe(aN)
		if !exists {
			logger.Error("Can`t find %v in memory", aN)
			return
		}

		slice, ok := cd.Value.([]string)
		if !ok {
			logger.Error("%s isn`t array", aN)
			return
		}

		if len(slice) < q {
			logger.Error("Current q is bigger then len")
			q = len(slice)
		}

		slice = slice[:q]

		if len(slice) == 0 {
			delete(c.GetData(), aN)
		} else {
			cd.Value = slice
		}

		cd.Requests++

		logger.Success("OK")
	})
}

// Get list of an array in a given range
func SRANGE(c *storage.Cache, aN string, s []string) {
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

	cd, exists := c.GetUnsafe(aN)

	if !exists {
		logger.Error("Can`t find %v in memory", aN)
		return
	}

	v := reflect.ValueOf(cd.Value)

	if v.Kind() != reflect.Slice {
		logger.Error("Value is not a slice")
		return
	}
	length := v.Len()

	if end == -1 || end > length {
		end = length
	}

	if start < 0 {
		start = 0
	}

	if start > end {
		logger.Error("Invalid range")
	}

	for i := start; i < end; i++ {
		fmt.Printf("%v) %v\n", i+1, v.Index(i).Interface())
	}
}

// Check if the values exist in the array and return their quantity
func CONTAINS(c *storage.Cache, aN string, ks []string) {
	var res int

	cd, exists := c.GetSafe(aN)
	if !exists {
		logger.Error("can`t find %v in memory", aN)
		return
	}

	// CHECK
	cd.Requests++

	s, ok := cd.Value.([]string)
	if !ok {
		logger.Error("%s isn`t array", aN)
		return
	}

	for _, v := range s {
		for _, k := range ks {
			if k == v {
				res++
			}
		}
	}

	fmt.Println(res)
}

// Check if the values exist in the array and return array
func LCONTAINS(c *storage.Cache, aN string, ks []string) {
	res := make([]int, len(ks))

	cd, exists := c.GetSafe(aN)
	if !exists {
		logger.Error("can`t find %v in memory", aN)
		return
	}

	// CHECK
	cd.Requests++

	s, ok := cd.Value.([]string)
	if !ok {
		logger.Error("%s isn`t array", aN)
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

	fmt.Println(res)
}

// Get length of the array
func LLEN(c *storage.Cache, aN string) {
	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(aN)

		if !exists {
			logger.Error("Can`t find %v in memory", aN)
			return
		} else {
			cd, ok := cd.Value.([]string)
			if !ok {
				logger.Error("%s isn`t array", aN)
				return
			}
			fmt.Println(len(cd))
		}

	})
}

// Clear the array
func LCLEAR(c *storage.Cache, aN string) {
	c.WithLock(func() {
		cd, exists := c.GetUnsafe(aN)

		if !exists {
			logger.Error("Can`t find %v in memory", aN)
			return
		} else {
			_, ok := cd.Value.([]string)
			if !ok {
				logger.Error("%s isn`t array", aN)
				return
			}

			cd.Value = []string{}
			cd.Requests++

			logger.Success("OK")
		}

	})
}
