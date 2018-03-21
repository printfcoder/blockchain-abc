# Golang 区块链入门 第一章 基本概念

区块链是21世纪重要的技术革命之一，虽然还没有成熟，但是仍然有很多潜力尚待发掘。基于区块链的本质，它就是一个分布式记录数据库。但是和私有数据库不同的是，区块链是公开的（私有链在域内也是公开的），也就是说每一个使用它的人都会有完整或者说部分副本。而且新的记录要增加的话，需要得到链中其它拥有者的同意。而且，区块链使得加密货币和智能合约成为可能。这一系列文章，将会阐述和实现基于简单的区块链来生成简单的加密货币。
[原文][原文]（略有删改）

## Block 块/区块
先从“区块链”中的“区块”说起。在区块链中，块存储了变量信息，比如，比特币的区块存储了交易、还有加密货币除了这些，区块包含了一些技术信息，比如版本、时间戳、还有排在前面的一个区块的hash值

```
Timestamp 时间戳也即是在区块被创建时的时间
``` 
``` 
Data 就是这个区块存储的变量信息
``` 
``` 
PrevBlockHash 前一区块的hash值
``` 
``` 
Hash 是当前区块的hash值
``` 

和比特币分开存储的数据结构不同的是 Timestamp、PrevBlockHash、Hash是区块的头（headers）信息，交易（transactions，我们
这里转成Data来称呼）是在数据（data）信息中。这里把这些概念放在一块，方便些:

### 块代码结构
```golang
type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
}
```
### hash
那为什么要计算hash呢？计算hash值在区块链中是非常重要的特点，这使得区块链是安全的。因为计算有指定特征的hash非常困难，即使在牛逼的计算机中也要花上一些时间计算出来（所以有的人就买更适合简单浮点运算的GPU去挖Bitcoin矿）。这么做是故意的，因为这样可以增加创建新块的难度，导致增加了区块的节点无法在增加后改动这个区块，而改动后，这个区块也就失效了，不被大家承认。
[区块链的hash算法][hash算法]。为了简单，我们现在基于SHA-256构造`SetHash`方法，

```golang
func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
	hash := sha256.Sum256(headers)
	b.Hash = hash[:]
}
```

### 创建区块
实现一个简单的创建区块方法：

```golang
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}}
	block.SetHash()
	return block
}
```


---

## blackchain 区块链

开篇说过，区块链的本质就是一个一定结构的数据库。它是一个有序的、首尾相连的链状列表，区块们都是顺序、每一个块都连接着前面的一个块。这个结构使得可以在区块链中快速找到最后一个区块，尤其是可以通过hash值找到区块。


### 定义简单的区块链

在golang里可以使用数组、map来实现，数组可以保证顺序，map实现hash->block组合的映射
不过，针对目前的进度，我们不需要实现能过hash找到区块的方法，所以这里只用数组来保证顺序即可。

```golang
type Blockchain struct {
	blocks []*Block
}
```

然后给区块链添加增加区块的能力：

```golang
func (bc *Blockchain) AddBlock(data string) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	newBlock := NewBlock(data, prevBlock.Hash)
	bc.blocks = append(bc.blocks, newBlock)
}
```

### 创世区块

为了创建新的区块，需要一个已经存在的区块，但是现在还没有任何一个区块。而在区块链中，第一个区块，就是“创世区块”。

```golang
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}
```

使用创世区块来引导区块链

```golang
func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{NewGenesisBlock()}}
}
```



## 运行

现在可以在命令行时输入 `go run *.go`运行创建区块链

```golang
func main() {
	bc := NewBlockchain()

	bc.AddBlock("Send 1 BTC to Ivan")
	bc.AddBlock("Send 2 more BTC to Ivan")

	for _, block := range bc.blocks {
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Println()
	}
}
```

控制台会输出区块链的内部信息：

```log

Prev. hash:
Data: Genesis Block
Hash: 4f729464de88e6a01c59a54707b11d3efd5fb036637fa81f4fbcce437b7b0738

Prev. hash: 4f729464de88e6a01c59a54707b11d3efd5fb036637fa81f4fbcce437b7b0738
Data: Send 1 BTC to Ivan
Hash: 53655b661290d9d4c9973618dfd2b5cb71c8c2981b5c955fa70af5d6a30b02be

Prev. hash: 53655b661290d9d4c9973618dfd2b5cb71c8c2981b5c955fa70af5d6a30b02be
Data: Send 2 more BTC to Ivan
Hash: 22cbee453308893beca8fd023c77d61a05eca29db99272a2795d0ae7af7d306d

```

## 本章总结

我们创建了简单的区块链原型：只有一个数组来维护的链，每个块都拥有前一个块的hash值来保证彼此的连接。真正的区块链自然是要比这里的复杂得多的。这里的区块链产生很简单也很快，但是真正的区块链产生需要做很多工作，如果要获得一个区块，那么需要做大量而繁重的计算，这一机制被称为工作量证明（Proof-of-Work）。区块链是分布式的且没有决定者（去中心化）。这就是说，新的区块增加需要得到其它网络中参与运算的节点认可（共识）。在我们上面的例子中，还没有一笔交易，所以，不算正式意义上的区块链！

以后的文章会继续讨论每一个特性。

## 相关链接

### 延伸与代码

1. [本文代码][本文代码]
2. [hash算法][hash算法]

### 本序列文章：

1. [Golang 区块链入门 第一章 基本概念][本序列第一篇]
2. [Golang 区块链入门 第二章 工作量证明][本序列第二篇]
3. [Golang 区块链入门 第三章 持久化和命令行接口][本序列第三篇]
4. [Golang 区块链入门 第四章 交易 第一节][本序列第四篇]
5. [Golang 区块链入门 第五章 地址][本序列第五篇]
6. [Golang 区块链入门 第六章 交易 第二节][本序列第六篇]
7. [Golang 区块链入门 第七章 网络][本序列第七篇]

[本序列第一篇]: https://printfcoder.github.io/myblog/myblog/blockchain/abc/2018/03/05/abc-building-blockchain-in-go-part-1-basic-prototype/
[本序列第二篇]: https://printfcoder.github.io/myblog/myblog/blockchain/abc/2018/03/06/abc-building-blockchain-in-go-part-2-proof-of-work/
[本序列第三篇]: https://printfcoder.github.io/myblog/myblog/blockchain/abc/2018/03/07/abc-building-blockchain-in-go-part-3-persistence-and-cli/
[本序列第四篇]: https://printfcoder.github.io/myblog/myblog/blockchain/abc/2018/03/09/abc-building-blockchain-in-go-part-4-transactions-1/
[本序列第五篇]: https://printfcoder.github.io/myblog/myblog/blockchain/abc/2018/03/14/abc-building-blockchain-in-go-part-5-address/
[本序列第六篇]: https://printfcoder.github.io/myblog/myblog/blockchain/abc/2018/03/17/abc-building-blockchain-in-go-part-6-transactions-2/
[本序列第七篇]: https://printfcoder.github.io/myblog/myblog/blockchain/abc/2018/03/20/abc-building-blockchain-in-go-part-7-network/

[原文]: https://jeiwan.cc/posts/building-blockchain-in-go-part-1/
[hash算法]: https://en.bitcoin.it/wiki/Block_hashing_algorithm
[本文代码]:  https://github.com/printfcoder/blockchain-abc/tree/part_1