package xdpiface

import (
	"testing"
	"bytes"
	"fmt"
)

func TestXdpIface(t *testing.T) {
	xdp_iface, err := XdpIfaceNew(XDP_IFACE_DEFAULT)
	if err != nil {
		t.Errorf("Failed to create XDP iface")
	}
	defer xdp_iface.Destroy()
}

func TestXdpSock(t *testing.T) {
	batch_size := uint32(30)
	var frames_rcvd uint32 = 0

	var oBufferSize int = 1000
	oBuffer := bytes.Repeat([]byte{0x55}, oBufferSize)

	var iBufferSize int = 0
	iBuffer := make([]byte, 9000)

	XdpLogLevelSet(XDP_LOG_INFO)

	xdp_iface, err := XdpIfaceNew(XDP_IFACE_DEFAULT)
	if err != nil {
		t.Errorf("Failed to create XDP iface")
	}
	defer xdp_iface.Destroy()

	xdp_iface.LoadProgram(XDP_IFACE_XDP_PROG_DEFAULT)

	xdp_sock, err := XdpSockNew(xdp_iface)
	if err != nil {
		t.Errorf("Failed to create XDP sock")
	}
	defer xdp_sock.Destroy()
	xdp_sock.LoopUpBpfMap(xdp_iface, XDP_SOCK_XSKS_MAP_DEFAULT, 4, 4)

    xdp_sock.SetSockopt(SO_PREFER_BUSY_POLL, 1)
    xdp_sock.SetSockopt(SO_BUSY_POLL, 20)
    xdp_sock.SetSockopt(SO_BUSY_POLL_BUDGET, int(batch_size))

    xdp_sock.TxBatchSetSize(batch_size);
  	for i := 1; i <= int(batch_size); i++ {
        xdp_sock.Send (oBuffer, oBufferSize)
	}
    xdp_sock.TxBatchRelease(batch_size);
    fmt.Printf("--- Frames sent: %d\n", batch_size)

    xdp_sock.RxBatchGetSize (&frames_rcvd, batch_size);
	for i := 1; i <= int(frames_rcvd); i ++ {
        xdp_sock.Recv (iBuffer, &iBufferSize)
    }
    xdp_sock.RxBatchRelease(frames_rcvd)
    fmt.Printf("--- Frames received: %d\n", frames_rcvd)

	xdp_iface.UnloadProgram()
}
