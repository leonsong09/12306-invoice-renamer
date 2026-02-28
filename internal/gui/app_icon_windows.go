//go:build windows

package gui

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	iconSizeLarge = 32
	iconSizeSmall = 16
)

const (
	iconColorTransparent uint32 = 0x00000000
	iconColorFill        uint32 = 0xFF2B6CB0
	iconColorBorder      uint32 = 0xFF1A365D
	iconColorPerforation uint32 = 0xFFF7FAFC
)

const (
	iconBorderThickness = 1
)

const (
	winBoolTrue = 1
)

const (
	dibRedMask   uint32 = 0x00FF0000
	dibGreenMask uint32 = 0x0000FF00
	dibBlueMask  uint32 = 0x000000FF
	dibAlphaMask uint32 = 0xFF000000
)

const (
	maskRowAlignBits = 16
	bitsPerByte      = 8
)

const (
	dibPlanes     uint16 = 1
	dibBitCount32 uint16 = 32

	monoBitmapPlanes   uintptr = 1
	monoBitmapBitCount uintptr = 1

	minTicketPadding       = 2
	minPerforationStep     = 2
	perforationEdgePadding = 2
)

func (a *app) initIcons(hwnd syscall.Handle) error {
	big, err := createAppIcon(iconSizeLarge)
	if err != nil {
		return err
	}
	small, err := createAppIcon(iconSizeSmall)
	if err != nil {
		destroyIcon(big)
		return err
	}

	a.iconBig = big
	a.iconSmall = small
	sendMessage(hwnd, wmSetIcon, uintptr(iconBig), uintptr(big))
	sendMessage(hwnd, wmSetIcon, uintptr(iconSmall), uintptr(small))
	return nil
}

func (a *app) onDestroy() {
	destroyIcon(a.iconBig)
	destroyIcon(a.iconSmall)
	a.iconBig = 0
	a.iconSmall = 0
}

func destroyIcon(icon syscall.Handle) {
	if icon == 0 {
		return
	}
	procDestroyIcon.Call(uintptr(icon))
}

func createAppIcon(size int) (syscall.Handle, error) {
	colorBmp, bits, err := createARGBDIBSection(size)
	if err != nil {
		return 0, err
	}
	pixels := unsafe.Slice((*uint32)(bits), size*size)
	drawTicketIcon(pixels, size)

	maskBits := buildMaskBitsFromAlpha(pixels, size)
	maskBmp, err := createMonochromeBitmap(size, maskBits)
	if err != nil {
		deleteGdiObject(colorBmp)
		return 0, err
	}

	icon, err := createIconFromBitmaps(colorBmp, maskBmp)
	deleteGdiObject(maskBmp)
	deleteGdiObject(colorBmp)
	return icon, err
}

func createARGBDIBSection(size int) (syscall.Handle, unsafe.Pointer, error) {
	header := bitmapV5Header{
		bV5Size:        uint32(unsafe.Sizeof(bitmapV5Header{})),
		bV5Width:       int32(size),
		bV5Height:      -int32(size),
		bV5Planes:      dibPlanes,
		bV5BitCount:    dibBitCount32,
		bV5Compression: biBitFields,
		bV5RedMask:     dibRedMask,
		bV5GreenMask:   dibGreenMask,
		bV5BlueMask:    dibBlueMask,
		bV5AlphaMask:   dibAlphaMask,
		bV5CSType:      lcsSrgb,
		bV5Intent:      lcsGmImages,
	}

	var bits unsafe.Pointer
	r1, _, err := procCreateDIBSection.Call(
		0,
		uintptr(unsafe.Pointer(&header)),
		dibRGBColors,
		uintptr(unsafe.Pointer(&bits)),
		0,
		0,
	)
	if r1 == 0 || bits == nil {
		return 0, nil, fmt.Errorf("CreateDIBSection failed: %w", err)
	}
	return syscall.Handle(r1), bits, nil
}

