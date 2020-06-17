package gearbox

import (
	"unsafe"
)

// #nosec G103
// getString gets the content of a string as a []byte without copying
func getString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
