package commands

import (
	"encoding/json"
	"nimble/protocol"
	"nimble/storage"
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
func ESET(c *storage.Cache, l string) string {
	var result string

	c.WithLock(func() {
		arr := make([]string, 0)

		c.SetUnsafe(l, &storage.CacheData{
			Value:     serializeList(arr),
			Type:      storage.List,
			Requests:  1,
			CreatedAt: time.Now(),
		})

		result = protocol.Success()
	})

	return result
}

// Set value at index in the list
func LSET(c *storage.Cache, l string, s ...string) string {
	var result string

	if len(s) == 0 || len(s)%2 != 0 {
		return protocol.ErrNotEnoughValues.Error()
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(l)

		var arr []string
		var q int

		if !exists {
			arr = []string{}
		} else {
			var ok bool
			arr, ok = parseList(cd.Value)
			if !ok {
				// result = protocol.ErrorMessage("%s isn`t list", l)
				result = protocol.ErrMismatchType.Error()
				return
			}
		}

		for i := 0; i < len(s); i += 2 {
			index, err := strconv.Atoi(s[i])
			if err != nil {
				// result = protocol.ErrorMessage("Index must be a number: %s", s[i])
				// result = protocol.ErrNotANumber.Error()
				continue
			}

			if index < 0 {
				// result = protocol.ErrorMessage("Index out of range: %v", index)
				result = protocol.ErrInvalidRange.Error()
				continue
			}

			if index >= len(arr) {
				nS := make([]string, index+1)
				copy(nS, arr)
				arr = nS
			}

			arr[index] = s[i+1]
			q++
		}

		if !exists {
			c.SetUnsafe(l, &storage.CacheData{
				Value:     serializeList(arr),
				Type:      storage.List,
				Requests:  1,
				CreatedAt: time.Now(),
			})
		} else {
			cd.Value = serializeList(arr)
			cd.Requests++
		}

		result = protocol.Number(q)
	})

	return result
}

// Get value at index in the list
func LGET(c *storage.Cache, l string, s ...string) string {
	var result string

	if len(s) == 0 {
		return protocol.ErrNotEnoughValues.Error()
	}

	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(l)
		if !exists {
			// result = protocol.ErrorMessage("Can`t find %v in memory", l)
			result = protocol.Array("[]")
			return
		} else {
			list, ok := parseList(cd.Value)
			if !ok {
				// result = protocol.ErrorMessage("%s isn`t list", l)
				result = protocol.ErrMismatchType.Error()
				return
			}

			res := make([]string, 0, len(s))
			for _, v := range s {
				index, err := strconv.Atoi(v)
				if err != nil {
					// result = protocol.ErrorMessage("Index must be a number: %s", s[i])
					// result = protocol.ErrNotANumber.Error()
					return
				}

				if index < 0 || index >= len(list) {
					res = append(res, protocol.Nil())
					continue
				}

				res = append(res, list[index])
			}

			result = protocol.Array(serializeList(res))
		}
	})

	return result
}

// Push to the start of the list
func SPUSH(c *storage.Cache, l string, s ...string) string {
	var result string

	if len(s) == 0 {
		return protocol.ErrNotEnoughValues.Error()
	}

	// if ok := removeQuotes(&s, 0, 1); !ok {
	// 	return
	// }

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(l)

		if !exists {
			value := append([]string{}, s...)

			c.SetUnsafe(l, &storage.CacheData{
				Value:     serializeList(value),
				Type:      storage.List,
				Requests:  1,
				CreatedAt: time.Now(),
			})
		} else {
			arr, ok := parseList(cd.Value)
			if !ok {
				// result = protocol.ErrorMessage("%s isn`t list", l)
				result = protocol.ErrMismatchType.Error()
				return
			}

			arr = append(s, arr...)
			cd.Value = serializeList(arr)
			cd.Requests++
		}

		result = protocol.Number(len(s))
	})

	return result
}

// Push to the end of the list
func EPUSH(c *storage.Cache, l string, s ...string) string {
	var result string

	if len(s) == 0 {
		return protocol.ErrNotEnoughValues.Error()
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(l)

		if !exists {
			value := append([]string{}, s...)

			c.SetUnsafe(l, &storage.CacheData{
				Value:     serializeList(value),
				Type:      storage.List,
				Requests:  1,
				CreatedAt: time.Now(),
			})
		} else {
			arr, ok := parseList(cd.Value)
			if !ok {
				// result = protocol.ErrorMessage("%s isn`t list", l)
				result = protocol.ErrMismatchType.Error()
				return
			}

			arr = append(arr, s...)
			cd.Value = serializeList(arr)
			cd.Requests++
		}

		result = protocol.Number(len(s))
	})

	return result
}

// Remove from the start of the list
func SPOP(c *storage.Cache, l, s string) string {
	var result string

	q := 1

	if s != "" {
		var err error
		q, err = strconv.Atoi(s)
		if err != nil {
			return protocol.ErrNotANumber.Error()
		}
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(l)
		if !exists {
			// result = protocol.ErrorMessage("Can`t find %v in memory", l)
			result = protocol.Failure()
			return
		}

		list, ok := parseList(cd.Value)
		if !ok {
			// protocol.ErrorMessage("%s isn`t list", l)
			result = protocol.ErrMismatchType.Error()
			return
		}

		if len(list) == 0 {
			result = protocol.Array("[]")
			return
		}

		if len(list) < q {
			q = len(list)
		}

		rm := append([]string{}, list[:q]...)
		cd.Value = serializeList(list[q:])
		cd.Requests++

		result = protocol.Array(serializeList(rm))
	})

	return result
}

