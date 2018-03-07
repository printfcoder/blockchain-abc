package main

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"time"
)

// Block 简单的区块
// 先从“区块链”中的“区块”说起。在区块链中，块存储了变量信息，比如，比特币的区块存储了交易、还有加密货币
// 除了这些，区块包含了一些技术信息，比如版本、时间戳、还有排在前面的一个区块的hash值
// Timestamp 时间戳也即是在区块被创建时的时间
// Data 就是这个区块存储的变量信息，
// PrevBlockHash 前一区块的hash值
// Hash 是当前区块的hash值
// 和比特币分开存储的数据结构不同的是 Timestamp、PrevBlockHash、Hash是区块的头（headers）信息，交易（transactions，我们
// 这里转成Data来称呼）是在数据（data）信息中。这里把这些概念放在一块，方便些。
type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
}

// SetHash 设置本区块的hash值，让区块串联起来
// 那为什么要计算hash呢？
// 计算hash值在区块链中是非常重要的特性，这一特性使得区块链是安全的。因为计算指定有特征的hash非常困难，即使在牛逼的计算机中也要花上一些时间
// 计算出来（所以有的人就买更适合简单浮点运算的GPU去挖Bitcoin矿）。这么做是故意的，因为这样可以增加创建新块的难度，导致增加了区块的节点无法
// 在增加后改动这个区块，而改动后，这个区块也就失效了，不被大家承认。
func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
	hash := sha256.Sum256(headers)
	b.Hash = hash[:]
}

// NewBlock 实现一个简单的创建区块方法
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}}
	block.SetHash()
	return block
}

// NewGenesisBlock 创世区块
// 为了创建新的区块，需要一个已经存在的区块，但是现在还没有任何一个区块。而在区块链中，第一个区块，就是“创世区块”。
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}
