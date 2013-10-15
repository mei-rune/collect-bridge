package main

import (
	"code.google.com/p/winsvc/winapi"
	"errors"
	"syscall"
	"unsafe"
)

type Key struct {
	Handle syscall.Handle
}

func OpenKey(parent syscall.Handle, path string) (*Key, error) {
	var h syscall.Handle
	e := syscall.RegOpenKeyEx(
		parent, syscall.StringToUTF16Ptr(path),
		0, syscall.KEY_ALL_ACCESS, &h)
	if e != nil {
		return nil, e
	}
	return &Key{Handle: h}, nil
}

func (k *Key) Close() error {
	return syscall.RegCloseKey(k.Handle)
}

func (k *Key) CreateSubKey(name string) (nk *Key, openedExisting bool, err error) {
	var h syscall.Handle
	var d uint32
	e := winapi.RegCreateKeyEx(k.Handle, syscall.StringToUTF16Ptr(name),
		0, nil, winapi.REG_OPTION_NON_VOLATILE,
		syscall.KEY_ALL_ACCESS, nil, &h, &d)
	if e != nil {
		return nil, false, e
	}
	return &Key{Handle: h}, d == winapi.REG_OPENED_EXISTING_KEY, nil
}

func (k *Key) DeleteSubKey(name string) error {
	return winapi.RegDeleteKey(k.Handle, syscall.StringToUTF16Ptr(name))
}

func (k *Key) SetUInt32(name string, value uint32) error {
	return winapi.RegSetValueEx(
		k.Handle, syscall.StringToUTF16Ptr(name),
		0, syscall.REG_DWORD,
		(*byte)(unsafe.Pointer(&value)), uint32(unsafe.Sizeof(value)))
}

func (k *Key) GetUInt32(name string) (uint32, error) {
	value := uint32(0)
	typ := uint32(0)
	n := uint32(unsafe.Sizeof(value))

	if e := syscall.RegQueryValueEx(
		k.Handle, syscall.StringToUTF16Ptr(name),
		nil, &typ, (*byte)(unsafe.Pointer(&value)), &n); nil != e {
		return 0, e
	}

	if typ != syscall.REG_DWORD { // null terminated strings only
		return 0, errors.New("key is not a REG_DWORD")
	}
	return value, nil
}

func (k *Key) GetString(name string) (string, error) {
	// 	LONG RegQueryValueEx(
	// 　　HKEY hKey, // handle to key
	// 　　LPCTSTR lpValueName, // value name
	// 　　LPDWORD lpReserved, // reserved
	// 　　LPDWORD lpType, // type buffer
	// 　　LPBYTE lpData, // data buffer
	// 　　LPDWORD lpcbData // size of data buffer
	// 　　);

	var typ, n uint32
	if e := syscall.RegQueryValueEx(
		k.Handle, syscall.StringToUTF16Ptr(name),
		nil, &typ, nil, &n); nil != e {
		return "", e
	}

	if typ != syscall.REG_SZ { // null terminated strings only
		return "", errors.New("key is not a string")
	}

	buf := make([]uint16, int(n), int(n)+16)
	n = uint32(cap(buf) * 2)
	if e := syscall.RegQueryValueEx(
		k.Handle, syscall.StringToUTF16Ptr(name),
		nil, &typ, (*byte)(unsafe.Pointer(&buf[0])), &n); nil != e {
		return "", e
	}

	return syscall.UTF16ToString(buf[0:int(n/2)]), nil
}

func (k *Key) SetString(name string, value string) error {
	buf := syscall.StringToUTF16(value)
	return winapi.RegSetValueEx(
		k.Handle, syscall.StringToUTF16Ptr(name),
		0, syscall.REG_SZ,
		(*byte)(unsafe.Pointer(&buf[0])), uint32(len(buf)*2))
}
