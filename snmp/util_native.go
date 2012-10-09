package snmp

// #include <stdint.h>
// #include <sys/types.h>
// #include <string.h>
// #include <stdlib.h>
// #include <string.h>
import "C"

import (
	"errors"
	"unsafe"
)

func memcpy(dst *C.uint8_t, capacity int, src []byte) error {
	if 0 == len(src) {
		return nil
	}
	if capacity < len(src) {
		return errors.New("bytes too long.")
	}

	C.memcpy(unsafe.Pointer(dst), unsafe.Pointer(&src[0]), C.size_t(len(src)))
	return nil
}

func readGoBytes(src *C.uint8_t, length C.uint32_t) []byte {

	if 0 == length {
		return []byte{}
	}

	return C.GoBytes(unsafe.Pointer(src), C.int(length))
}

func readGoString(src *C.char, capacity int) string {
	if 0 == capacity {
		return ""
	}
	length := int(C.strlen(src))
	if capacity < length {
		panic("string too long.")
		return "" //errors.New("string too long.")
	}
	return C.GoStringN(src, C.int(length))
}

func strcpy(dst *C.char, capacity int, src string) error {
	if 0 == len(src) {
		return nil
	}
	if capacity < len(src) {
		return errors.New("string too long.")
	}
	s := C.CString(src)
	C.strcpy(dst, s)
	C.free(unsafe.Pointer(s))
	return nil
}
