package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"syscall"
	"unsafe"
)

const (
	MF_SOURCE_READER_FIRST_VIDEO_STREAM uintptr = 0xFFFFFFFC

	COINIT_APARTMENTTHREADED uintptr = 0x2

	MF_VERSION uintptr = 0x00020070

	NV12  uint32 = 0x3231564E
	YUY2  uint32 = 0x32595559
	RGB24 uint32 = 0x00000014
	RGB32 uint32 = 0x00000016
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

type mediaAttributesTable struct {
	QueryInterface     uintptr
	AddRef             uintptr
	Release            uintptr
	GetItem            uintptr
	GetItemType        uintptr
	CompareItem        uintptr
	Compare            uintptr
	GetUINT32          uintptr
	GetUINT64          uintptr
	GetDouble          uintptr
	GetGUID            uintptr
	GetStringLength    uintptr
	GetString          uintptr
	GetAllocatedString uintptr
	GetBlobSize        uintptr
	GetBlob            uintptr
	GetAllocatedBlob   uintptr
	GetUnknown         uintptr
	SetItem            uintptr
	DeleteItem         uintptr
	DeleteAllItems     uintptr
	SetUINT32          uintptr
	SetUINT64          uintptr
	SetDouble          uintptr
	SetGUID            uintptr
	SetString          uintptr
	SetBlob            uintptr
	SetUnknown         uintptr
	LockStore          uintptr
	UnlockStore        uintptr
	GetCount           uintptr
	GetItemByIndex     uintptr
	CopyAllItems       uintptr
}

type mediaAttributes struct {
	Table *mediaAttributesTable
}

func (a *mediaAttributes) release() {
	syscall.SyscallN(a.Table.Release, uintptr(unsafe.Pointer(a)))
}

func (a *mediaAttributes) setGUID(key *syscall.GUID, val *syscall.GUID) error {
	res, _, _ := syscall.SyscallN(
		a.Table.SetGUID,
		uintptr(unsafe.Pointer(a)),
		uintptr(unsafe.Pointer(key)),
		uintptr(unsafe.Pointer(val)),
	)

	if res != 0 {
		return fmt.Errorf("failed to call mediaAttributes.setGUID: 0x%X", res)
	}

	return nil
}

func (a *mediaAttributes) getGUID(key *syscall.GUID) (syscall.GUID, error) {
	var result syscall.GUID

	res, _, _ := syscall.SyscallN(
		a.Table.GetGUID,
		uintptr(unsafe.Pointer(a)),
		uintptr(unsafe.Pointer(key)),
		uintptr(unsafe.Pointer(&result)),
	)

	if res != 0 {
		return syscall.GUID{}, fmt.Errorf("failed to call mediaAttributes.getGUID: 0x%X", res)
	}

	return result, nil
}

func (a *mediaAttributes) getUInt64(key *syscall.GUID) (uint64, error) {
	var result uint64

	res, _, _ := syscall.SyscallN(
		a.Table.GetUINT64,
		uintptr(unsafe.Pointer(a)),
		uintptr(unsafe.Pointer(key)),
		uintptr(unsafe.Pointer(&result)),
	)

	if res != 0 {
		return 0, fmt.Errorf("failed to call mediaAttributes.getUInt64: 0x%X", res)
	}

	return result, nil
}

type mediaActivateTable struct {
	mediaAttributesTable
	ActivateObject uintptr
	ShutdownObject uintptr
	DetachObject   uintptr
}

type mediaActivate struct {
	Table *mediaActivateTable
}

func (a *mediaActivate) release() {
	syscall.SyscallN(a.Table.Release, uintptr(unsafe.Pointer(a)))
}

func (a *mediaActivate) activate(riid *syscall.GUID) (uintptr, error) {
	var source uintptr

	res, _, _ := syscall.SyscallN(
		a.Table.ActivateObject,
		uintptr(unsafe.Pointer(a)),
		uintptr(unsafe.Pointer(riid)),
		uintptr(unsafe.Pointer(&source)),
	)

	if res != 0 {
		return 0, fmt.Errorf("failed to call mediaActivate.activate: 0x%X", res)
	}

	return source, nil
}

type mediaReaderTable struct {
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

type mediaReader struct {
	Table *mediaReaderTable
}

func (a *mediaReader) release() {
	syscall.SyscallN(a.Table.Release, uintptr(unsafe.Pointer(a)))
}

func (a *mediaReader) setMedia(media *mediaType) error {
	res, _, _ := syscall.SyscallN(
		a.Table.SetCurrentMediaType,
		uintptr(unsafe.Pointer(a)),
		MF_SOURCE_READER_FIRST_VIDEO_STREAM,
		0,
		uintptr(unsafe.Pointer(media)),
	)

	if res != 0 {
		return fmt.Errorf("failed to call mediaReader.setMedia: 0x%X", res)
	}

	return nil
}

func (a *mediaReader) getMedia() (*mediaType, error) {
	var media *mediaType

	res, _, _ := syscall.SyscallN(
		a.Table.GetCurrentMediaType,
		uintptr(unsafe.Pointer(a)),
		MF_SOURCE_READER_FIRST_VIDEO_STREAM,
		uintptr(unsafe.Pointer(&media)),
	)

	if res != 0 {
		return nil, fmt.Errorf("failed to call mediaReader.getMedia: 0x%X", res)
	}

	return media, nil
}

func (a *mediaReader) readSample() (*mediaSample, uint32, error) {
	var stream uint32
	var flags uint32
	var time int64
	var sample *mediaSample

	res, _, _ := syscall.SyscallN(
		a.Table.ReadSample,
		uintptr(unsafe.Pointer(a)),
		MF_SOURCE_READER_FIRST_VIDEO_STREAM,
		0,
		uintptr(unsafe.Pointer(&stream)),
		uintptr(unsafe.Pointer(&flags)),
		uintptr(unsafe.Pointer(&time)),
		uintptr(unsafe.Pointer(&sample)),
	)

	if res != 0 {
		return nil, 0, fmt.Errorf("failed to call mediaReader.readSample: 0x%X", res)
	}

	return sample, flags, nil
}

type mediaSampleTable struct {
	mediaAttributesTable
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

type mediaSample struct {
	Table *mediaSampleTable
}

func (a *mediaSample) release() {
	syscall.SyscallN(a.Table.Release, uintptr(unsafe.Pointer(a)))
}

func (a *mediaSample) convertToBuffer() (*mediaBuffer, error) {
	var buffer *mediaBuffer

	res, _, _ := syscall.SyscallN(
		a.Table.ConvertToContiguousBuffer,
		uintptr(unsafe.Pointer(a)),
		uintptr(unsafe.Pointer(&buffer)),
	)

	if res != 0 {
		return nil, fmt.Errorf("failed to call mediaSample.convertToBuffer: 0x%X", res)
	}

	return buffer, nil
}

type mediaBufferTable struct {
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
	Table *mediaBufferTable
}

func (a *mediaBuffer) release() {
	syscall.SyscallN(a.Table.Release, uintptr(unsafe.Pointer(a)))
}

func (a *mediaBuffer) lock() (uintptr, uint32, error) {
	var data uintptr
	var maxLen uint32
	var curLen uint32

	res, _, _ := syscall.SyscallN(
		a.Table.Lock,
		uintptr(unsafe.Pointer(a)),
		uintptr(unsafe.Pointer(&data)),
		uintptr(unsafe.Pointer(&maxLen)),
		uintptr(unsafe.Pointer(&curLen)),
	)

	if res != 0 {
		return 0, 0, fmt.Errorf("failed to call mediaBuffer.lock: 0x%X", res)
	}

	return data, curLen, nil
}

func (a *mediaBuffer) unlock() {
	syscall.SyscallN(a.Table.Unlock, uintptr(unsafe.Pointer(a)))
}

type mediaTypeTable struct {
	mediaAttributesTable
	GetMajorType       uintptr
	IsCompressedFormat uintptr
	IsEqual            uintptr
	GetRepresentation  uintptr
	FreeRepresentation uintptr
}

type mediaType struct {
	Table *mediaTypeTable
}

func (a *mediaType) release() {
	syscall.SyscallN(a.Table.Release, uintptr(unsafe.Pointer(a)))
}

func (a *mediaType) attributes() *mediaAttributes {
	return (*mediaAttributes)(unsafe.Pointer(a))
}

func CaptureWebcam() (*bytes.Buffer, error) {
	initCOM.Call(0, COINIT_APARTMENTTHREADED)
	defer uninitCOM.Call()

	res, _, err := mediaStart.Call(MF_VERSION, 0)
	if res != 0 {
		return nil, err
	}

	defer mediaStop.Call()

	var atributes *mediaAttributes

	res, _, err = mediaCreateAttributes.Call(uintptr(unsafe.Pointer(&atributes)), 1)
	if res != 0 {
		return nil, err
	}

	defer atributes.release()

	err = atributes.setGUID(&attributeSource, &attributeSourceVideo)
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
		return nil, errors.New("no cameras found")
	}

	activate := (*mediaActivate)(unsafe.Pointer(*(*uintptr)(unsafe.Pointer(devices))))
	defer activate.release()

	source, err := activate.activate(&mediaSource)
	if err != nil {
		return nil, err
	}

	defer func() {
		if source != 0 {
			table := *(**[3]uintptr)(unsafe.Pointer(source))
			syscall.SyscallN(table[2], source)
		}
	}()

	var reader *mediaReader
	res, _, err = mediaCreateReader.Call(
		source,
		0,
		uintptr(unsafe.Pointer(&reader)),
	)

	if res != 0 {
		return nil, err
	}

	defer reader.release()

	var mediaType *mediaType

	res, _, _ = mediaCreateType.Call(uintptr(unsafe.Pointer(&mediaType)))
	if res == 0 {
		defer mediaType.release()

		mediaType.attributes().setGUID(&mediaMajortype, &videoMedia)
		mediaType.attributes().setGUID(&mediaSubtype, &videoFormat)

		reader.setMedia(mediaType)
	}

	var sample *mediaSample
	for sample == nil {
		theSample, flags, err := reader.readSample()
		if err != nil {
			return nil, err
		}

		if flags&0x4 != 0 {
			return nil, errors.New("failed to read sample")
		}

		sample = theSample
	}

	defer sample.release()

	gottenMedia, err := reader.getMedia()
	if err != nil {
		return nil, err
	}

	defer gottenMedia.release()

	subtype, err := gottenMedia.attributes().getGUID(&mediaSubtype)
	if err != nil {
		return nil, err
	}

	frameSize, err := gottenMedia.attributes().getUInt64(&frameSize)
	if err != nil {
		return nil, err
	}

	width := int(frameSize >> 32)
	height := int(frameSize & 0xFFFFFFFF)

	buffer, err := sample.convertToBuffer()
	if err != nil {
		return nil, err
	}

	defer buffer.release()

	data, length, err := buffer.lock()
	if err != nil {
		return nil, err
	}

	defer buffer.unlock()

	rawData := make([]byte, length)
	copy(rawData, unsafe.Slice((*byte)(unsafe.Pointer(data)), length))

	var pixels []byte

	switch subtype.Data1 {
	case NV12:

		pixels = fromNV12(rawData, width, height)
	case YUY2:

		pixels = fromYUY2(rawData, width, height)
	case RGB24:

		pixels = rawData
	case RGB32:

		pixels = make([]byte, width*height*3)
		for i := range width * height {
			pixels[i*3+0] = rawData[i*4+0]
			pixels[i*3+1] = rawData[i*4+1]
			pixels[i*3+2] = rawData[i*4+2]
		}

	default:
		return nil, fmt.Errorf("invalid format: 0x%08X", subtype.Data1)

	}

	return convertToPNG(pixels, width, height)
}

func clamp(num float64) float64 {
	if num < 0 {
		return 0
	}

	if num > 255 {
		return 255
	}

	return num
}

func fromNV12(pixels []byte, width int, height int) []byte {
	newPixels := make([]byte, width*height*3)

	lumma := pixels[:width*height]
	chroma := pixels[width*height:]

	for y := range height {
		for x := range width {
			block := (y/2)*width + (x/2)*2

			blue := float64(chroma[block]) - 128
			red := float64(chroma[block+1]) - 128
			alpha := float64(lumma[y*width+x])

			i := (y*width + x) * 3

			newPixels[i+0] = byte(clamp(alpha + 1.370705*red))
			newPixels[i+1] = byte(clamp(alpha - 0.698001*red - 0.337633*blue))
			newPixels[i+2] = byte(clamp(alpha + 1.732446*blue))
		}
	}

	return newPixels
}

func fromYUY2(pixels []byte, width int, height int) []byte {
	newPixels := make([]byte, width*height*3)

	for y := range height {
		for x := 0; x < width; x += 2 {
			base := (y*width + x) * 2

			leftAlpha := float64(pixels[base+0])
			blue := float64(pixels[base+1]) - 128

			rightAlpha := float64(pixels[base+2])
			red := float64(pixels[base+3]) - 128

			i0 := (y*width + x) * 3

			newPixels[i0+0] = byte(clamp(leftAlpha + 1.370705*red))
			newPixels[i0+1] = byte(clamp(leftAlpha - 0.698001*red - 0.337633*blue))
			newPixels[i0+2] = byte(clamp(leftAlpha + 1.732446*blue))

			i1 := i0 + 3

			if x+1 < width {
				newPixels[i1+0] = byte(clamp(rightAlpha + 1.370705*red))
				newPixels[i1+1] = byte(clamp(rightAlpha - 0.698001*red - 0.337633*blue))
				newPixels[i1+2] = byte(clamp(rightAlpha + 1.732446*blue))
			}
		}
	}

	return newPixels
}

func convertToPNG(pixels []byte, width int, height int) (*bytes.Buffer, error) {
	picture := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := range height {
		for x := range width {
			i := (y*width + x) * 3

			picture.Pix[(y*width+x)*4+0] = pixels[i+0]
			picture.Pix[(y*width+x)*4+1] = pixels[i+1]
			picture.Pix[(y*width+x)*4+2] = pixels[i+2]
			picture.Pix[(y*width+x)*4+3] = 255
		}
	}

	var buffer bytes.Buffer

	err := png.Encode(&buffer, picture)
	if err != nil {
		return nil, err
	}

	return &buffer, nil
}
