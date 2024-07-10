package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"libp2p-test-node/api"
	"libp2p-test-node/config"
	"libp2p-test-node/core"
	"libp2p-test-node/protocols"
	"log"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/multiformats/go-multiaddr"
)

func main() {

	port := flag.String("listen-port", "4001", "Port to listen on")
	peerAddresses := flag.String("bootstrap-peers", "", "Comma-separated list of peer multiaddresses to connect to")
	flag.Parse()

	// initialize config
	cfg := config.InitConfig()
	go func() {
		api.InitializeEchoServer(cfg)
	}()

	if *port != "" {
		cfg.Peer.ListenPort = *port
	}

	// initilaize new peering
	pkiSvc := core.NewPki()
	privKey, pubKey, err := pkiSvc.GenerateKeyPair()
	if err != nil {
		log.Fatal(err)
	}

	// initialize the new p2p host
	peeringSvc := core.NewPeering()
	h, listenAddr, err := peeringSvc.NewLibP2PHost(privKey, cfg.Peer.ListenPort)
	nodeID := h.ID().String()

	// handle peer connection
	peeringSvc.HandlePeerDisconnection(h)

	log.Printf("Node ID: %s\n", nodeID)
	log.Printf("Listening on: %s\n", listenAddr)
	if *peerAddresses != "" {
		peers := strings.Split(*peerAddresses, ",")
		for _, peerAddr := range peers {
			addr, err := multiaddr.NewMultiaddr(peerAddr)
			if err != nil {
				log.Println("Error parsing peer address:", err)
				continue
			}

			peerinfo, err := peer.AddrInfoFromP2pAddr(addr)
			if err != nil {
				log.Println("Error getting peer info:", err)
				continue
			}

			h.Peerstore().AddAddrs(peerinfo.ID, peerinfo.Addrs, peerstore.PermanentAddrTTL)
			if err := h.Connect(context.Background(), *peerinfo); err != nil {
				log.Println("Error connecting to peer:", err)
			} else {
				log.Printf("Connected to peer: %s\n", peerinfo.ID)
			}
		}
	}

	// initialize the discovert using ipfs dht
	var dhtIpfs *dht.IpfsDHT
	if dhtIpfs, err = peeringSvc.SetupDiscovery(h); err != nil {
		log.Fatal(err)
	}

	// initialize the pubsub
	pubSub := core.NewPubSub(h, cfg.PubSub.TopicName)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// run the price checker
	go func() {
		for {
			priceService := protocols.NewPriceService()

			ethPrice, err := priceService.GetPrice()
			if err != nil {
				log.Println(err)
				continue
			}
			// if eth price is 0, don't publish
			if ethPrice.Ethereum.Usd == 0 {
				continue
			}

			msg := &config.Message{
				Price:     ethPrice.Ethereum.Usd,
				Denom:     "eth",
				Timestamp: time.Now(),
				NodeID:    nodeID, // Set the origin node ID
			}
			msgJson, err := json.Marshal(msg)
			if err != nil {
				log.Println(err)
				continue
			}
			sig, err := privKey.Sign(msgJson)
			if err != nil {
				log.Println(err)
				continue
			}
			pubKeyBytes, err := crypto.MarshalPublicKey(pubKey)
			if err != nil {
				log.Println(err)
				continue
			}
			signedMsg := &core.SignedMessage{
				Message:    *msg,
				Signatures: map[string]string{hex.EncodeToString(pubKeyBytes): hex.EncodeToString(sig)},
			}
			signedMsgJson, err := json.Marshal(signedMsg)
			if err != nil {
				log.Println(err)
				continue
			}

			if err := pubSub.Topic.Publish(context.Background(), signedMsgJson); err != nil {
				log.Println(err)
			}
			//log.Printf("Published: %s\n", ethPrice)

			<-ticker.C
		}
	}()

	// store the message
	messageStore := make(map[string]*core.SignedMessage)
	cfg.DHT = dhtIpfs
	cfg.Host = h
	for {
		msg, err := pubSub.Sub.Next(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		//log.Printf("Received: %s\n", msg.Data)

		var receivedMsg core.SignedMessage
		if err := json.Unmarshal(msg.Data, &receivedMsg); err != nil {
			log.Println("Failed to unmarshal message:", err)
			continue
		}

		// Skip if the message was published by the same node
		if receivedMsg.Message.NodeID == nodeID {
			continue
		}

		// unmarshal the message
		msgJson, err := json.Marshal(receivedMsg.Message)
		if err != nil {
			log.Println("Failed to marshal message:", err)
			continue
		}

		// hash the message
		hash := sha256.Sum256(msgJson)
		msgID := hex.EncodeToString(hash[:]) // Use the hash as the message ID
		if storedMsg, exists := messageStore[msgID]; exists {
			for pubKeyHex, sig := range receivedMsg.Signatures {
				if _, exists := storedMsg.Signatures[pubKeyHex]; !exists {
					storedMsg.Signatures[pubKeyHex] = sig
				}
			}
		} else {
			messageStore[msgID] = &receivedMsg
		}

		// Verify the signatures
		validSignatures := 0
		for pubKeyHex, sig := range messageStore[msgID].Signatures {
			pubKeyBytes, err := hex.DecodeString(pubKeyHex)
			if err != nil {
				log.Println("Failed to decode public key:", err)
				continue
			}
			pubKeyS, err := crypto.UnmarshalPublicKey(pubKeyBytes)
			if err != nil {
				log.Println("Failed to unmarshal public key:", err)
				continue
			}
			if pkiSvc.VerifySignature(msgJson, sig, pubKeyS) {
				validSignatures++
			}
		}

		// If the message has 3 valid signatures, store it
		if validSignatures >= 3 {
			if err := cfg.DB.Create(&messageStore[msgID].Message).Error; err != nil {
				//log.Println("Failed to store message:", err)
			}
			delete(messageStore, msgID)
			validSignatures = 0
		} else {
			// Node signs the message and republishes it
			sig, err := privKey.Sign(msgJson)
			if err != nil {
				log.Println("Failed to sign message:", err)
				continue
			}
			pubKeyBytes, err := crypto.MarshalPublicKey(pubKey)
			if err != nil {
				log.Println("Failed to marshal public key:", err)
				continue
			}
			messageStore[msgID].Signatures[hex.EncodeToString(pubKeyBytes)] = hex.EncodeToString(sig)
			signedMsgJson, err := json.Marshal(messageStore[msgID])
			if err != nil {
				log.Println("Failed to marshal signed message:", err)
				continue
			}
			if err := pubSub.Topic.Publish(context.Background(), signedMsgJson); err != nil {
				log.Println("Failed to republish message:", err)
			}
		}
	}
}