func createMonochromeBitmap(size int, bits []byte) (syscall.Handle, error) {
	if len(bits) == 0 {
		return 0, fmt.Errorf("mask bits is empty")
	}
	r1, _, err := procCreateBitmap.Call(
		uintptr(size),
		uintptr(size),
		monoBitmapPlanes,
		monoBitmapBitCount,
		uintptr(unsafe.Pointer(&bits[0])),
	)
	if r1 == 0 {
		return 0, fmt.Errorf("CreateBitmap failed: %w", err)
	}
	return syscall.Handle(r1), nil
}

func createIconFromBitmaps(colorBmp syscall.Handle, maskBmp syscall.Handle) (syscall.Handle, error) {
	ii := iconInfo{
		fIcon:    winBoolTrue,
		hbmMask:  maskBmp,
		hbmColor: colorBmp,
	}
	r1, _, err := procCreateIconIndirect.Call(uintptr(unsafe.Pointer(&ii)))
	if r1 == 0 {
		return 0, fmt.Errorf("CreateIconIndirect failed: %w", err)
	}
	return syscall.Handle(r1), nil
}

func deleteGdiObject(obj syscall.Handle) {
	if obj == 0 {
		return
	}
	procDeleteObject.Call(uintptr(obj))
}

func buildMaskBitsFromAlpha(pixels []uint32, size int) []byte {
	stride := ((size + (maskRowAlignBits - 1)) / maskRowAlignBits) * (maskRowAlignBits / bitsPerByte)
	mask := make([]byte, stride*size)
	for y := 0; y < size; y++ {
		rowOff := y * stride
		for x := 0; x < size; x++ {
			if (pixels[(y*size)+x]>>24)&0xFF != 0 {
				continue
			}
			mask[rowOff+(x/bitsPerByte)] |= 1 << ((bitsPerByte - 1) - (uint(x) % bitsPerByte))
		}
	}
	return mask
}

func drawTicketIcon(pixels []uint32, size int) {
	left, top, right, bottom, radius := ticketBounds(size)
	innerRadius := radius - iconBorderThickness
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if !insideRoundedRect(x, y, left, top, right, bottom, radius) {
				pixels[(y*size)+x] = iconColorTransparent
				continue
			}
			if insideRoundedRect(x, y, left+iconBorderThickness, top+iconBorderThickness, right-iconBorderThickness, bottom-iconBorderThickness, innerRadius) {
				pixels[(y*size)+x] = iconColorFill
				continue
			}
			pixels[(y*size)+x] = iconColorBorder
		}
	}
	drawPerforation(pixels, size, left, top, right, bottom)
}

func ticketBounds(size int) (left int, top int, right int, bottom int, radius int) {
	margin := maxInt(minTicketPadding, size/8)
	left = margin
	top = margin
	right = size - margin - 1
	bottom = size - margin - 1
	radius = maxInt(minTicketPadding, size/6)
	return left, top, right, bottom, radius
}

func drawPerforation(pixels []uint32, size int, left int, top int, right int, bottom int) {
	x := (left + right) / 2
	step := maxInt(minPerforationStep, size/8)
	for y := top + perforationEdgePadding; y <= bottom-perforationEdgePadding; y += step {
		pixels[(y*size)+x] = iconColorPerforation
	}
}

func insideRoundedRect(x int, y int, left int, top int, right int, bottom int, radius int) bool {
	if x < left || x > right || y < top || y > bottom {
		return false
	}
	if radius <= 0 {
		return true
	}
	if x >= left+radius && x <= right-radius {
		return true
	}
	if y >= top+radius && y <= bottom-radius {
		return true
	}

	cx := left + radius
	cy := top + radius
	if x > (left+right)/2 {
		cx = right - radius
	}
	if y > (top+bottom)/2 {
		cy = bottom - radius
	}
	dx := x - cx
	dy := y - cy
	return (dx*dx)+(dy*dy) <= radius*radius
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
