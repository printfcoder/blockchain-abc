package main

/**
实现一个简单的区块链，区块链的本质就是一个一定结构的数据库：
   有序的、首尾相连的链状列表。即是，区块们都是顺序、每一个块都连接着前面的一个块。
   这个结构使得可以在区块链中快速找到最后一个区块，尤其是可以通过hash值找到区块
在golang里可以使用数组、map来实现，数组可以保证顺序，map实现hash->block组合的映射
不过，针对目前的进度，我们不需要实现能过hash找到区块的方法，所以这里只用数组来保证顺序即可。
**/

// Blockchain 区块链结构
// 这就是一个简单的区块链了
type Blockchain struct {
	blocks []*Block
}

// AddBlock 给区块链添加增加区块的能力
func (bc *Blockchain) AddBlock(data string) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	newBlock := NewBlock(data, prevBlock.Hash)
	bc.blocks = append(bc.blocks, newBlock)
}

// NewBlockchain 构建新区块链
// 使用创世区块来引导区块链
func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{NewGenesisBlock()}}
}
