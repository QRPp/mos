//
// Copyright (c) 2014-2019 Cesanta Software Limited
// All rights reserved
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// +build linux darwin
// +build !old_libftdi

package cc3200

// Implements FTDI chip bit-bang pin access using libftdi.
// Linux (Debian/Ubuntu): apt-get install libftdi1-dev

/*
#include <ftdi.h>

#cgo pkg-config: libftdi1
#cgo CFLAGS: -I/usr/local/Cellar/libftdi/1.5_1/include/libftdi1
*/
import "C"

import (
	"unsafe"

	"github.com/juju/errors"
)

type ftdiLibWrapper struct {
	ctx  *C.struct_ftdi_context
	open bool
}

const (
	channelA         = 0x01
	modeAsyncBitBang = 0x01
)

func openFTDI(vendor, product int, serial string) (FTDI, error) {
	ctx, err := C.ftdi_new()
	if ctx == nil {
		return nil, err
	}
	f := &ftdiLibWrapper{ctx: ctx}
	if e := C.ftdi_set_interface(ctx, C.enum_ftdi_interface(channelA)); e < 0 {
		return nil, errors.Errorf("failed to set interface: %d", e)
	}

	var snb []byte // To keep the slice snPtr points to alive.
	var snPtr *C.char
	if serial != "" {
		snb = []byte(serial)
		snb = append(snb, 0) // NUL-terminate
		snPtr = (*C.char)(unsafe.Pointer(&snb[0]))
	}
	if e := C.ftdi_usb_open_desc(
		f.ctx, C.int(vendor), C.int(product), (*C.char)(nil) /* description */, snPtr); e < 0 {
		return nil, errors.Errorf("failed to open: %d", e)
	}
	if e := C.ftdi_write_data_set_chunksize(f.ctx, C.uint(1)); e < 0 {
		return nil, errors.Errorf("failed to set write chunk size: %d", e)
	}
	return f, nil
}

func (f *ftdiLibWrapper) SetBitBangMode(mask byte) error {
	if e := C.ftdi_set_bitmode(f.ctx, C.uchar(mask), C.uchar(modeAsyncBitBang)); e < 0 {
		return errors.Errorf("failed to set bit mode: %d", e)
	}
	return nil
}

func (f *ftdiLibWrapper) WriteByte(data byte) error {
	if n := C.ftdi_write_data(f.ctx, (*C.uchar)(&data), 1); n != 1 {
		return errors.Errorf("failed to write byte")
	}
	return nil
}

func (f *ftdiLibWrapper) Close() {
	if f.ctx != nil {
		if f.open {
			C.ftdi_usb_close(f.ctx)
		}
		C.ftdi_free(f.ctx)
		f.ctx = nil
	}
}
