# Golang 区块链入门 第六章 交易2

在本系列最早的一篇文章里我说过区块链就是一个分布式数据库。后来，我们决定跳开“分布式”的部分而把注意力放到”数据库“上。直到现在，我们实现了区块链数据库的大部分工作。在这篇文章中，我们将会把之间跳过的一些特性覆盖掉，下面的部分我们将会开始用区块链的分布式自然特性。

前面的章节:
1. [Golang 区块链入门 第一章 基本概念][本序列第一篇]
2. [Golang 区块链入门 第二章 工作量证明][本序列第二篇]
3. [Golang 区块链入门 第三章 持久化和命令行接口][本序列第三篇]
4. [Golang 区块链入门 第四章 交易 第一节][本序列第四篇]
5. [Golang 区块链入门 第五章 地址][本序列第五篇]

[原文][原文]（略有删改）

> 本节的阐述会有重大的代码改变，如果在这里讲就有点麻烦了。请跳到[这里](https://github.com/printfcoder/blockchain-abc//compare/part_5...part_6#files_bucket)来看所有的改变。

## Reward（奖励）

有一件前面的章节中跳过了一个小细节，挖矿奖励。现在我们准备实现这个。

这个奖励也就是coinbase交易。当一个节点开始挖新的区块时，它会把队列中的交易并准备好coinbase交易放到区块中。这笔coinbase交易也仅仅是一个包含了矿工的公钥hash的output。

实现奖励很简单，只用更新一下**send**命令：

```golang
func (cli *CLI) send(from, to string, amount int) {
    ...
    bc := NewBlockchain()
    UTXOSet := UTXOSet{bc}
    defer bc.db.Close()

    tx := NewUTXOTransaction(from, to, amount, &UTXOSet)
    cbTx := NewCoinbaseTX(from, "")
    txs := []*Transaction{cbTx, tx}

    newBlock := bc.MineBlock(txs)
    fmt.Println("Success!")
}
```
在我们的实现中，创建交易的人挖出了新的区块，得到奖励。

## UTXO Set

在第三章[持久化和命令行接口][本序列第三篇]中，我们学习了比特币中存储区块到数据库的方式。文中提到区块被存放在**blocks** 数据库，交易output存放在**chainstate**数据库中。这里说一下**chainstate**的结构：

> 1. 'c' + 32-byte transaction hash -> unspent transaction output record for that transaction
> 2. 'B' -> 32-byte block hash: the block hash up to which the database represents the unspent transaction outputs

翻译一下

> 1. 'c' + 32-byte 交易的hash值 -> 未完成的交易记录
> 2. 'B' -> 32-byte 块hash值: 数据库记录的未使用的交易的output的块hash
 
第三篇文章里我们已经实现了交易，但是没有使用**chainstate**来保存他们的output，现在来实现这个。

**chainstate**不存放交易，相反，它保存UTXO（unspent transaction outputs，有结余交易的output）集合。除此之外，它保存“数据库记录的未使用的交易的output的块hash”，我们会忽略这个特性，因为我们没有使用区块的高度（下一篇里会讨论实现）。

那为什么我要有**UTXO**集合？

考虑到我们此前实现的方法**Blockchain.FindUnspentTransactions**：
```golang
func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
    ...
    bci := bc.Iterator()

    for {
        block := bci.Next()

        for _, tx := range block.Transactions {
            ...
        }

        if len(block.PrevBlockHash) == 0 {
            break
        }
    }
    ...
}
```
该函数负责把有未消费完output的交易找出来。因为交易是存放在区块中的，所以这个方法会迭代所有的区块链中的区块，并检测区块中的每一个交易。到2017年9月18号，比特币中已经有485860个区块，而全部的数据用了140+GB的磁盘空间。这意为着如果要验证交易，则要检测所有的节点。而且，验证交易将需要遍历很多区块。

而解决方案是要给未消费完的output建立索引，这就是UTXO的作用：这个缓存是基于所有区块链中交易（通过遍历了所有的区块，当然了，只执行了一次）创建的，然后就用来计算余额和验证新的交易。这个UTXO的大小在2017年9月大概是有2.5Gb。

好了，我们要想一下要如何改造UTXO的实现方法。当前，下面这些方法是用于查找交易的：

1. Blockchain.FindUnspentTransactions 找到所有含有未消费output的交易主函数。遍历所有的区块在该函数里执行。
2. Blockchain.FindSpendableOutputs 当有新的交易创建时使用。如果找足够交易所需数的output。会调用**Blockchain.FindUnspentTransactions**方法
3. Blockchain.FindUTXO 找到未消费的output来创建公钥hash，调用 **Blockchain.FindUnspentTransactions**方法。
4. Blockchain.FindTransaction 通过交易的ID在区块链中找到交易。它会遍历所有区块直到找到该交易。

可以看到，这些方法遍历了整个数据库中的所有区块。但是现在我们不能改善这些方法，因为UTXO集合没有存放在所有的交易，而只有那些含未消费output的。因此，还不能在**Blockchain.FindTransaction**使用。

所以我们需要下面的这些方法：

1. Blockchain.FindUTXO 通过遍历所有区块找到所有未消费的output
2. UTXOSet.Reindex 调用**FindUTXO**来查找所有没有消费的output，然后存储到数据库中。这里是缓存执行的地方。
3. UTXOSet.FindSpendableOutputs 类似**Blockchain.FindSpendableOutputs**，但是使用的是UTXO集合。
4. UTXOSet.FindUTXO 类似**Blockchain.FindUTXO**，但是使用的是UTXO集合。
5. Blockchain.FindTransaction 保持不变

因此，用得最多的两个方法从现在开始就会使用缓存，开始码代码：

```golang
type UTXOSet struct {
    Blockchain *Blockchain
}
```
我们使用同一个数据库，但是把UTXO集合放到另一个桶（bucket）中。所以，**UTXOSet**和**Blockchain**是耦合的（共用了一个数据库）：

```golang
func (u UTXOSet) Reindex() {
    db := u.Blockchain.db
    bucketName := []byte(utxoBucket)

    err := db.Update(func(tx *bolt.Tx) error {
        err := tx.DeleteBucket(bucketName)
        _, err = tx.CreateBucket(bucketName)
    })

    UTXO := u.Blockchain.FindUTXO()

    err = db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket(bucketName)

        for txID, outs := range UTXO {
            key, err := hex.DecodeString(txID)
            err = b.Put(key, outs.Serialize())
        }
    })
}
```
这个方法创建并初始化UTXO集合。首先移除所有的存在的桶，然后从区块链中找到所有未消费的output，最后把这些output存到桶中去。

**Blockchain.FindUTXO**几乎与**Blockchain.FindUnspentTransactions**是相同的，但是它返回的是**TransactionID → TransactionOutputs**映射组合的map。

现在，UTXO集合可以发送币了：

```golang
func (u UTXOSet) FindSpendableOutputs(pubkeyHash []byte, amount int) (int, map[string][]int) {
    unspentOutputs := make(map[string][]int)
    accumulated := 0
    db := u.Blockchain.db

    err := db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(utxoBucket))
        c := b.Cursor()

        for k, v := c.First(); k != nil; k, v = c.Next() {
            txID := hex.EncodeToString(k)
            outs := DeserializeOutputs(v)

            for outIdx, out := range outs.Outputs {
                if out.IsLockedWithKey(pubkeyHash) && accumulated < amount {
                    accumulated += out.Value
                    unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
                }
            }
        }
    })

    return accumulated, unspentOutputs
}
```

然后检测余额：

```golang
func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
    var UTXOs []TXOutput
    db := u.Blockchain.db

    err := db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(utxoBucket))
        c := b.Cursor()

        for k, v := c.First(); k != nil; k, v = c.Next() {
            outs := DeserializeOutputs(v)

            for _, out := range outs.Outputs {
                if out.IsLockedWithKey(pubKeyHash) {
                    UTXOs = append(UTXOs, out)
                }
            }
        }

        return nil
    })

    return UTXOs
}
```

这些方法和**Blockchain**相应版本的方法相比，有些轻微的改动。而那些**Blockchain**中对应的方法就没有用了。

用了UTXO集合我们（交易的）数据就可以分开存放了：实际的交易存放在区块链中，未消费的output则存放在UTXO集合里。这样的分离需要坚固的同步机制，因为我们得让UTXO集合总是能更新和保存所有最近交易的output。但是我们不需要每次新区块挖出来时重排索引，因为我们要避免频繁的区块链查找。因此，需要一个机制来更新UTXO集合。

```golang
func (u UTXOSet) Update(block *Block) {
    db := u.Blockchain.db

    err := db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(utxoBucket))

        for _, tx := range block.Transactions {
            if tx.IsCoinbase() == false {
                for _, vin := range tx.Vin {
                    updatedOuts := TXOutputs{}
                    outsBytes := b.Get(vin.Txid)
                    outs := DeserializeOutputs(outsBytes)

                    for outIdx, out := range outs.Outputs {
                        if outIdx != vin.Vout {
                            updatedOuts.Outputs = append(updatedOuts.Outputs, out)
                        }
                    }

                    if len(updatedOuts.Outputs) == 0 {
                        err := b.Delete(vin.Txid)
                    } else {
                        err := b.Put(vin.Txid, updatedOuts.Serialize())
                    }

                }
            }

            newOutputs := TXOutputs{}
            for _, out := range tx.Vout {
                newOutputs.Outputs = append(newOutputs.Outputs, out)
            }

            err := b.Put(tx.ID, newOutputs.Serialize())
        }
    })
}
```
这个方法看上去有点大，但是它做的事还是比较简单粗暴的。当挖出新的区块时，UTXO集合就会被更新。更新意味着会清除掉被消费了的output，及增加新挖出的交易中未消费的output。如果一笔交易的output被移除了，内部也没有其它output时，它也会除掉。相当简单。

在必要的地方使用UTXO：

```golang
func (cli *CLI) createBlockchain(address string) {
    ...
    bc := CreateBlockchain(address)
    defer bc.db.Close()

    UTXOSet := UTXOSet{bc}
    UTXOSet.Reindex()
    ...
}
```

重置索引在新的区块链创建后发生才正确。现在，只有在这里**Reindex**才用到，不过由于在一开始区块链只有一个区块一笔交易，导致看上去有点用力过度，而且**Update**也不会作为代替使用。但是我们还是后面我们还是会用到重置索引机制的。

```golang
func (cli *CLI) send(from, to string, amount int) {
    ...
    newBlock := bc.MineBlock(txs)
    UTXOSet.Update(newBlock)
}
```
在新的区块挖出来后，UTXO集合就会被更新。

检测一下是否工作：

```shell
$ blockchain_go createblockchain -address 1JnMDSqVoHi4TEFXNw5wJ8skPsPf4LHkQ1
00000086a725e18ed7e9e06f1051651a4fc46a315a9d298e59e57aeacbe0bf73

Done!

$ blockchain_go send -from 1JnMDSqVoHi4TEFXNw5wJ8skPsPf4LHkQ1 -to 12DkLzLQ4B3gnQt62EPRJGZ38n3zF4Hzt5 -amount 6
0000001f75cb3a5033aeecbf6a8d378e15b25d026fb0a665c7721a5bb0faa21b

Success!

$ blockchain_go send -from 1JnMDSqVoHi4TEFXNw5wJ8skPsPf4LHkQ1 -to 12ncZhA5mFTTnTmHq1aTPYBri4jAK8TacL -amount 4
000000cc51e665d53c78af5e65774a72fc7b864140a8224bf4e7709d8e0fa433

Success!

$ blockchain_go getbalance -address 1JnMDSqVoHi4TEFXNw5wJ8skPsPf4LHkQ1
Balance of '1F4MbuqjcuJGymjcuYQMUVYB37AWKkSLif': 20

$ blockchain_go getbalance -address 12DkLzLQ4B3gnQt62EPRJGZ38n3zF4Hzt5
Balance of '1XWu6nitBWe6J6v6MXmd5rhdP7dZsExbx': 6

$ blockchain_go getbalance -address 12ncZhA5mFTTnTmHq1aTPYBri4jAK8TacL
Balance of '13UASQpCR8Nr41PojH8Bz4K6cmTCqweskL': 4
```

**1JnMDSqVoHi4TEFXNw5wJ8skPsPf4LHkQ1**地址收到三个奖励：

1. 挖出创世区块的奖励
2. 挖出**0000001f75cb3a5033aeecbf6a8d378e15b25d026fb0a665c7721a5bb0faa21b**区块
3. 挖出**000000cc51e665d53c78af5e65774a72fc7b864140a8224bf4e7709d8e0fa433**区块

## 默克尔树

在这里要再多讨论一个的优化机制。

前面说到，完整的区块链数据库（即区块链）花掉了140Gb的磁盘存储空间。因为去中心化的特性，每个在网络中的节点都必须独立且足够自主，也即每个节点都必须保存整个区块链的副本。随着人们开始使用比特币，这一规则就会变得困难，每个人都要运行所有节点显然是不合适的。还有，因为节点都是网络中完全成熟的部分，它们都有责任：必须验证交易和区块。另外，得能连上网络与其它节点交互和下载新的区块。

在中本始发表的最初的比特币[论文](https://bitcoin.org/bitcoin.pdf)中，已经有一个解决方案来处理这一问题，简化支付验证（Simplified Payment Verification，SPV）。SPV是一个轻量的比特币节点，不会下载整个区块连，**也不验证区块和交易**。相反，它会找出区块（用于验证交易）中的交易，且和所有的节点连接来检索必要的数据。这一机制允许运行多个轻量的钱包节点和只需一个全量节点。

为了实现SPV的可行性，得有一种方式能检测在不下载整个区块的情况下，判断该区块包含了指定的交易。为了解决这个问题，需要引入默克尔树。

默克尔树被用于比特币来获取交易hash值，该hash存放在block的头部以及在工作证明中会被用到。直到现在，我们也只是把区块中的每一个交易hash串起来，再用**SHA-256**计算它们。这当然也是获取唯一的区块中的交易描述的好方式，但是并没有默克尔树的优点。

看看默克尔树：

![](https://printfcoder.github.io/myblog/assets/images/blockchain/abc/merkle-tree-diagram.png)

默克尔树为了每一个区块而创建，开始于叶（树的底部）节点，叶子就是一个交易的hash值（比特币使用两次SHA256计算）。叶子的数量必须是偶数的，但是并不是每一个区块都含有偶数个交易。如果有奇数个交易，最后一个交易就会重复（**在默克尔树里是这样，不是区块中！**）。

从底而上，叶子被分成组成一对，它们的hash是串起来的，并且串起来的hash也会生成新的hash。新的hash生成新的树节点。这一过程会一直持续直到只剩下一个节点，也就是树的根节点。根节点的hash就会被当成这些交易的描述存放在区块的头部，然后在工作量证明中会用到。

使用默克尔树的好处就是节点可以清楚与指定交易的关系，而不需要下载整个区块。只需要一个交易hash，默克尔树的根节点hash，还有树的路径即可。

开始撸代码：

```golang
type MerkleTree struct {
    RootNode *MerkleNode
}

type MerkleNode struct {
    Left  *MerkleNode
    Right *MerkleNode
    Data  []byte
}
```
从结构开始，每一个**MerkleNode**都会持有数据和连接到它的分支。**MerkleTree**实际上就是和后面的节点连接的根节点，而它们与其它节点连接，等等。

创建新的节点：

```golang
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
    mNode := MerkleNode{}

    if left == nil && right == nil {
        hash := sha256.Sum256(data)
        mNode.Data = hash[:]
    } else {
        prevHashes := append(left.Data, right.Data...)
        hash := sha256.Sum256(prevHashes)
        mNode.Data = hash[:]
    }

    mNode.Left = left
    mNode.Right = right

    return &mNode
}
```
每一个节点包含了一些数据。当节点是叶子节点时，数据来自外方（从我们的角度看就是序列化的交易）。当节点连接到其它（左右）节点时，它就会把这左右两个节点的数据串起来，然后计算串起来的hash值作为自己的数据。

```golang
func NewMerkleTree(data [][]byte) *MerkleTree {
    var nodes []MerkleNode

    if len(data)%2 != 0 {
        data = append(data, data[len(data)-1])
    }

    for _, datum := range data {
        node := NewMerkleNode(nil, nil, datum)
        nodes = append(nodes, *node)
    }

    for i := 0; i < len(data)/2; i++ {
        var newLevel []MerkleNode

        for j := 0; j < len(nodes); j += 2 {
            node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
            newLevel = append(newLevel, *node)
        }

        nodes = newLevel
    }

    mTree := MerkleTree{&nodes[0]}

    return &mTree
}
```
当新的树创建好时，第一件事就是确定是否有偶数个叶子。然后，**数据**（data，交易的序列化数组）会被转换成树叶，并且新树会基于这些叶子长出来。

现在修改**Block.HashTransactions**，它的作用是在工作量证明时获取交易的hash：
```golang
func (b *Block) HashTransactions() []byte {
    var transactions [][]byte

    for _, tx := range b.Transactions {
        transactions = append(transactions, tx.Serialize())
    }
    mTree := NewMerkleTree(transactions)

    return mTree.RootNode.Data
}
```
首先，区块中的交易会串起来，然后被序列化（使用**encoding/gob**），然后用于创建新的默克尔树。该树的根节点会充当该区块的所有交易的标识。

## P2PKH

还有一个问题在这里讨论一下：

我们说过，比特币中使用了*Script脚本*编程语言，它被用在给交易的output加锁，然后交易的input提供数据来解锁output。这个语言很简单，代码也只是一串序列和一些操作符。

看这个例子：

```
5 2 OP_ADD 7 OP_EQUAL
```

**5**，**2**，和**7**是数据。**OP_ADD**和**OP_EQUAL**就是操作符。*Script*代码是可以从左到右被执行的，每一片代码被放到栈中，然后下一个操作符就会用于栈顶的元素。*Script*栈仅是简单的FILO方式使用内存，第一个元素进栈会最后一个出栈，后面的元素会被放到前一个的上面。

我们拆开上面的代码来逐步分析：

|序号|栈|脚本|
|--|--|--|
|1|空|5 2 OP_ADD 7 OP_EQUAL|
|2|5|2 OP_ADD 7 OP_EQUAL|
|3|5 2|OP_ADD 7 OP_EQUAL|
|4|7|7 OP_EQUAL|
|5|7 7|OP_EQUAL|
|6|true|empty|

操作**OP_ADD**就是从栈顶先后取出两个元素相加，然后把结果压到栈顶。**OP_EQUAL**同样从栈顶取出两个元素判断相等，把结果压入栈顶。当脚本执行完时，栈顶的元素就是该脚本的执行结果：在我们的这个例子中，结果就是**true**，也即是说脚本成功执行完成。

现在看看比特币中用于支付的脚本：

```
<signature> <pubKey> OP_DUP OP_HASH160 <pubKeyHash> OP_EQUALVERIFY OP_CHECKSIG
```
这种叫Pay to Public Key Hash（P2PKH），是比特币中用得最广泛的脚本。它会逐个支付给公钥的hash，即会锁住指定公钥下的币。这是**比特币的支付核心**：没有账户，彼此之间没有现金交换；也仅是有一段脚本来检测提供的签名和公钥是否正确。

这段脚本实际上保存有两个部分：

1. 第一段：**<signature> <pubKey>**存放了input的**ScriptSig**栏位。
2. 第二段：**OP_DUP OP_HASH160 <pubKeyHash> OP_EQUALVERIFY OP_CHECKSIG**存放output的**ScriptPubKey**。

因此，这段脚本定义了解锁的逻辑，就是input提供了数据来解锁output。执行这段脚本：

|序号|栈|脚本|
|--|--|--|
|1|空|**\<signature\> \<pubKey\> OP_DUP OP_HASH160 \<pubKeyHash\> OP_EQUALVERIFY OP_CHECKSIG**|
|2|**\<signature\>**|**\<pubKey\> OP_DUP OP_HASH160 \<pubKeyHash\> OP_EQUALVERIFY OP_CHECKSIG**|
|3|**\<signature\> \<pubKey\>**|**OP_DUP OP_HASH160 \<pubKeyHash\> OP_EQUALVERIFY OP_CHECKSIG**|
|4|**\<signature\> \<pubKey\> \<pubKey\>**|**OP_HASH160 \<pubKeyHash\> OP_EQUALVERIFY OP_CHECKSIG**|
|5|**\<signature\> \<pubKey\> \<pubKeyHash\>**|**\<pubKeyHash\> OP_EQUALVERIFY OP_CHECKSIG**|
|6|**\<signature\> \<pubKey\> \<pubKeyHash\> \<pubKeyHash\>**|**OP_EQUALVERIFY OP_CHECKSIG**|
|7|**\<signature\> \<pubKey\>**|**OP_CHECKSIG**|
|8|**true 或 false**|空|

**OP_DUP**操作会复制栈顶的元素。**OP_HASH160**获取栈顶的元素并使用（**RIPEMD160**）算法计算其hash值，把结果压到栈顶。**OP_EQUALVERIFY**比较栈顶的两个元素，如果不等，则打断脚本。**OP_CHECKSIG**通过计算交易的hash和使用**<signature>**及**<pubKey>**验证交易的签名。后面的操作比较复杂：使用一个被修整过的交易副本，计算它的hash（因为它就是一个被签名了的交易的hash），然后用提供的**<signature>**及**<pubKey>**检测这个签名是否正确。

拥有脚本语言的使得比特币可能成为智能合约平台：这个语言除了能支持每次交易都使用单一的密钥转移比特币，其它的支付场景也成为可能。

## 总结

好了，我们实现了大多数基于区块链加密货币的差关键特性。区块链、地址、挖矿、交易。但是，还有一个赋予这些特性生命的机制，创造了比特币的全局系统，一致性。 下一章我们开始实现区块链的一部分--“去中心化”。静请期待！

## 相关链接

[本文代码][本文代码]

本序列文章：

1. [Golang 区块链入门 第一章 基本概念][本序列第一篇]
2. [Golang 区块链入门 第二章 工作量证明][本序列第二篇]
3. [Golang 区块链入门 第三章 持久化和命令行接口][本序列第三篇]
4. [Golang 区块链入门 第四章 交易 第一节][本序列第四篇]
5. [Golang 区块链入门 第五章 地址][本序列第五篇]
6. [Golang 区块链入门 第六章 交易 第二节][本序列第六篇]

[本序列第一篇]: /myblog/blockchain/abc/2018/03/05/abc-building-blockchain-in-go-part-1-basic-prototype/
[本序列第二篇]: /myblog/blockchain/abc/2018/03/06/abc-building-blockchain-in-go-part-2-proof-of-work/
[本序列第三篇]: /myblog/blockchain/abc/2018/03/07/abc-building-blockchain-in-go-part-3-persistence-and-cli/
[本序列第四篇]: /myblog/blockchain/abc/2018/03/09/abc-building-blockchain-in-go-part-4-transactions-1/
[本序列第五篇]: /myblog/blockchain/abc/2018/03/14/abc-building-blockchain-in-go-part-5-address/
[本序列第六篇]: /myblog/blockchain/abc/2018/03/17/abc-building-blockchain-in-go-part-6-transactions-2/

[原文]: https://jeiwan.cc/posts/building-blockchain-in-go-part-6/
[本文代码]: https://github.com/printfcoder/blockchain-abc/tree/part_6
[比特币地址样例]: https://blockchain.info/address/1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa
[bip-0039]: https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki
[椭圆曲线]: http://andrea.corbellini.name/2015/05/17/elliptic-curve-cryptography-a-gentle-introduction/
