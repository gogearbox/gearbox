package gearbox

import (
	"fmt"
)

// ExampleCache tests Cache set and get methods
func ExampleCache() {
	cache := newCache(3)
	cache.Set("user1", 1)
	fmt.Println(cache.Get("user1").(int))

	cache.Set("user2", 2)
	fmt.Println(cache.Get("user2").(int))

	cache.Set("user3", 3)
	fmt.Println(cache.Get("user3").(int))

	cache.Set("user4", 4)
	fmt.Println(cache.Get("user1"))
	fmt.Println(cache.Get("user2").(int))

	cache.Set("user5", 5)
	fmt.Println(cache.Get("user3"))

	cache.Set("user5", 6)
	fmt.Println(cache.Get("user5").(int))

	cache2 := newCache(0)
	cache2.Set("user1", 1)
	fmt.Println(cache2.Get("user1").(int))

	// Output:
	// 1
	// 2
	// 3
	// <nil>
	// 2
	// <nil>
	// 6
	// 1
}
