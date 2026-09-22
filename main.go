package win_cam

import (
	"bytes"
	"errors"
	"fmt"
	"syscall"
	"unsafe"
)

const (
	_COINIT_APARTMENTTHREADED uintptr = 0x2
	_MF_VERSION               uintptr = 0x00020070
)

const (
	_NV12  uint32 = 0x3231564E
	_YUY2  uint32 = 0x32595559
	_RGB24 uint32 = 0x00000014
	_RGB32 uint32 = 0x00000016
)

var (
	ole32  = syscall.NewLazyDLL("ole32.dll")
	mfplat = syscall.NewLazyDLL("mfplat.dll")
	mf     = syscall.NewLazyDLL("mf.dll")
	mfrw   = syscall.NewLazyDLL("mfreadwrite.dll")

	initCOM   = ole32.NewProc("CoInitialize")
	uninitCOM = ole32.NewProc("CoUninitialize")
	createCOM = ole32.NewProc("CoCreateInstance")

	mediaStart            = mfplat.NewProc("MFStartup")
	mediaStop             = mfplat.NewProc("MFShutdown")
	mediaCreateAttributes = mfplat.NewProc("MFCreateAttributes")
	mediaCreateType       = mfplat.NewProc("MFCreateMediaType")

	mediaEnumerateDevices = mf.NewProc("MFEnumDeviceSources")

	mediaCreateReader = mfrw.NewProc("MFCreateSourceReaderFromMediaSource")
)

var (
	attributeSource = syscall.GUID{
		Data1: 0xC60AC5FE,
		Data2: 0x252A,
		Data3: 0x478F,
		Data4: [8]byte{0xA0, 0xEF, 0xBC, 0x8F, 0xA5, 0xF7, 0xCA, 0xD3},
	}

	attributeSourceVideo = syscall.GUID{
		Data1: 0x8AC3587A,
		Data2: 0x4AE7,
		Data3: 0x42D8,
		Data4: [8]byte{0x99, 0xE0, 0x0A, 0x60, 0x13, 0xEE, 0xF9, 0x0F},
	}

	mediaSource = syscall.GUID{
		Data1: 0x279A808D,
		Data2: 0xAEC7,
		Data3: 0x40C8,
		Data4: [8]byte{0x9C, 0x6B, 0xA6, 0xB4, 0x92, 0xC7, 0x8A, 0x66},
	}

	mediaMajortype = syscall.GUID{
		Data1: 0x48ABA7A7,
		Data2: 0xA5A2,
		Data3: 0x4FAB,
		Data4: [8]byte{0xA2, 0x3B, 0x4A, 0x40, 0x5B, 0x93, 0xA3, 0x18},
	}

	mediaSubtype = syscall.GUID{
		Data1: 0xF7E34C9A,
		Data2: 0x42E8,
		Data3: 0x4714,
		Data4: [8]byte{0xB7, 0x4B, 0xCB, 0x29, 0xD7, 0x2C, 0x35, 0xE5},
	}

	videoMedia = syscall.GUID{
		Data1: 0x73646976,
		Data2: 0x0000,
		Data3: 0x0010,
		Data4: [8]byte{0x80, 0x00, 0x00, 0xAA, 0x00, 0x38, 0x9B, 0x71},
	}

	videoFormat = syscall.GUID{
		Data1: 0x00000014,
		Data2: 0x0000,
		Data3: 0x0010,
		Data4: [8]byte{0x80, 0x00, 0x00, 0xAA, 0x00, 0x38, 0x9B, 0x71},
	}

	frameSize = syscall.GUID{
		Data1: 0x1652C33D,
		Data2: 0xD6B2,
		Data3: 0x4012,
		Data4: [8]byte{0xB8, 0x34, 0x72, 0x03, 0x08, 0x49, 0xA3, 0x7D},
	}
)

var (
	ErrNoCamsFound        = errors.New("no cameras found")
	ErrFailedToReadSample = errors.New("failed to read sample")
	ErrInvalidDataFormat  = errors.New("invalid data format")
)

var (
	lastInvalidData           = []byte{}
	lastInvalidDataFormatCode = ""
)

