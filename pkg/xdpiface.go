package xdpiface

/*
#cgo LDFLAGS: -lxdpiface
#include <stdlib.h>
#include <xdpiface.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

const XDP_LOG_TRACE int = C.XDP_LOG_TRACE
const XDP_LOG_DEBUG int = C.XDP_LOG_DEBUG
const XDP_LOG_INFO int = C.XDP_LOG_INFO
const XDP_LOG_WARNING int = C.XDP_LOG_WARNING
const XDP_LOG_ERROR int = C.XDP_LOG_ERROR
const XDP_LOG_CRITICAL int = C.XDP_LOG_CRITICAL
const XDP_LOG_LVLS int = C.XDP_LOG_LVLS

func XdpLogLevelSet (level int)  {
	levelCint := C.int(level)
	C.xdp_log_level_set(levelCint)
}

const XDP_IFACE_DEFAULT string  = C.XDP_IFACE_DEFAULT
const XDP_IFACE_XDP_PROG_DEFAULT string  = C.XDP_IFACE_XDP_PROG_DEFAULT

type XdpIface struct {
	ctx *C.xdp_iface_t
}

func XdpIfaceNew (xdpInterface string) (*XdpIface, error) {
	self := new(XdpIface)
	xdpInterfaceCstr := C.CString(xdpInterface)
	defer C.free(unsafe.Pointer(xdpInterfaceCstr))
	self.ctx = C.xdp_iface_new(xdpInterfaceCstr)
	if self.ctx == nil {
		return nil, errors.New("Failed to create new XDP interface!")
	}

	return self, nil
}

func (self XdpIface) Destroy () {
	C.xdp_iface_destroy(&self.ctx)
	self.ctx = nil
}

func (self XdpIface) LoadProgram(xdpProgPath string) error {
	xdpProgPathCstr := C.CString(xdpProgPath)
	defer C.free(unsafe.Pointer(xdpProgPathCstr))

	ret := C.xdp_iface_load_program (self.ctx, xdpProgPathCstr)
	if ret != 0 {
		return errors.New("Failed to load program!")
	}

	return nil
}

func (self XdpIface) UnloadProgram() {
	C.xdp_iface_unload_program (self.ctx)
}

type XdpSock struct {
	ctx *C.xdp_sock_t
}

const XDP_SOCK_XSKS_MAP_DEFAULT string = C.XDP_SOCK_XSKS_MAP_DEFAULT
const (
	SO_PREFER_BUSY_POLL int = C.XDP_SOCK_SO_PREFER_BUSY_POLL
	SO_BUSY_POLL = C.XDP_SOCK_SO_BUSY_POLL
	SO_BUSY_POLL_BUDGET = C.XDP_SOCK_SO_BUSY_POLL_BUDGET
)

func XdpSockNew (xdpInterface *XdpIface) (*XdpSock, error) {
	self := new(XdpSock)
	self.ctx = C.xdp_sock_new(xdpInterface.ctx)
	if self.ctx == nil {
		return nil, errors.New("Failed to create new XDP socket!")
	}

	return self, nil
}

func (self XdpSock) Destroy () {
	C.xdp_sock_destroy(&self.ctx)
	self.ctx = nil
}

func (self XdpSock) LoopUpBpfMap (xdpInterface *XdpIface, mapName string,
	key_size uint32, value_size uint32) error {
	key_sizeC := C.uint32_t(key_size)
	value_sizeC := C.uint32_t(value_size)
	mapNameCstr := C.CString(mapName)
	defer C.free(unsafe.Pointer(mapNameCstr))

	ret := C.xdp_sock_lookup_bpf_map (self.ctx, xdpInterface.ctx, mapNameCstr, key_sizeC, value_sizeC)
	if ret != 0 {
		return errors.New("Failed to loop up BPF map")
	}

	return nil
}

func (self XdpSock) SetSockopt (optType int, optValue int) error {
	optTypeC := C.int(optType)
	optValueC := C.int(optValue)

	ret := C.xdp_sock_set_sockopt (self.ctx, optTypeC, optValueC)
	if ret != 0 {
		return errors.New("Failed to set sock opt")
	}

	return nil
}

func (self XdpSock) GetFd () (int, error) {
	ret := C.xdp_sock_get_fd (self.ctx)
	sockFd := int(ret)
	if sockFd <= 0 {
		return sockFd, errors.New("Failed to get sock dexcriptor!")
	}

	return sockFd, nil
}

func (self XdpSock) RxBatchGetSize (framesRecvd *uint32, batchSize uint32) error {
	var framesRecvdC C.uint32_t
	batchSizeC := C.uint32_t(batchSize)

	ret := C.xdp_sock_rx_batch_get_size (self.ctx, &framesRecvdC, batchSizeC)
	if ret != 0 {
		return errors.New("Failed to get batch size!")
	}
	*framesRecvd = uint32(framesRecvdC)

	return nil
}

func (self XdpSock) RxBatchRelease (framesRecvd uint32) error {
	framesRecvdC := C.uint32_t(framesRecvd)

	ret := C.xdp_sock_rx_batch_release (self.ctx, framesRecvdC)
	if ret != 0 {
		return errors.New("Failed to release batch!")
	}

	return nil
}

func (self XdpSock) Recv (buffer []byte, bufferSize *int) error {
	var bufferSizeC C.size_t
	bufferCPtr := (*C.char)(unsafe.Pointer(&buffer[0]))
	ret := C.xdp_sock_recv (self.ctx, bufferCPtr, &bufferSizeC)
	if ret != 0 {
		return errors.New("Recv Failed!")
	}
	*bufferSize = int(bufferSizeC)

	return nil
}

func (self XdpSock) TxBatchSetSize (batchSize uint32) error {
	batchSizeC := C.uint32_t(batchSize)

	ret := C.xdp_sock_tx_batch_set_size (self.ctx, batchSizeC)
	if ret != 0 {
		return errors.New("Failed to set batch size!")
	}

	return nil
}

func (self XdpSock) TxBatchRelease (framesSend uint32) error {
	framesSendC := C.uint32_t(framesSend)

	ret := C.xdp_sock_tx_batch_release (self.ctx, framesSendC)
	if ret != 0 {
		return errors.New("Failed to release batch!")
	}

	return nil
}

func (self XdpSock) Send (buffer []byte, bufferSize int) error {
	bufferSizeC := C.size_t(bufferSize)
	bufferCPtr := (*C.char)(unsafe.Pointer(&buffer[0]))
	ret := C.xdp_sock_send (self.ctx, bufferCPtr, bufferSizeC)
	if ret != 0 {
		return errors.New("Send Failed!")
	}

	return nil
}
