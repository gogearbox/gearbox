package gearbox

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// ExampleTST tests TST set and get methods
func ExampleTST() {
	tst := newTST()
	tst.Set([]byte("user"), 1)
	fmt.Println(tst.Get([]byte("user")).(int))
	fmt.Println(tst.Get([]byte("us")))
	fmt.Println(tst.Get([]byte("user1")))
	fmt.Println(tst.Get([]byte("not-existing")))
	fmt.Println(tst.GetString(("not-existing")))

	tst.Set([]byte("account"), 5)
	tst.Set([]byte("account"), 6)
	fmt.Println(tst.Get([]byte("account")).(int))

	tst.Set([]byte("acc@unt"), 12)
	fmt.Println(tst.Get([]byte("acc@unt")).(int))

	tst.Set([]byte("حساب"), 15)
	fmt.Println(tst.Get([]byte("حساب")).(int))
	tst.Set([]byte(""), 14)
	fmt.Println(tst.Get([]byte("")))
	// Output:
	// 1
	// <nil>
	// <nil>
	// <nil>
	// <nil>
	// 6
	// 12
	// 15
	// <nil>
}

// RandBytes generates random string from English letters
func RandBytes() []byte {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, rand.Intn(100))
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return b
}

func BenchmarkTSTLookup(b *testing.B) {
	tst := newTST()
	rand.Seed(time.Now().UnixNano())
	for n := 0; n < rand.Intn(2000); n++ {
		tst.Set(RandBytes(), rand.Intn(10000))
	}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		tst.Get([]byte("user"))
	}
}
