package core

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/libp2p/go-libp2p/core/crypto"
	"log"
)

type Pki struct {
}

func NewPki() *Pki {
	return &Pki{}
}

func (p *Pki) GenerateKeyPair() (crypto.PrivKey, crypto.PubKey, error) {
	priv, _, err := crypto.GenerateECDSAKeyPair(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	pub := priv.GetPublic()
	return priv, pub, nil
}

func (p *Pki) VerifySignature(msg []byte, sig string, pubKey crypto.PubKey) bool {
	signature, err := hex.DecodeString(sig)
	if err != nil {
		log.Println("Failed to decode signature:", err)
		return false
	}

	valid, err := pubKey.Verify(msg, signature)
	if err != nil || !valid {
		log.Println("Failed to verify signature:", err)
		return false
	}

	return true
}
