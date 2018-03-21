# Golang 区块链入门 第三节 持久化和命令行接口

在本系列的前面几篇文章中，我们编写了一个拥有挖矿能力的PoW系统，我们的实现逐步接近全功能的区块链，但是我们没有实现一些非常重要的特性。今天我们要实现区块链持久化，然后再做一个简单的CLI[command-line interface]来操作区块链。本质上，区块链就是分布式的数据库。我们暂先忽略“分布式”，把精力集中在“数据库”上。

[原文][原文]（略有删改）

# 数据库选型

直到现在，我们的区块链实现中还没有用到数据库，我们只是把每次启动程序计算得到的区块储存在内存中。我们不能复用一个之前生成的区块链，也不能与他人分享，因此，现在我们要把它存在磁盘上。

那该选择什么样的数据库？其实任何一种都可以。在[比特币文档][bitcoin_pdf]中，没有说要一个具体的数据库，所以这取决于开发者。[Bitcoin Core](https://github.com/bitcoin/bitcoin)用的是[LevelDB](https://github.com/google/leveldb)。本篇教程中使用`BoltDB`。

# BoltDB

BoltDB有如下特性：

1. 小而简约
2. 使用Go实现
3. 不需要单独部署
4. 支持我们的数据结构

它的[Github](https://github.com/boltdb/bolt)中这样描述

> Bolt is a pure Go key/value store inspired by Howard Chu's LMDB project. The goal of the project is to provide a simple，fast, and reliable database for projects that don't require a full database server such as Postgres or MySQL. 

> Bolt受Howard Chu的LMDB项目启发，纯Golang编写的key/value数据库。应运只需要简单、快速、可靠，不需要全数据库（如Mysql）功能的项目而生。

> Since Bolt is meant to be used as such a low-level piece of functionality, simplicity is key. The API will be small and only focus on getting values and setting values. That's it.

> 使用Bolt意味着只需要用到很少的（数据库）功能，所以足够简单是关键。而它的API只专注于值的读写。

是吧，我们只要这些功能。再稍稍多赘述一点它的信息。

BoltDB是基于key/value存储，即是没有像SQL关系性数据库（MySQL、PG）那样的的表，也没有行、列。而数据只存在于Key-value结构中（和Golang的maps很像）。Key-value存放在和SQL的表功能差不多的桶（buckets）中，所以要得到值，就得知道“桶”和“key”。

还有一点比较重要的是，BoltDB是没有数据类型的，key和value都是byte型的数组。因为我们要存储Golang的结构体（比如`Block`），所以会把这些结构体序列化。我们会使用[encoding/gob](https://golang.org/pkg/encoding/gob/)来`序列/解序列化`结构体，当然也可以使用 `JSON`、`XML`、`Protocol Buffers`等方案，使用它主要是简单，而且它也是Golang库标准的一部分。

# 数据结构

在实现持久化之前，我们得先搞清楚要怎么存储，先看看Bitcoin Core是怎么搞的。

简单而言，Bitcoin Core用了两个“buckets”来储存数据：

1. `blocks` 存储了该链中所有的区块的元数据
2. `chainstate` 存储链的状态，储存当前未完成的事务信息及其它一些元数据。

各区块是存储在磁盘上独立的文件当中。这么做的机制是为了保证读取一个区块不会加载所有（或部分）区块到内存中。这个特性我们现在也不去实现它。

在 **blocks** 中，key->value对有：

> 1. 'b' + 32-byte block hash -> block index record
> 2. 'f' + 4-byte file number -> file information record
> 3. 'l' -> 4-byte file number: the last block file number used
> 4. 'R' -> 1-byte boolean: whether we're in the process of reindexing
> 5. 'F' + 1-byte flag name length + flag name string -> 1 byte boolean: various flags that can be on or off
> 6. 't' + 32-byte transaction hash -> transaction index record

翻译一下

> 1. 'b' + 32-byte 该块的hash码 -> 块索引记录
> 2. 'f' + 4-byte 文件编号 -> 文件信息记录
> 3. 'l' -> 4-byte 文件编号: 最后一块文件的编号
> 4. 'R' -> 1-byte 布尔值: 标记是否正在重置索引
> 5. 'F' + 1-byte 标记名长度 + 标记名 -> 1 byte boolean: 各种可关可开的标记
> 6. 't' + 32-byte 交易的hash值 -> 交易的索引记录

在 **chainstate**, key->value对有：

> 1. 'c' + 32-byte transaction hash -> unspent transaction output record for that transaction
> 2. 'B' -> 32-byte block hash: the block hash up to which the database represents the unspent transaction outputs

翻译一下

> 1. 'c' + 32-byte 交易的hash值 -> 未完成的交易记录
> 2. 'B' -> 32-byte 块hash值: 块的hash值，直到数据库记录交易完成
 
[更为详细的解释](https://en.bitcoin.it/wiki/Bitcoin_Core_0.11_(ch_2):_Data_Storage)

因为我们现在还没有交易，所以暂时只有 **Blocks**，还有就是现在我们不把区块各自存在独立的文件中，而把整个DB当作一个文件存储Blocks。所以我们不需要任何关联到文件的数字。

所以，**Blocks**就简化成这样：

> 1. 32-byte block-hash -> Block structure (serialized)
> 2. 'l' -> the hash of the last block in a chain

下面开始实现持久化机制

# 序列化

由于BoltDB只能存储byte数组，所以先给**Block**实现序列化方法。

```golang
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	...
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
    ...
	return result.Bytes()
}
```

再实现解序列化方法

```golang
func DeserializeBlock(d []byte) *Block {
	var block Block
    ...
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
    ...
	return &block
}
```

# 持久化

我们先从优化 **NewBlockchain** 方法开始。之前这个方法只能创建新的区块链再增加创世区块到链中。现在它加上以下这些能力：

1. 打开DB文件
2. 检测是否已经有区块链存在
3. 如果存在
   1. 创建新**区块链**实例
   2. 把刚建的这个区块链信息的作为最后一块区块hash塞到DB中。
4. 如果不存在
   1. 创建新的创世区块
   2. 存储到DB中
   3. 把创世区块的hash作为末端hash
   4. 创建新的区块链，把它的信息指向创世区块

转化为代码：

```golang
func NewBlockchain() *Blockchain {
	var tip []byte

	db, err := bolt.Open(dbFile, 0600, nil)
    ...
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			genesis := NewGenesisBlock()
			b, err := tx.CreateBucket([]byte(blocksBucket))
			err = b.Put(genesis.Hash, genesis.Serialize())
			err = b.Put([]byte("l"), genesis.Hash)
			tip = genesis.Hash
		} else {
			tip = b.Get([]byte("l"))
		}

		return nil
	})

	bc := Blockchain{tip, db}

	return &bc
}
```

分析一下代码

```golang
db, err := bolt.Open(dbFile, 0600, nil)
```

这是打开BoltDB数据库文件的标准方式，**切记：即使没有找到文件，也不会返回错误**

```golang
err = db.Update(func(tx *bolt.Tx) error {
...
})
```

操作BoltDB需要使用一个参数为事务的回调函数。这里的事务有两种类型--**read-only**，**read-write**。因为我们会把创世区块放到DB中，所以我们使用**read-write**的事务，也就是`db.Update(...)`

```golang
b := tx.Bucket([]byte(blocksBucket))

if b == nil {
	genesis := NewGenesisBlock()
	b, err := tx.CreateBucket([]byte(blocksBucket))
	err = b.Put(genesis.Hash, genesis.Serialize())
	err = b.Put([]byte("l"), genesis.Hash)
	tip = genesis.Hash
} else {
	tip = b.Get([]byte("l"))
}
```
这一段是核心，先获取一个`Bucket`用来存储区块：如果桶存在，那么读取 **l**值；如果不存在，则创建创世区块，再创建桶，然后把块扔到桶里，把块的hash值设为 **l** 值。

还有注意新那区块链的方式：

```golang
bc := Blockchain{tip, db}
```

这里不再把所有的区块放到区块链中，而是只设置区块的提示信息和db的**连接**（因为在整个程序运行时，区块链会一直保持与数据库的连接）。所以，区块链的结构会被改成：

```golang
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}
```

下一步是修改 **AddBlock**方法，增加新的区块不再像之前直接把数据传过去那么简单了，现在要把区块存储到db中：

```golang
func (bc *Blockchain) AddBlock(data string) {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})

	newBlock := NewBlock(data, lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		err = b.Put([]byte("l"), newBlock.Hash)
		bc.tip = newBlock.Hash

		return nil
	})
}
```

逐段分析一下：

```golang
err := bc.db.View(func(tx *bolt.Tx) error {
	b := tx.Bucket([]byte(blocksBucket))
	lastHash = b.Get([]byte("l"))

	return nil
})
```

这里使用的是 **read-only事务**的 **Get** 方法，从l中读取最后一块区块的编码，我们挖下一新块时会作为参数用到。

```golang
newBlock := NewBlock(data, lastHash)
b := tx.Bucket([]byte(blocksBucket))
err := b.Put(newBlock.Hash, newBlock.Serialize())
err = b.Put([]byte("l"), newBlock.Hash)
bc.tip = newBlock.Hash
```

在挖出新块，将其序列化存储到数据库后，把最新的区块hash值更新到 **l** 值中。

# 检查区块

到这一步，区块都保存到数据库了，现在可以把区块链重新加载然后把新块加到里面。但是现在不能再打印区块链中的区块了，因为已经不是把区块保存在数组中了。现在修复这个缺陷。

BoltDB支持遍历一个桶中的所有key，但是这些key都是基于byte-sorted顺序排序的，而我们需要让它们按在区块中的顺序打印出来，我们也不加载所有的区块到内存中（区块可能会很大，没有必要加载完，或者，假装加载完了），先一个一个读取。现在需要一个blockchain的遍历器：

```golang
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}
```

在每次我们要去遍历整个区块链中的区块时会创建一个该遍历器。遍历器会保存当前遍历到的区块hash和保持与数据库的链接，后者也使得遍历器和该区块链在逻辑上是结合的，因为遍历器数据库连接用的是区块链的同一个，所以，**Blockchain** 会负责创建遍历器：

```golang
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}

	return bci
}
```

注意遍历器用区块链的顶端tip初始化，因此，区块是从顶端到末端，也就是从最老的区块到最新区块。事实上，**选择这个tip意味着给区块链“投票”**。一个区块链会有很多分支，而最长的那支会被认为是主分支。在获致到tip（可以是该区块链中的任何一个区块）之后，就可以重建整个区块链，算出它的长度和重建这个区块的工作量。所以，tip也可以认为是区块链的一个标识符。

**BlockchainIterator** 只做一件事：它负责返回区块链中的下一个区块：

```golang
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)

		return nil
	})

	i.currentHash = block.PrevBlockHash

	return block
}
```

到此，整DB小节完了

# CLI 命令行接口

直到现在，我们的实现还没有提供任何操作接口给外界使用。我们先前的例子中在 **main** 函数中执行新建区块链 **NewBlockchain**，还有新增区块 **bc.AddBlock** 的方法。现在可以改善，增加命令行操作接口了。我们需要如下这样的命令：

```shell
$ blockchain_go addblock "Pay 0.031337 for a coffee"
$ blockchain_go printchain
```

所有的命令行关联的操作会被 **CLI** 结构处理：

```golang
type CLI struct {
	bc *Blockchain
}
```

在 **Run**中加入CLI的接入点：

```golang
func (cli *CLI) Run() {
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	addBlockData := addBlockCmd.String("data", "", "Block data")

	switch os.Args[1] {
	case "addblock":
		err := addBlockCmd.Parse(os.Args[2:])
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			os.Exit(1)
		}
		cli.addBlock(*addBlockData)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}
```

使用标准的 **flag** 解析这些参数

```golang
addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
addBlockData := addBlockCmd.String("data", "", "Block data")
```

首先，创建两个子命令， **addblock** 和 **printfchain**，用 **-add** 标记作为 **addblock** 的参数数据标识。**printfchain** 不用参数：

```golang
switch os.Args[1] {
case "addblock":
	err := addBlockCmd.Parse(os.Args[2:])
case "printchain":
	err := printChainCmd.Parse(os.Args[2:])
default:
	cli.printUsage()
	os.Exit(1)
}
```

检测用户输入的参数和解析相关的 **flag** 子命令。

```golang
if addBlockCmd.Parsed() {
	if *addBlockData == "" {
		addBlockCmd.Usage()
		os.Exit(1)
	}
	cli.addBlock(*addBlockData)
}

if printChainCmd.Parsed() {
	cli.printChain()
}
```

解析出的子命令该执行的相关函数：

```golang
func (cli *CLI) addBlock(data string) {
	cli.bc.AddBlock(data)
	fmt.Println("Success!")
}

func (cli *CLI) printChain() {
	bci := cli.bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}
```

现在的代码很像我们之些写的那些。比较不同的是现在使用的是 **BlockchainIterator**去遍历整个区块链中的区块。

最后修改 **main** 函数：

```golang
func main() {
	bc := NewBlockchain()
	defer bc.db.Close()

	cli := CLI{bc}
	cli.Run()
}
```

注意，第一次执行时，如果BoltDB中没有区块链，则无论输入什么参数，都会创建一个区块链。

现在可以检测一下我们的代码是否工作OK了：

先安装BoltDB
```shell
$ go get github.com/boltdb/bolt/...
```
执行程序：

```shell
$ blockchain_go printchain
No existing blockchain found. Creating a new one...
Mining the block containing "Genesis Block"
000000edc4a82659cebf087adee1ea353bd57fcd59927662cd5ff1c4f618109b

Prev. hash:
Data: Genesis Block
Hash: 000000edc4a82659cebf087adee1ea353bd57fcd59927662cd5ff1c4f618109b
PoW: true

$ blockchain_go addblock -data "Send 1 BTC to Ivan"
Mining the block containing "Send 1 BTC to Ivan"
000000d7b0c76e1001cdc1fc866b95a481d23f3027d86901eaeb77ae6d002b13

Success!

$ blockchain_go addblock -data "Pay 0.31337 BTC for a coffee"
Mining the block containing "Pay 0.31337 BTC for a coffee"
000000aa0748da7367dec6b9de5027f4fae0963df89ff39d8f20fd7299307148

Success!

$ blockchain_go printchain
Prev. hash: 000000d7b0c76e1001cdc1fc866b95a481d23f3027d86901eaeb77ae6d002b13
Data: Pay 0.31337 BTC for a coffee
Hash: 000000aa0748da7367dec6b9de5027f4fae0963df89ff39d8f20fd7299307148
PoW: true

Prev. hash: 000000edc4a82659cebf087adee1ea353bd57fcd59927662cd5ff1c4f618109b
Data: Send 1 BTC to Ivan
Hash: 000000d7b0c76e1001cdc1fc866b95a481d23f3027d86901eaeb77ae6d002b13
PoW: true

Prev. hash:
Data: Genesis Block
Hash: 000000edc4a82659cebf087adee1ea353bd57fcd59927662cd5ff1c4f618109b
PoW: true
```



# 本章总结

本章我们实现了区块的持久化，还有完善了遍历信息来支持按序打印所有的区块。下一章我们将会实现 **address**，**wallet**，**transaction**。敬请期待！

# 相关链接

[本文代码][本文代码]

[二三章代码差异](https://github.com/printfcoder/blockchain-abc/compare/part_2...part_3)

[bitcoin_pdf][bitcoin_pdf]

[Bitcoin Core](https://github.com/bitcoin/bitcoin)

[Bitcoin 存储](https://en.bitcoin.it/wiki/Bitcoin_Core_0.11_(ch_2):_Data_Storage)

[boltDB](https://github.com/boltdb/bolt)

本序列文章：

1. [Golang 区块链入门 第一章 基本概念][本序列第一篇]
2. [Golang 区块链入门 第二章 工作量证明][本序列第二篇]
3. [Golang 区块链入门 第三章 持久化和命令行接口][本序列第三篇]
4. [Golang 区块链入门 第四章 交易 第一节][本序列第四篇]
5. [Golang 区块链入门 第五章 地址][本序列第五篇]
6. [Golang 区块链入门 第六章 交易 第二节][本序列第六篇]

[原文]: https://jeiwan.cc/posts/building-blockchain-in-go-part-3/
[bitcoin_pdf]: https://bitcoin.org/bitcoin.pdf
[本文代码]: https://github.com/printfcoder/blockchain-abc/tree/part_3

[本序列第一篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/05/abc-building-blockchain-in-go-part-1-basic-prototype/
[本序列第二篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/06/abc-building-blockchain-in-go-part-2-proof-of-work/
[本序列第三篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/07/abc-building-blockchain-in-go-part-3-persistence-and-cli/
[本序列第四篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/09/abc-building-blockchain-in-go-part-4-transactions-1/
[本序列第五篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/14/abc-building-blockchain-in-go-part-5-address/
[本序列第六篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/17/abc-building-blockchain-in-go-part-6-transactions-2/