package discovery

import (
	"context"
	"log"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/multiformats/go-multiaddr"
)

func DiscoverPeers(serverMode bool, bootstrapServerAddressFlag *string, ctx context.Context, h host.Host) {
	// if bootstrapServerAddressFlag is nil
	// setup DHT with empty discovery peers
	// so this will be a discovery peer for others
	// this peer should run on cloud (with public ip address)
	// setup peer discovery

	discoveryPeers := []multiaddr.Multiaddr{}
	if !serverMode {
		log.Println("running in normal host mode")

		multiAddr, err := multiaddr.NewMultiaddr(*bootstrapServerAddressFlag)
		if err != nil {
			panic(err)
		}
		discoveryPeers = append(discoveryPeers, multiAddr)
	} else {
		log.Println("running in server mode")
	}
	dht, err := NewDHT(ctx, h, discoveryPeers)
	if err != nil {
		panic(err)
	}

	var routingDiscovery = routing.NewRoutingDiscovery(dht)

	go discoverLoop(ctx, h, dht, routingDiscovery, "discovery")
}

func discoverLoop(ctx context.Context, h host.Host, dht *dht.IpfsDHT, routingDiscovery *routing.RoutingDiscovery, rendezvous string) {
	dutil.Advertise(ctx, routingDiscovery, rendezvous)
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			peers, err := dutil.FindPeers(ctx, routingDiscovery, rendezvous)
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("found %d peers", len(peers))
			for _, p := range peers {
				if p.ID == h.ID() {
					continue
				}
				err = h.Connect(ctx, p)
				if err != nil {
					log.Printf("error connecting to peer %s: %s\n", p.ID.Pretty(), err)
				} else {
					log.Printf("connected to peer %s\n", p.ID.Pretty())
				}
				// log.Printf("peer ID %s\n", p.ID.Pretty())
				// log.Printf("peer addresses\n")
				// for i, addr := range p.Addrs {
				// 	log.Printf("[%d] %s\n", i, addr.String())
				// }
			}
		}
	}
}
