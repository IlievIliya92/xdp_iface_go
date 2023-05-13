package main

import (
	"os"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/IlievIliya92/xdp_iface_go/pkg"
)

func main () {
	batch_size := uint32(30)
	var frames_rcvd uint32 = 0
	var iBufferSize int = 0
	iBuffer := make([]byte, 9000)

	xdpiface.XdpLogLevelSet(xdpiface.XDP_LOG_INFO)

	xdp_iface, err := xdpiface.XdpIfaceNew(xdpiface.XDP_IFACE_DEFAULT)
	if err != nil {
		fmt.Println("Failed to create XDP iface", err)
	}
	defer xdp_iface.Destroy()

	xdp_iface.LoadProgram(xdpiface.XDP_IFACE_XDP_PROG_DEFAULT)
	defer xdp_iface.UnloadProgram()

	xdp_sock, err := xdpiface.XdpSockNew(xdp_iface)
	if err != nil {
		fmt.Println("Failed to create XDP sock", err)
	}
	defer xdp_sock.Destroy()
	xdp_sock.LoopUpBpfMap(xdp_iface, xdpiface.XDP_SOCK_XSKS_MAP_DEFAULT, 4, 4)

    xdp_sock.SetSockopt(xdpiface.SO_PREFER_BUSY_POLL, 1)
    xdp_sock.SetSockopt(xdpiface.SO_BUSY_POLL, 20)
    xdp_sock.SetSockopt(xdpiface.SO_BUSY_POLL_BUDGET, int(batch_size))

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	var stop bool = false
	go func() {
	    <-interrupt
	    fmt.Printf("Clossing app...")
	    stop = true
	}()

	for {
	    xdp_sock.RxBatchGetSize (&frames_rcvd, batch_size);
	    fmt.Println("Frames received: ", frames_rcvd)
		for i := 1; i <= int(frames_rcvd); i ++ {
	        xdp_sock.Recv (iBuffer, &iBufferSize)
			// Decode a packet
			packet := gopacket.NewPacket(iBuffer, layers.LayerTypeEthernet, gopacket.Default)
			// Iterate over all layers, printing out each layer type
			for _, layer := range packet.Layers() {
			  fmt.Println("LAYER:", layer.LayerType())
			}
	    }

	     if stop {
		 	break
		 }
	}
    xdp_sock.RxBatchRelease(frames_rcvd)
}
