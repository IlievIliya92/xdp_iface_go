package main

import (
	"os"
	"fmt"
	"flag"
	"os/signal"
	"syscall"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/IlievIliya92/xdp_iface_go/pkg"
)

func main () {
	// Define flags
	net_iface := flag.String("net_iface", xdpiface.XDP_IFACE_DEFAULT, "First argument")
	batch_size := flag.Int("batch_size", 30, "Batch Size")

	// Parse command-line arguments
	flag.Parse()

	// Access the flag values
	fmt.Println("Net Interface:", *net_iface)
	fmt.Println("Batch size:", *batch_size)

	var frames_rcvd uint32 = 0
	var iBufferSize int = 0
	iBuffer := make([]byte, 9000)

	xdpiface.XdpLogLevelSet(xdpiface.XDP_LOG_INFO)
	xdp_iface, err := xdpiface.XdpIfaceNew(*net_iface)
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
    xdp_sock.SetSockopt(xdpiface.SO_BUSY_POLL_BUDGET, *batch_size)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	var stop bool = false
	go func() {
	    <-interrupt
	    fmt.Printf("Clossing app...")
	    stop = true
	}()

	var udp_dgrams uint64 = 0
	for {
	    xdp_sock.RxBatchGetSize (&frames_rcvd, uint32(*batch_size));
		for i := 1; i <= int(frames_rcvd); i ++ {
	        xdp_sock.Recv (iBuffer, &iBufferSize)
			// Decode a packet
			packet := gopacket.NewPacket(iBuffer, layers.LayerTypeEthernet, gopacket.NoCopy)
			if err := packet.ErrorLayer(); err != nil {
				fmt.Println("Error decoding part of the packet:", err)
			}

			if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
				udpPacket, _ := udpLayer.(*layers.UDP)
				udp_dgrams++
				// Access UDP packet information
				fmt.Printf("Source Port: %d\n", udpPacket.SrcPort)
				fmt.Printf("Destination Port: %d\n", udpPacket.DstPort)
				fmt.Printf("Payload: %x (%s)\n", udpPacket.Payload, udpPacket.Payload)
				fmt.Println("--------------------------------------")
			}
	    }
    	xdp_sock.RxBatchRelease(frames_rcvd)

	    if stop {
		 	break
		}
	}
	fmt.Printf("\nUDP dgrams received: %d\n", udp_dgrams)
}
