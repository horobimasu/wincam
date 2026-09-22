package wincam

import (
	"syscall"
	"unsafe"
)

type MediaTypeVTable struct {
	MediaAttributesVTable
	GetMajorType       uintptr
	IsCompressedFormat uintptr
	IsEqual            uintptr
	GetRepresentation  uintptr
	FreeRepresentation uintptr
}

type MediaType struct {
	VTable *MediaTypeVTable
}

func (media *MediaType) release() {
	syscall.SyscallN(media.VTable.Release, uintptr(unsafe.Pointer(media)))
}

func (media *MediaType) attributes() *MediaAttributes {
	return (*MediaAttributes)(unsafe.Pointer(media))
}
