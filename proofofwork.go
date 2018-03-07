package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
	"time"
)

var (
	maxNonce = math.MaxInt64
)

const targetBits = 24

// ProofOfWork 工作量证明结构
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// NewProofOfWork 新建工作证明
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{b, target}

	return pow
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.Data,
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

// Run 执行PoW实际上的计算
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0
	start := time.Now()
	fmt.Println("start:", start)
	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)
	fmt.Print("\n")
	for nonce < maxNonce {
		data := pow.prepareData(nonce)

		hash = sha256.Sum256(data)
		// fmt 可能会阻塞，造成总体计算过慢，可选择不打印 fmt.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}

	fmt.Printf("\r%x\n", hash)
	elapsed := time.Since(start)
	fmt.Println("end: ", time.Now())
	fmt.Println("elapsed time: ", elapsed)

	fmt.Print("\n\n")
	return nonce, hash[:]
}

// Validate 难PoW是否合法
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}
