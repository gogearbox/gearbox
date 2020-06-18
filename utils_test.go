package gearbox

import "fmt"

// ExampleGetString tests converting []byte to string
func ExampleGetString() {
	b := []byte("ABC€")
	str := GetString(b)
	fmt.Println(str)
	fmt.Println(len(b) == len(str))

	b = []byte("مستخدم")
	str = GetString(b)
	fmt.Println(str)
	fmt.Println(len(b) == len(str))

	b = nil
	str = GetString(b)
	fmt.Println(str)
	fmt.Println(len(b) == len(str))
	fmt.Println(len(str))

	// Output:
	// ABC€
	// true
	// مستخدم
	// true
	//
	// true
	// 0
}
