package commands

import (
	"cache/logger"
	"cache/storage"
	"fmt"
	"strconv"
	"unsafe"
)

// Set max memory size
func MAXMEM(c *storage.Cache, n string) {
	N_INT, err := strconv.Atoi(n)
	if err != nil {
		logger.Error("Can`t parse number")
		return
	}
	storage.MAX_MEMO = N_INT
}

// Check if the memory sizes of the data are equal
func COMPARE(c *storage.Cache, ks []string) {
	c.WithRWLock(func() {
		cd := c.GetData()

		for _, k := range ks {
			_, exists := cd[k]
			if !exists {
				logger.Error("Can`t find %v in memory", k)
				return
			}
		}

		fmt.Println(unsafe.Sizeof(*cd[ks[0]]) == unsafe.Sizeof(*cd[ks[1]]))
	})
}

// Check the memory size of the data
func SIZEOF(c *storage.Cache, ks []string) {
	c.WithLock(func() {
		cd := c.GetData()

		// TODO
		var b uint64
		if len(ks) == 1 && ks[0] == "*" {
			for _, k := range cd {
				b += uint64(unsafe.Sizeof(*k))
			}
			fmt.Printf("Total size is %v\n", b)
			return
		}

		for _, k := range ks {
			// cd, exists := cd[k]
			// if !exists {
			// 	logger.Error("Can`t find %v in memory", k)
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
func TYPE(c *storage.Cache, k string) {
	cd, exists := c.GetSafe(k)
	if !exists {
		logger.Error("Can`t find %v in memory", k)
		return
	}

	fmt.Println(cd.Type)
}
