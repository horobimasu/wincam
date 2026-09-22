package win_cam

import (
	"fmt"
	"syscall"
	"unsafe"
)

type MediaBufferVTable struct {
	QueryInterface   uintptr
	AddRef           uintptr
	Release          uintptr
	Lock             uintptr
	Unlock           uintptr
	GetCurrentLength uintptr
	SetCurrentLength uintptr
	GetMaxLength     uintptr
}

type mediaBuffer struct {
	VTable *MediaBufferVTable
}

func (buffer *mediaBuffer) Release() {
	syscall.SyscallN(buffer.VTable.Release, uintptr(unsafe.Pointer(buffer)))
}

func (buffer *mediaBuffer) Lock() (uintptr, uint32, error) {
	var data uintptr
	var maxLen uint32
	var curLen uint32

	res, _, _ := syscall.SyscallN(
		buffer.VTable.Lock,
		uintptr(unsafe.Pointer(buffer)),
		uintptr(unsafe.Pointer(&data)),
		uintptr(unsafe.Pointer(&maxLen)),
		uintptr(unsafe.Pointer(&curLen)),
	)

	if res != 0 {
		return 0, 0, fmt.Errorf("failed to lock media buffer: 0x%X", res)
	}

	return data, curLen, nil
}

func (buffer *mediaBuffer) Unlock() {
	syscall.SyscallN(buffer.VTable.Unlock, uintptr(unsafe.Pointer(buffer)))
}
