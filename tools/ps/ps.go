package main

import (
	"bytes"
	"fmt"
	"syscall"
	"unsafe"
)

type ulong int32
type ulong_ptr uintptr

type PROCESSENTRY32 struct {
	dwSize              ulong
	cntUsage            ulong
	th32ProcessID       ulong
	th32DefaultHeapID   ulong_ptr
	th32ModuleID        ulong
	cntThreads          ulong
	th32ParentProcessID ulong
	pcPriClassBase      ulong
	dwFlags             ulong
	szExeFile           [412]byte
}

var (
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	CreateToolhelp32Snapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
	Process32First           = kernel32.NewProc("Process32First")
	Process32Next            = kernel32.NewProc("Process32Next")
	CloseHandle              = kernel32.NewProc("CloseHandle")
)

func nilString(b []byte) string {
	i := bytes.IndexByte(b, byte(0))
	if -1 == i {
		return string(b)
	}
	return string(b[0:i])
}

func main() {
	pHandle, _, _ := CreateToolhelp32Snapshot.Call(uintptr(0x2), uintptr(0x0))
	if int(pHandle) == -1 {
		return
	}
	defer CloseHandle.Call(pHandle)

	var proc PROCESSENTRY32
	h, p, rt := uintptr(pHandle), uintptr(unsafe.Pointer(&proc)), uintptr(0)
	proc.dwSize = ulong(unsafe.Sizeof(proc))

	fmt.Println("ProcessID\tProcessName\r\n")
	for rt, _, _ = Process32First.Call(h, p); 0 != int(rt); rt, _, _ = Process32Next.Call(h, p) {
		fmt.Printf("%d\t%s\r\n", int(proc.th32ProcessID), nilString(proc.szExeFile[0:]))
	}
}
