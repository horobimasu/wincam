package win_cam

import (
	"fmt"
	"syscall"
	"unsafe"
)

const _MF_SOURCE_READER_FIRST_VIDEO_STREAM uintptr = 0xFFFFFFFC

type MediaReaderVTable struct {
	QueryInterface           uintptr
	AddRef                   uintptr
	Release                  uintptr
	GetStreamSelection       uintptr
	SetStreamSelection       uintptr
	GetNativeMediaType       uintptr
	GetCurrentMediaType      uintptr
	SetCurrentMediaType      uintptr
	SetCurrentPosition       uintptr
	ReadSample               uintptr
	Flush                    uintptr
	GetServiceForStream      uintptr
	GetPresentationAttribute uintptr
}

type MediaReader struct {
	VTable *MediaReaderVTable
}

func (reader *MediaReader) Release() {
	syscall.SyscallN(reader.VTable.Release, uintptr(unsafe.Pointer(reader)))
}

func (reader *MediaReader) SetMedia(media *MediaType) error {
	res, _, _ := syscall.SyscallN(
		reader.VTable.SetCurrentMediaType,
		uintptr(unsafe.Pointer(reader)),
		_MF_SOURCE_READER_FIRST_VIDEO_STREAM,
		0,
		uintptr(unsafe.Pointer(media)),
	)

	if res != 0 {
		return fmt.Errorf("failed to set media: 0x%X", res)
	}

	return nil
}

func (reader *MediaReader) GetMedia() (*MediaType, error) {
	var media *MediaType

	res, _, _ := syscall.SyscallN(
		reader.VTable.GetCurrentMediaType,
		uintptr(unsafe.Pointer(reader)),
		_MF_SOURCE_READER_FIRST_VIDEO_STREAM,
		uintptr(unsafe.Pointer(&media)),
	)

	if res != 0 {
		return nil, fmt.Errorf("failed to get media: 0x%X", res)
	}

	return media, nil
}

func (reader *MediaReader) ReadMedia() (*MediaSample, uint32, error) {
	var stream uint32
	var flags uint32
	var time int64
	var sample *MediaSample

	res, _, _ := syscall.SyscallN(
		reader.VTable.ReadSample,
		uintptr(unsafe.Pointer(reader)),
		_MF_SOURCE_READER_FIRST_VIDEO_STREAM,
		0,
		uintptr(unsafe.Pointer(&stream)),
		uintptr(unsafe.Pointer(&flags)),
		uintptr(unsafe.Pointer(&time)),
		uintptr(unsafe.Pointer(&sample)),
	)

	if res != 0 {
		return nil, 0, fmt.Errorf("failed to read media: 0x%X", res)
	}

	return sample, flags, nil
}
