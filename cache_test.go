package gearbox

import "fmt"

// ExampleCache tests Cache set and get methods
func ExampleCache() {
	cache := newCache(3)
	cache.Set([]byte("user1"), 1)
	fmt.Println(cache.Get([]byte("user1")).(int))

	cache.Set([]byte("user2"), 2)
	fmt.Println(cache.Get([]byte("user2")).(int))

	cache.Set([]byte("user3"), 3)
	fmt.Println(cache.Get([]byte("user3")).(int))

	cache.Set([]byte("user4"), 4)
	fmt.Println(cache.Get([]byte("user1")))
	fmt.Println(cache.Get([]byte("user2")).(int))

	cache.Set([]byte("user5"), 5)
	fmt.Println(cache.Get([]byte("user3")))

	cache.Set([]byte("user5"), 6)
	fmt.Println(cache.Get([]byte("user5")).(int))

	cache2 := newCache(0)
	cache2.Set([]byte("user1"), 1)
	fmt.Println(cache2.Get([]byte("user1")).(int))
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
