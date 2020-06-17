package gearbox

import (
	"unsafe"
)

// GetString gets the content of a string as a []byte without copying
// #nosec G103
func GetString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
