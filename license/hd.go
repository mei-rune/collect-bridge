package license

// int GetDiskInfo(int driver, char* szModelNumber, char* szSerialNumber, int len);
import "C"
import (
	"bytes"
	"unsafe"
)

func nilString(b []byte) string {
	i := bytes.IndexByte(b, byte(0))
	if -1 == i {
		return string(b)
	}
	return string(b[0:i])
}

func GetHD(i int) (string, string, int) {
	szModelNumber := make([]byte, 64)
	szSerialNumber := make([]byte, 64)
	ret := C.GetDiskInfo(C.int(i), (*C.char)(unsafe.Pointer(&szModelNumber[0])),
		(*C.char)(unsafe.Pointer(&szSerialNumber[0])), 64)
	if 0 != ret {
		return "", "", int(ret)
	}
	return nilString(szModelNumber), nilString(szSerialNumber), 0
}

func GetAllHD() [][2]string {
	ret := make([][2]string, 0, 1)
	for i := 0; i < 256; i++ {
		m, s, r := GetHD(i)
		if 0 != r {
			return ret
		}
		ret = append(ret, [2]string{m, s})
	}
	return ret
}
