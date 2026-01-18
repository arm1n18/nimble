package commands

import (
	"fmt"
	"nimble/formatter"
	"nimble/storage"
	"strconv"
	"unsafe"
)

// Set max memory size
func MAXMEM(c *storage.Cache, n string) {
	N_INT, err := strconv.Atoi(n)
	if err != nil {
		// formatter.ErrNotANumber.Error()
		return
	}
	storage.MAX_MEMO = N_INT
}

// Check if the memory sizes of the data are equal
func COMPARE(c *storage.Cache, args ...string) string {
	var result string

	c.WithRWLock(func() {
		cd := c.GetData()

		for _, k := range args {
			_, exists := cd[k]
			if !exists {
				result = formatter.ErrorMessage("Can`t find %v in memory", k)
				return
			}
		}

		result = fmt.Sprint(unsafe.Sizeof(*cd[args[0]]) == unsafe.Sizeof(*cd[args[1]]))
	})

	return result
}

// Check the memory size of the data
func SIZEOF(c *storage.Cache, args ...string) {
	c.WithLock(func() {
		cd := c.GetData()

		// TODO
		var b uint64
		if len(args) == 1 && args[0] == "*" {
			for _, k := range cd {
				b += uint64(unsafe.Sizeof(*k))
			}
			fmt.Printf("Total size is %v\n", b)
			return
		}

		for _, k := range args {
			// cd, exists := cd[k]
			// if !exists {
			// 	formatter.ErrorMessage("Can`t find %v in memory", k)
			// 	return
			// }

			// switch v := cd.Type {
			// case string:
			// 	b += uint64(unsafe.Sizeof(v))
			// 	b += uint64(len(v))
			// case int:
			// 	b += uint64(unsafe.Sizeof(v))
			// case []string:
			// 	for _, i := range v {
			// 		b += uint64(unsafe.Sizeof(i))
			// 		b += uint64(len(i))
			// 	}
			// case map[string]string:
			// 	b += uint64(unsafe.Sizeof(v))
			// 	for k, v := range v {
			// 		b += uint64(len(k))
			// 		b += uint64(unsafe.Sizeof(k))
			// 		b += uint64(len(v))
			// 		b += uint64(unsafe.Sizeof(v))
			// 	}
			// }

			fmt.Printf("Size of %s is %v\n", k, b)
		}
	})
}

// Show the type of data stored in the cache
func TYPE(c *storage.Cache, k string) string {
	var result string

	c.WithRWLock(func() {
		cd, exists := c.GetUnsafe(k)
		if !exists {
			result = formatter.ErrorMessage("Can`t find %v in memory", k)
			return
		}

		result = string(cd.Type)
	})

	return result
}
