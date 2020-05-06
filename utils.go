package gearbox

import (
	"unsafe"
)

// getString gets the content of a string as a []byte without copying
func getString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
