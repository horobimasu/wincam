package wincam

import (
	"fmt"
	"syscall"
	"unsafe"
)

type MediaActivateVTable struct {
	MediaAttributesVTable
	ActivateObject uintptr
	ShutdownObject uintptr
	DetachObject   uintptr
}

type MediaActivate struct {
	VTable *MediaActivateVTable
}

func (activate *MediaActivate) Release() {
	syscall.SyscallN(activate.VTable.Release, uintptr(unsafe.Pointer(activate)))
}

func (activate *MediaActivate) Activate(key *syscall.GUID) (uintptr, error) {
	var source uintptr

	res, _, _ := syscall.SyscallN(
		activate.VTable.ActivateObject,
		uintptr(unsafe.Pointer(activate)),
		uintptr(unsafe.Pointer(key)),
		uintptr(unsafe.Pointer(&source)),
	)

	if res != 0 {
		return 0, fmt.Errorf("failed to activate media object: 0x%X", res)
	}

	return source, nil
}