func CaptureWebcam() (*bytes.Buffer, error) {
	initCOM.Call(0, _COINIT_APARTMENTTHREADED)
	defer uninitCOM.Call()

	res, _, err := mediaStart.Call(_MF_VERSION, 0)
	if res != 0 {
		return nil, err
	}

	defer mediaStop.Call()

	var atributes *MediaAttributes

	res, _, err = mediaCreateAttributes.Call(uintptr(unsafe.Pointer(&atributes)), 1)
	if res != 0 {
		return nil, err
	}

	defer atributes.Release()

	err = atributes.SetGUID(&attributeSource, &attributeSourceVideo)
	if err != nil {
		return nil, err
	}

	var devices uintptr
	var count uint32

	res, _, err = mediaEnumerateDevices.Call(
		uintptr(unsafe.Pointer(atributes)),
		uintptr(unsafe.Pointer(&devices)),
		uintptr(unsafe.Pointer(&count)),
	)

	if res != 0 {
		return nil, err
	}

	if count == 0 {
		return nil, ErrNoCamsFound
	}

	activate := (*MediaActivate)(unsafe.Pointer(*(*uintptr)(unsafe.Pointer(devices))))
	defer activate.Release()

	source, err := activate.Activate(&mediaSource)
	if err != nil {
		return nil, err
	}

	defer func() {
		if source != 0 {
			table := *(**[3]uintptr)(unsafe.Pointer(source))
			syscall.SyscallN(table[2], source)
		}
	}()

	var reader *MediaReader
	res, _, err = mediaCreateReader.Call(
		source,
		0,
		uintptr(unsafe.Pointer(&reader)),
	)

	if res != 0 {
		return nil, err
	}

	defer reader.Release()

	var mediaType *MediaType

	res, _, _ = mediaCreateType.Call(uintptr(unsafe.Pointer(&mediaType)))
	if res == 0 {
		defer mediaType.release()

		mediaType.attributes().SetGUID(&mediaMajortype, &videoMedia)
		mediaType.attributes().SetGUID(&mediaSubtype, &videoFormat)

		reader.SetMedia(mediaType)
	}

	var sample *MediaSample
	for sample == nil {
		theSample, flags, err := reader.ReadMedia()
		if err != nil {
			return nil, err
		}

		if flags&0x4 != 0 {
			return nil, ErrFailedToReadSample
		}

		sample = theSample
	}

	defer sample.Release()

	gottenMedia, err := reader.GetMedia()
	if err != nil {
		return nil, err
	}

	defer gottenMedia.release()

	subtype, err := gottenMedia.attributes().GetGUID(&mediaSubtype)
	if err != nil {
		return nil, err
	}

	frameSize, err := gottenMedia.attributes().GetUInt64(&frameSize)
	if err != nil {
		return nil, err
	}

	width := int(frameSize >> 32)
	height := int(frameSize & 0xFFFFFFFF)

	buffer, err := sample.ConvertBuffer()
	if err != nil {
		return nil, err
	}

	defer buffer.Release()

	data, length, err := buffer.Lock()
	if err != nil {
		return nil, err
	}

	defer buffer.Unlock()

	rawData := make([]byte, length)
	copy(rawData, unsafe.Slice((*byte)(unsafe.Pointer(data)), length))

	var pixels []byte

	switch subtype.Data1 {
	case _NV12:
		pixels = ParseNV12(rawData, width, height)

	case _YUY2:
		pixels = ParseYUY2(rawData, width, height)

	case _RGB24:
		pixels = rawData

	case _RGB32:
		pixels = make([]byte, width*height*3)
		for i := range width * height {
			pixels[i*3+0] = rawData[i*4+0]
			pixels[i*3+1] = rawData[i*4+1]
			pixels[i*3+2] = rawData[i*4+2]
		}

	default:
		lastInvalidData = rawData
		lastInvalidDataFormatCode = fmt.Sprintf("0x%08X", subtype.Data1)

		return nil, ErrInvalidDataFormat
	}

	return ConvertToPNG(pixels, width, height)
}

// if the raw data format is not NV12, YUY2, RGB24 or RGB32
// you can use this to get it if you want
func GetLastInvalidFormatErrorData() (string, []byte) {
	return lastInvalidDataFormatCode, lastInvalidData
}
