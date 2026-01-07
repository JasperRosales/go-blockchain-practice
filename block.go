package main

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"time"
)



// Simplified block structure
type Block struct {
	timestamp int64
	data      []byte
	prevHash  []byte
	hash      []byte
	nonce     int 		//required to verify a proof
}



// Sha256 hashing of block's contents
func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.timestamp, 10))
	headers := bytes.Join([][]byte{b.prevHash, b.data, timestamp}, []byte{})
	hash := sha256.Sum256(headers)
	b.hash = hash[:]
}

func NewBlock(data string, prevHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	block.hash = hash[:]
	block.nonce = nonce

	return block
}

func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}



