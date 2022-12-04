package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/multiformats/go-multiaddr"
	"github.com/renato0307/p2p-estimator/pkg/chatroom"
	"github.com/renato0307/p2p-estimator/pkg/discovery"
	"github.com/renato0307/p2p-estimator/pkg/ui"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/host/autorelay"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
)

// DiscoveryInterval is how often we re-publish our mDNS records.
const DiscoveryInterval = time.Hour

// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
const DiscoveryServiceTag = "pubsub-chat-example"

func main() {
	f, err := os.OpenFile("test.log", os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Printf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	// parse some flags to set our nickname and the room to join
	nickFlag := flag.String("nick", "", "nickname to use in estimation room. will be generated if empty")
	roomFlag := flag.String("room", "awesome-estimation-room", "name of chat room to join")
	ipAddressFlag := flag.String("addr", "0.0.0.0", "the ipv4 address to listen")
	ipPortFlag := flag.String("port", "0", "the ipv4 port to listen")
	bootstrapServerAddressFlag := flag.String("bootstrap-addr", "/ip4/176.78.115.195/tcp/4001/p2p/12D3KooWD6anii5FQwT2KPoutwQcBMcuXCME3YSkiJKZhtt1SzXQ", "address of the bootstrap server")
	flag.Parse()

	ctx := context.Background()
	serverMode := bootstrapServerAddressFlag == nil || *bootstrapServerAddressFlag == ""
	room := *roomFlag // join the room from the cli flag, or the flag default

	relayAddr, err := multiaddr.NewMultiaddr(*bootstrapServerAddressFlag)
	if err != nil {
		panic(err)
	}

	relayInfo, err := peer.AddrInfoFromP2pAddr(relayAddr)
	if err != nil {
		panic(err)
	}

	// create a new libp2p Host that listens on a random TCP port
	h, err := libp2p.New(
		libp2p.ForceReachabilityPrivate(),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/%s/tcp/%s", *ipAddressFlag, *ipPortFlag)),
		libp2p.EnableAutoRelay(autorelay.WithStaticRelays([]peer.AddrInfo{*relayInfo})),
	)
	if err != nil {
		panic(err)
	}

	discovery.DiscoverPeers(serverMode, bootstrapServerAddressFlag, ctx, h)

	// create a new PubSub service using the GossipSub router
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		panic(err)
	}

	// setup local mDNS discovery
	// if err := setupDiscovery(h); err != nil {
	// 	panic(err)
	// }

	if err := h.Connect(context.Background(), *relayInfo); err != nil {
		log.Printf("failed to connect to relay: %v", err)
		return
	}

	// Hosts that want to have messages relayed on their behalf need to reserve a slot
	// with the circuit relay service host
	_, err = client.Reserve(context.Background(), h, *relayInfo)
	if err != nil {
		log.Printf("failed to receive a relay reservation from relay. %v", err)
		return
	}

	// view host details and addresses
	log.Printf("host ID %s\n", h.ID().Pretty())
	log.Printf("following are the assigned addresses\n")
	for i, addr := range h.Addrs() {
		log.Printf("[%d] %s\n", i, addr.String())
	}

	// use the nickname from the cli flag, or a default if blank
	nick := *nickFlag
	if len(nick) == 0 {
		nick = defaultNick(h.ID())
	}

	if serverMode {
		log.Println("bootstrap server started!")
		select {}
	}

	// join the chat room
	cr, err := chatroom.JoinChatRoom(ctx, ps, h.ID(), nick, room)
	if err != nil {
		panic(err)
	}

	// draw the UI
	estimationUI := ui.NewEstimationUI(cr)
	if err = estimationUI.Run(); err != nil {
		log.Printf("error running text UI: %s", err)
	}
}

// defaultNick generates a nickname based on the $USER environment variable and
// the last 8 chars of a peer ID.
func defaultNick(p peer.ID) string {
	return os.Getenv("USER")
}

// // discoveryNotifee gets notified when we find a new peer via mDNS discovery
// type discoveryNotifee struct {
// 	h host.Host
// }

// // HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// // the PubSub system will automatically start interacting with them if they also
// // support PubSub.
// func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
// 	//TODO - show new peer notif - log.Printf("discovered new peer %s\n", pi.ID.Pretty())
// 	err := n.h.Connect(context.Background(), pi)
// 	if err != nil {
// 		return
// 		//TODO: show error - log.Printf("error connecting to peer %s: %s\n", pi.ID.Pretty(), err)
// 	}
// }

// // setupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
// // This lets us automatically discover peers on the same LAN and connect to them.
// func setupDiscovery(h host.Host) error {
// 	// setup mDNS discovery to find local peers
// 	s := mdns.NewMdnsService(h, DiscoveryServiceTag, &discoveryNotifee{h: h})
// 	return s.Start()
// }
