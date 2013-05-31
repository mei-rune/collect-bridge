package main

import (
	"bytes"
	"fmt"
	"os"
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
	kernel32                  = syscall.NewLazyDLL("kernel32.dll")
	CreateToolhelp32Snapshot  = kernel32.NewProc("CreateToolhelp32Snapshot")
	Process32First            = kernel32.NewProc("Process32First")
	Process32Next             = kernel32.NewProc("Process32Next")
	CloseHandle               = kernel32.NewProc("CloseHandle")
	QueryFullProcessImageName = kernel32.NewProc("QueryFullProcessImageName")
)

func nilString(b []byte) string {
	i := bytes.IndexByte(b, byte(0))
	if -1 == i {
		return string(b)
	}
	return string(b[0:i])
}

func queryFullProcessImageName(pid int) (string, error) {
	const da = syscall.STANDARD_RIGHTS_READ |
		syscall.PROCESS_QUERY_INFORMATION | syscall.SYNCHRONIZE
	h, e := syscall.OpenProcess(da, false, uint32(pid))
	if e != nil {
		return "", os.NewSyscallError("OpenProcess", e)
	}

	var fileName [1024]byte
	var size uintptr = 1024
	_, _, e = QueryFullProcessImageName.Call(uintptr(h), uintptr(0),
		uintptr(unsafe.Pointer(&fileName[0])), uintptr(unsafe.Pointer(&size)))
	if nil != e {
		return "", os.NewSyscallError("OpenProcess", e)
	}
	return string(fileName[:int(size)]), nil
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
		//rt, _, _ = QueryFullProcessImageName(proc.)
		//         BOOL WINAPI QueryFullProcessImageName(
		//   _In_     HANDLE hProcess,
		//   _In_     DWORD dwFlags,
		//   _Out_    LPTSTR lpExeName,
		//   _Inout_  PDWORD lpdwSize
		// );
		n, e := queryFullProcessImageName(int(proc.th32ProcessID))
		if nil != e {
			fmt.Println(e)
			fmt.Printf("%d\t%s\r\n", int(proc.th32ProcessID), nilString(proc.szExeFile[0:]))
		} else {
			fmt.Printf("%d\t%s\r\n", int(proc.th32ProcessID), n)
		}
	}
}
