//go:build windows
// +build windows

package activity

import (
	"syscall"
	"unsafe"
)

var (
	user32                       = syscall.NewLazyDLL("user32.dll")
	kernel32                     = syscall.NewLazyDLL("kernel32.dll")
	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW     = user32.NewProc("GetWindowTextLengthW")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procGetLastInputInfo         = user32.NewProc("GetLastInputInfo")
	procGetTickCount64           = kernel32.NewProc("GetTickCount64")
	procOpenProcess              = kernel32.NewProc("OpenProcess")
	procCloseHandle              = kernel32.NewProc("CloseHandle")
	psapi                        = syscall.NewLazyDLL("psapi.dll")
	procGetModuleBaseNameW       = psapi.NewProc("GetModuleBaseNameW")
)

const (
	PROCESS_QUERY_INFORMATION = 0x0400
	PROCESS_VM_READ           = 0x0010
)

type lastInputInfo struct {
	cbSize uint32
	dwTime uint32
}

func GetActiveWindow() WindowInfo {
	info := WindowInfo{}

	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return info
	}

	var pid uint32
	procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	info.ProcessID = int32(pid)

	length, _, _ := procGetWindowTextLengthW.Call(hwnd)
	if length > 0 {
		buf := make([]uint16, length+1)
		procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), length+1)
		info.WindowTitle = syscall.UTF16ToString(buf)
	}

	hProcess, _, _ := procOpenProcess.Call(PROCESS_QUERY_INFORMATION|PROCESS_VM_READ, 0, uintptr(pid))
	if hProcess != 0 {
		defer procCloseHandle.Call(hProcess)
		buf := make([]uint16, 256)
		n, _, _ := procGetModuleBaseNameW.Call(hProcess, 0, uintptr(unsafe.Pointer(&buf[0])), 256)
		if n > 0 {
			info.AppName = syscall.UTF16ToString(buf[:n])
			info.BundleID = info.AppName
		}
	}

	return info
}

func GetIdleTimeMs() int64 {
	lii := lastInputInfo{cbSize: uint32(unsafe.Sizeof(lastInputInfo{}))}
	ret, _, _ := procGetLastInputInfo.Call(uintptr(unsafe.Pointer(&lii)))
	if ret == 0 {
		return 0
	}

	tickCount, _, _ := procGetTickCount64.Call()
	return int64(tickCount) - int64(lii.dwTime)
}
