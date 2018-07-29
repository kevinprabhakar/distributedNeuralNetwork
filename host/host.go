package host

import (
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"log"

	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"
	pstore "gx/ipfs/QmPgDWmTmuzvP7QE5zwo1TmjbJme9pmZHNujB2453jkCTr/go-libp2p-peerstore"
	swarm "gx/ipfs/QmU219N3jn7QadVCeBUqGnAkwoXoUomrCwDuVQVuL7PB5W/go-libp2p-swarm"
	ma "gx/ipfs/QmXY77cVe7rVRQXZZQRioukUM7aRW3BTcAgJe12MCtb3Ji/go-multiaddr"
	peer "gx/ipfs/QmXYjuNuxVzXKJCfWasQk1RqkhVLDM9jtUKhqc2WPQmFSB/go-libp2p-peer"
	crypto "gx/ipfs/QmaPbCnUMBohSGo3KnxEa2bHqyJVVeEEcwtqJAYxerieBo/go-libp2p-crypto"
	"strconv"
	//"net/http"
)

type ClientPayload struct{
	Model 		[]byte 		`json:"model"`
	Inputs 		[]string		`json:"inputs"`
	Outputs 	[]string 		`json:"outputs"`
	BatchSize	int			`json:"batchSize"`
	Epochs 		int 		`json:"epochs"`
}

type ListenerPayload struct{
	Weights 	string 		`json:"weights"`
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func NewHost(port int) (*bhost.BasicHost, string, string){
	// Generate an identity keypair using go's cryptographic randomness source
	priv, pub, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		panic(err)
	}

	// A peers ID is the hash of its public key
	pid, err := peer.IDFromPublicKey(pub)
	if err != nil {
		panic(err)
	}

	// We've created the identity, now we need to store it.
	// A peerstore holds information about peers, including your own
	ps := pstore.NewPeerstore()
	ps.AddPrivKey(pid, priv)
	ps.AddPubKey(pid, pub)

	localIP := GetOutboundIP().String()
	url := fmt.Sprintf("/ip4/%s/tcp/%s", localIP, strconv.Itoa(port))

	maddr, err := ma.NewMultiaddr(url)
	if err != nil {
		panic(err)
	}

	// Make a context to govern the lifespan of the swarm
	ctx := context.Background()

	// Put all this together
	netw, err := swarm.NewNetwork(ctx, []ma.Multiaddr{maddr}, pid, ps, nil)
	if err != nil {
		panic(err)
	}

	myhost := bhost.New(netw,nil)
	if err != nil {
		panic(err)
	}

	//fmt.Printf("Hello World, my hosts ID is %s\n", myhost.ID())

	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", myhost.ID().Pretty()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	fullAddr := maddr.Encapsulate(hostAddr)

	//fmt.Printf("Now run \"./<executable name> -l %d -d %s\" on a different terminal\n", port+1, fullAddr)

	return myhost, fullAddr.String(), myhost.ID().String()
}

func GetPIDandTargetAddrFromTargetURL(target string)(*peer.ID, *ma.Multiaddr) {
	ipfsaddr, err := ma.NewMultiaddr(target)
	if err != nil {
		log.Fatalln(err)
		return nil,nil
	}

	pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		log.Fatalln(err)
		return nil,nil
	}

	peerid, err := peer.IDB58Decode(pid)
	if err != nil {
		log.Fatalln(err)
		return nil,nil
	}

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	targetPeerAddr, _ := ma.NewMultiaddr(
		fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
	targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

	return &peerid, &targetAddr
}
