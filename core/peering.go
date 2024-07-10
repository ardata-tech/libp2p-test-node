package core

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/multiformats/go-multiaddr"
	"libp2p-test-node/config"
	"log"
	"time"
)

type Peering struct {
}

func NewPeering() *Peering {
	return &Peering{}

}

type SignedMessage struct {
	Message    config.Message
	Signatures map[string]string // publicKey:signature
}

// NewLibP2PHost creates a new libp2p host
func (p *Peering) NewLibP2PHost(privKey crypto.PrivKey, listenPort string) (host.Host, multiaddr.Multiaddr, error) {

	listenAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%s", listenPort))
	if err != nil {
		log.Fatal(err)
	}
	opts := []libp2p.Option{
		libp2p.NATPortMap(),
		libp2p.DefaultTransports,
		libp2p.Identity(privKey),
		libp2p.ListenAddrs(listenAddr),
	}
	h, err := libp2p.New(
		opts...,
	)
	if err != nil {
		log.Fatal(err)
	}
	return h, listenAddr, nil
}

// SetupDiscovery sets up mDNS and DHT discovery
func (p *Peering) SetupDiscovery(h host.Host) (*dht.IpfsDHT, error) {
	// Set up DHT
	kadDHT, err := dht.New(context.Background(), h)
	if err != nil {
		return nil, err
	}
	if err := kadDHT.Bootstrap(context.Background()); err != nil {
		return nil, err
	}

	routingDiscovery := routing.NewRoutingDiscovery(kadDHT)
	util.Advertise(context.Background(), routingDiscovery, "eth-price")

	// Set up mDNS
	mdnsService := mdns.NewMdnsService(h, "eth-price", &mdnsNotifee{h: h})
	if err := mdnsService.Start(); err != nil {
		return nil, err
	}

	// Ensure the DHT is well-connected
	go func() {
		for {
			peers := kadDHT.RoutingTable().ListPeers()
			if len(peers) < 3 {
				log.Println("Insufficient peers, attempting to connect to more.")
				if err := kadDHT.Bootstrap(context.Background()); err != nil {
					log.Println("Error bootstrapping DHT:", err)
				}
			}
			time.Sleep(15 * time.Second)
		}
	}()

	return kadDHT, nil
}

type mdnsNotifee struct {
	h host.Host
}

// HandlePeerFound handles peer found events
func (n *mdnsNotifee) HandlePeerFound(pi peer.AddrInfo) {
	log.Printf("Found peer via mDNS: %s\n Addrs: %v\n", pi.ID, pi.Addrs)
	n.h.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.PermanentAddrTTL)
	n.h.Connect(context.Background(), pi)
}

// HandlePeerDisconnection handles peer disconnection events
func (p *Peering) HandlePeerDisconnection(h host.Host) {
	fmt.Println("Handling peer disconnection")
	sub, err := h.EventBus().Subscribe(new(event.EvtPeerConnectednessChanged))
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for e := range sub.Out() {
			connEvent := e.(event.EvtPeerConnectednessChanged)
			if connEvent.Connectedness == network.NotConnected {
				log.Printf("Peer disconnected: %s", connEvent.Peer)
				// Handle reconnection logic if needed
			} else {
				log.Printf("Peer connected: %s", connEvent.Peer)
			}
		}
	}()
}
