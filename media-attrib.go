package win_cam

import (
	"fmt"
	"syscall"
	"unsafe"
)

type MediaAttributesVTable struct {
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

type MediaAttributes struct {
	VTable *MediaAttributesVTable
}

func (attribs *MediaAttributes) Release() {
	syscall.SyscallN(attribs.VTable.Release, uintptr(unsafe.Pointer(attribs)))
}

func (attribs *MediaAttributes) SetGUID(key *syscall.GUID, val *syscall.GUID) error {
	res, _, _ := syscall.SyscallN(
		attribs.VTable.SetGUID,
		uintptr(unsafe.Pointer(attribs)),
		uintptr(unsafe.Pointer(key)),
		uintptr(unsafe.Pointer(val)),
	)

	if res != 0 {
		return fmt.Errorf("failed to set media attribs guid: 0x%X", res)
	}

	return nil
}

func (attribs *MediaAttributes) GetGUID(key *syscall.GUID) (syscall.GUID, error) {
	var result syscall.GUID

	res, _, _ := syscall.SyscallN(
		attribs.VTable.GetGUID,
		uintptr(unsafe.Pointer(attribs)),
		uintptr(unsafe.Pointer(key)),
		uintptr(unsafe.Pointer(&result)),
	)

	if res != 0 {
		return syscall.GUID{}, fmt.Errorf("failed to get media attribs guid: 0x%X", res)
	}

	return result, nil
}

func (attribs *MediaAttributes) GetUInt64(key *syscall.GUID) (uint64, error) {
	var result uint64

	res, _, _ := syscall.SyscallN(
		attribs.VTable.GetUINT64,
		uintptr(unsafe.Pointer(attribs)),
		uintptr(unsafe.Pointer(key)),
		uintptr(unsafe.Pointer(&result)),
	)

	if res != 0 {
		return 0, fmt.Errorf("failed to get media attribs uint64: 0x%X", res)
	}

	return result, nil
}
