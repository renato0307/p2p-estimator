package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/renato0307/p2p-estimator/pkg/chatroom"
	"github.com/renato0307/p2p-estimator/pkg/ui"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

// DiscoveryInterval is how often we re-publish our mDNS records.
const DiscoveryInterval = time.Hour

// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
const DiscoveryServiceTag = "pubsub-chat-example"

func main() {
	// parse some flags to set our nickname and the room to join
	nickFlag := flag.String("nick", "", "nickname to use in estimation room. will be generated if empty")
	roomFlag := flag.String("room", "awesome-estimation-room", "name of chat room to join")
	ipAddressFlag := flag.String("addr", "0.0.0.0", "the ipv4 address to listen")
	ipPortFlag := flag.String("port", "0", "the ipv4 port to listen")
	// TODO: uncomment when handling public nodes discovery
	// bootstrapServerAddressFlag := flag.String("bootstrap-addr", "", "address of the bootstrap server")
	flag.Parse()

	ctx := context.Background()
	// TODO: uncomment when handling public nodes discovery
	// serverMode := bootstrapServerAddressFlag == nil || *bootstrapServerAddressFlag == ""
	room := *roomFlag // join the room from the cli flag, or the flag default

	// create a new libp2p Host that listens on a random TCP port
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/%s/tcp/%s", *ipAddressFlag, *ipPortFlag)))
	if err != nil {
		panic(err)
	}

	// view host details and addresses
	log.Printf("host ID %s\n", h.ID().Pretty())
	log.Printf("following are the assigned addresses\n")
	for _, addr := range h.Addrs() {
		log.Printf("%s\n", addr.String())
	}
	log.Printf("\n")

	// create a new PubSub service using the GossipSub router
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		panic(err)
	}

	// TODO: uncomment when handling public nodes discovery
	// discoveryPeers := []multiaddr.Multiaddr{}
	// if !serverMode {
	// 	log.Println("running in normal host mode")
	// 	// if bootstrapServerAddressFlag is nil
	// 	// setup DHT with empty discovery peers
	// 	// so this will be a discovery peer for others
	// 	// this peer should run on cloud(with public ip address)
	// 	multiAddr, err := multiaddr.NewMultiaddr(*bootstrapServerAddressFlag)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	discoveryPeers = append(discoveryPeers, multiAddr)
	// } else {
	// 	log.Println("running in server mode")
	// }
	// dht, err := discovery.NewDHT(ctx, h, discoveryPeers)
	// if err != nil {
	// 	panic(err)
	// }

	// // setup peer discovery
	// go discovery.Discover(ctx, h, dht, room)

	// setup local mDNS discovery
	if err := setupDiscovery(h); err != nil {
		panic(err)
	}

	// use the nickname from the cli flag, or a default if blank
	nick := *nickFlag
	if len(nick) == 0 {
		nick = defaultNick(h.ID())
	}

	// TODO: uncomment when handling public nodes discovery
	// if serverMode {
	// 	log.Println("bootstrap server started!")
	// 	select {}
	// }

	// join the chat room
	cr, err := chatroom.JoinChatRoom(ctx, ps, h.ID(), nick, room)
	if err != nil {
		panic(err)
	}

	// draw the UI
	estimationUI := ui.NewEstimationUI(cr)
	if err = estimationUI.Run(); err != nil {
		printErr("error running text UI: %s", err)
	}
}

// printErr is like log.Printf, but writes to stderr.
func printErr(m string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, m, args...)
}

// defaultNick generates a nickname based on the $USER environment variable and
// the last 8 chars of a peer ID.
func defaultNick(p peer.ID) string {
	return os.Getenv("USER")
}

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h host.Host
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	//TODO - show new peer notif - log.Printf("discovered new peer %s\n", pi.ID.Pretty())
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		return
		//TODO: show error - log.Printf("error connecting to peer %s: %s\n", pi.ID.Pretty(), err)
	}
}

// setupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
// This lets us automatically discover peers on the same LAN and connect to them.
func setupDiscovery(h host.Host) error {
	// setup mDNS discovery to find local peers
	s := mdns.NewMdnsService(h, DiscoveryServiceTag, &discoveryNotifee{h: h})
	return s.Start()
}
