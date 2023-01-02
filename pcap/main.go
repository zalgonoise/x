package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

var (
	snapshot_len int32         = 1024
	promiscuous  bool          = false
	timeout      time.Duration = 30 * time.Second
	w                          = os.Stdout // writer can be a file too
)

func main() {
	// set up channel on which to send signal notifications.
	//
	// must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// get all devices
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal(err)
	}
	var listening = []pcap.Interface{}

	// print device information
	fmt.Println("Devices found:")
	for _, device := range devices {
		fmt.Println("\nName: ", device.Name)
		fmt.Println("Description: ", device.Description)
		fmt.Println("Devices addresses: ", device.Description)
		for _, address := range device.Addresses {
			fmt.Println("- IP address: ", address.IP)
			fmt.Println("- Subnet mask: ", address.Netmask)
		}
		if len(device.Addresses) > 0 {
			listening = append(listening, device)
		}
	}

	// listen to each device with an address
	for _, device := range listening {
		device := device
		go func() {
			fmt.Println("\nListening on device: ", device.Name)
			handle, err := pcap.OpenLive(device.Name, snapshot_len, promiscuous, timeout)
			if err != nil {
				log.Fatal(err)
			}
			defer handle.Close()

			// use the handle as a packet source to process all packets
			packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
			for packet := range packetSource.Packets() {
				// process packet here
				_, _ = w.WriteString(packet.Dump())
			}
		}()
	}

	// block until a signal is received.
	s := <-c
	fmt.Println("Got signal:", s)
}