// Remove from the end of the list
func EPOP(c *storage.Cache, l, s string) string {
	var result string

	q := 1

	if s != "" {
		var err error
		q, err = strconv.Atoi(s)
		if err != nil {
			return protocol.ErrNotANumber.Error()
		}
	}

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(l)
		if !exists {
			// result = protocol.ErrorMessage("Can`t find %v in memory", l)
			result = protocol.Failure()
			return
		}

		list, ok := parseList(cd.Value)
		if !ok {
			// result = protocol.ErrorMessage("%s isn`t list", l)
			result = protocol.ErrMismatchType.Error()
			return
		}

		if len(list) == 0 {
			result = protocol.Array("[]")
			return
		}

		if q > len(list) {
			q = len(list)
		}

		start := len(list) - q
		rm := append([]string{}, list[start:]...)

		cd.Value = serializeList(list[:start])
		cd.Requests++

		result = protocol.Array(serializeList(rm))
	})

	return result
}

// Get list of l list in a given range
func SRANGE(c *storage.Cache, l string, s []string) string {
	var result string

	if len(s) != 2 {
		return protocol.ErrNotEnoughValues.Error()
	}

	start, err := strconv.Atoi(s[0])
	if err != nil {
		return protocol.ErrNotANumber.Error()
	}

	end, err := strconv.Atoi(s[1])
	if err != nil {
		return protocol.ErrNotANumber.Error()
	}

	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(l)

		if !exists {
			// result = protocol.ErrorMessage("Can`t find %v in memory", l)
			result = protocol.Array("[]")
			return
		}

		v, ok := parseList(cd.Value)
		if !ok {
			// result = protocol.ErrorMessage("%s isn`t list", l)
			result = protocol.ErrMismatchType.Error()
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
			result = protocol.ErrInvalidRange.Error()
			return
		}

		arr := make([]string, 0, end-start)

		for i := start; i < end; i++ {
			arr = append(arr, v[i])
		}

		result = protocol.Array(serializeList(arr))
	})

	return result
}

// Check if the values exist in the list and return their quantity
func CONTAINS(c *storage.Cache, l string, ks []string) string {
	var result string

	c.WithRWLock(func() {
		var q int

		cd, exists := c.GetUnsafe(l)
		if !exists {
			// result = protocol.ErrorMessage("can`t find %v in memory", l)
			result = protocol.Number(0)
			return
		}

		cd.Requests++

		s, ok := parseList(cd.Value)
		if !ok {
			// result = protocol.ErrorMessage("%s isn`t list", l)
			result = protocol.ErrMismatchType.Error()
			return
		}

		m := make(map[string]int)
		for _, v := range s {
			m[v]++
		}

		for _, k := range ks {
			q += m[k]
		}

		result = protocol.Number(q)
	})

	return result
}

// Check if the values exist in the list and return list
func LCONTAINS(c *storage.Cache, l string, ks []string) string {
	var result string

	c.WithRWLock(func() {
		arr := make([]string, len(ks))

		cd, exists := c.GetUnsafe(l)
		if !exists {
			// result = protocol.ErrorMessage("can`t find %v in memory", l)
			result = protocol.Array("[]")
			return
		}

		cd.Requests++

		s, ok := parseList(cd.Value)
		if !ok {
			// result = protocol.ErrorMessage("%s isn`t list", l)
			result = protocol.ErrMismatchType.Error()
			return
		}

		tM := make(map[string]struct{})
		for _, v := range s {
			tM[v] = struct{}{}
		}

		for i, k := range ks {
			if _, exists := tM[k]; exists {
				arr[i] = "1"
			} else {
				arr[i] = "0"
			}
		}

		result = protocol.Array(serializeList(arr))
	})

	return result
}

// Return index of target value in list
func INDEXOF(c *storage.Cache, l, k string) string {
	var result string

	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(l)
		if !exists {
			// result = protocol.ErrorMessage("can`t find %v in memory", l)
			result = protocol.Number(-1)
			return
		}

		cd.Requests++

		s, ok := parseList(cd.Value)
		if !ok {
			// result = protocol.ErrorMessage("%s isn`t list", l)
			result = protocol.ErrMismatchType.Error()
			return
		}

		i := -1

		for j, v := range s {
			if v == k {
				i = j
				break
			}
		}

		result = protocol.Number(i)
	})

	return result
}

// Get length of the list
func LLEN(c *storage.Cache, l string) string {
	var result string

	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(l)

		if !exists {
			// result = protocol.ErrorMessage("Can`t find %v in memory", l)
			result = protocol.Number(-1)
			return
		} else {
			list, ok := parseList(cd.Value)
			if !ok {
				// result = protocol.ErrorMessage("%s isn`t list", l)
				result = protocol.ErrMismatchType.Error()
				return
			}
			result = protocol.Number(len(list))
		}
	})

	return result
}

// Clear the list
func LCLEAR(c *storage.Cache, l string) string {
	var result string

	c.WithLock(func() {
		cd, exists := c.GetUnsafe(l)

		if !exists {
			// result = protocol.ErrorMessage("Can`t find %v in memory", l)
			result = protocol.Failure()
			return
		} else {
			_, ok := parseList(cd.Value)
			if !ok {
				// result = protocol.ErrorMessage("%s isn`t list", l)
				result = protocol.ErrMismatchType.Error()
				return
			}

			cd.Value = serializeList([]string{})
			cd.Requests++

			result = protocol.Success()
		}
	})

	return result
}
