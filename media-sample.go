package win_cam

import (
	"fmt"
	"syscall"
	"unsafe"
)

type MediaSampleVTable struct {
	MediaAttributesVTable
	GetSampleFlags            uintptr
	SetSampleFlags            uintptr
	GetSampleTime             uintptr
	SetSampleTime             uintptr
	GetSampleDuration         uintptr
	SetSampleDuration         uintptr
	GetBufferCount            uintptr
	GetBufferByIndex          uintptr
	ConvertToContiguousBuffer uintptr
	AddBuffer                 uintptr
	RemoveBufferByIndex       uintptr
	RemoveAllBuffers          uintptr
	GetTotalLength            uintptr
	CopyToBuffer              uintptr
}

type MediaSample struct {
	VTable *MediaSampleVTable
}

func (sample *MediaSample) Release() {
	syscall.SyscallN(sample.VTable.Release, uintptr(unsafe.Pointer(sample)))
}

func (sample *MediaSample) ConvertBuffer() (*mediaBuffer, error) {
	var buffer *mediaBuffer

	res, _, _ := syscall.SyscallN(
		sample.VTable.ConvertToContiguousBuffer,
		uintptr(unsafe.Pointer(sample)),
		uintptr(unsafe.Pointer(&buffer)),
	)

	if res != 0 {
		return nil, fmt.Errorf("failed to covert media sample: 0x%X", res)
	}

	return buffer, nil
}
