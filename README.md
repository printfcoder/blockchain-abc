# Golang 区块链入门 第四节 交易第一章

交易是Bitcoin比特币中的核心，而区块链的目标就是用安全可靠的方式存储交易，要使得没有人可以在交易和区块一旦被创建后再也不能被任何人篡改。本节我们开始实现交易，但是由于交易是区块链中相当大的课题，这里分成两个部分：本章，只实现普通的交易机制。第二章才会由简入深。

[原文][原文]（略有删改）

## There is no spoon <sup><a href="#there_is_no_spoon_mean">[1]</a><sup>

如果你曾经做过关于交易的web应用，那么会应该会创建类似的两张表，**account**，**transaction**。account表用于存放用户信息和余额，而金额交易记录会存在transaction表。而在比特币中，支付是完全不同的方式：

1. 没有账户
2. 没有余额
3. 没有地址
4. 没有货币
5. 没有支付方和收款方

因为区块链是公用和开放的数据库，所以并不会存放敏感的有关钱包的数据。货币并不在账户中，交易也不是把钱从一个地址转到另一个地址。也没有字段或属性来保存账户的余额。只有交易本身，那又有什么在交易里呢？所以，这将会摧毁生活给我们树下的交易固有概念，也就是说 **There is no spoon**。

## 比特币交易

比特币的交易结构中，input与output是在一起的（[input与output][input与output]进一步阐述）：

```golang
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}
```

新交易input会关联到前一笔output（有例外，稍后补充）。output是比特币真实存储的地方。下面的这张图展示了交易的关系：

![](https://printfcoder.github.io/myblog/assets/images/blockchain/abc/transactions-diagram.png)

注意：

1. 有output是没有与input关联的
2. 在一笔交易中，input可以与不同的交易中的output相关联。
3. 而input一定是会关联一笔output的

本章全篇，我们使用了“钱”、“币”、“消费”、“发送”、“账户”等等，而比特币里是没有这些概念的。交易中（比特币机制）会使用脚本（[script][script]）锁住相关的值，然后也只有加锁的才能解开这锁。

## 交易output

从output的结构开始：

```golang
type TXOutput struct {
	Value        int
	ScriptPubKey string
}
```

事实上，output保存了“币”（上面的**Value**）。保存的意思是使用一串无法破解的方式（谜，puzzle）锁住这些币，这个puzzle就存储在**ScriptPubKey**中。在内部，Bitcoin使用了一种叫做*Script*的脚本语言，用这个Script来定义output锁和解锁的逻辑。这个语言是相当原始的，故意这样做是为了避免被攻击和滥用，但是这里不进行深一步的讨论。可以在[这里][script]找到更详细的解释。

> In Bitcoin, the value field stores the number of satoshis, not the number of BTC. A satoshi is a hundred millionth of a bitcoin (0.00000001 BTC), thus this is the smallest unit of currency in Bitcoin (like a cent).

> 在比特币中，value保存了satoshis的数量，并不是BTC的值。一个satoshis就是一亿分之一个BTC，所以这是比特币当前最小的单位（差不多是相当于分）

因为我们现在还没有实现地址（address)，所以我们会避免整个和脚本有关的逻辑。**ScriptPubKey**也会随便插入一个字符串（用户定义的钱包地址）。

> 顺便说一句，使用脚本语言意味着比特币可以也作为智能合约平台。

还有一个重要的事情是output是不能分隔的，所以你不能只引用它的一部分。如果一个output在一个交易中被关联，那么它就会全部消费掉。而如果该output的值是大于交易所需的，那么会有一笔“change”产生并返回发送者（消费者）。这和现实生活中的交易是差不多的，比如花5美元的纸币去买值1美元的东西，那你会收到4美元的找零。

## 交易input

input的结构

```golang
type TXInput struct {
	Txid      []byte
	Vout      int
	ScriptSig string
}
```

先前提到，input引用了前面的output。**Txid**存储了交易的id，而**Vout**则保存该交易的中一个output索引。**ScriptSig**就是负责提供在与output的**ScriptPubKey**中对比的数据，如果数据正确，那么这个被引用的output就可以被解锁，而它里面的值可以产生新的output。如果不正确，这个output就不能被这个input引用。这个机制就避免了有人会去消费别人的比特币。

再强调一点，因为我们还没有地址（address)，**ScriptSig**仅只是保存了一个任意的用户定义的钱包地址。我们将在下一章中实现公钥和签名检测。

总结一下，output就是“币”存的位置。每一个output都来自一个解锁了的script，这人script决定了解锁这个output的逻辑。每一个新的交易都必须有一个input和output。而input关联的前面的交易中的output，并且提供数据（**ScriptSig**字段）去解锁output和它里面的币而后用这些币去创建新的output。

那接下来，是先有input还是output呢？

## 先有蛋再有鸡

在比特币的世界里，是先鸡再有蛋。输入关联输出的逻辑（ inputs-referencing-outputs logic ）就是经典的“先有鸡还是先有蛋”问题的情况：由input生成output，然后output使得input的过程行得通。而在比特币中，output比input出现得早，input是鸡，output是蛋。

当矿机开始去挖一个区块时，它增加了**[coinbase][Coinbase] transaction**的交易。而“coinbase transaction”是一种特殊类型的交易，它不需要任何output。它会无中生有output（比如：“币”）。从而蛋不是鸡生的。这是给矿工挖出新区块的奖励。

前面的章节里提到的**创世区块**就是整个区块链的起始点。就是这个创世区块在区块链中生成了第一个output。因为没有更早的交易，所以没有更早的output。

创建coinbase的交易：

```golang
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{[]byte{}, -1, data}
	txout := TXOutput{subsidy, to}
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()

	return &tx
}
```

一个coinbase交易只能有一个input。在我们的实现里，**Txid**是空的，而**Vout**是-1。另外，coinbase也不需要存储**ScriptSig**。相反，有任意的数据存储在这里。

> In Bitcoin, the very first coinbase transaction contains the following message: “The Times 03/Jan/2009 Chancellor on brink of second bailout for banks”. [You can see it yourself][first_transaction].

> 比特币中， 最新的coinbase交易消息里有这么一段：“[《泰晤士报》，2009年1月3日，财政大臣正站在第二轮救助银行业的边缘][first_transaction]”。

**subsidy**补贴就是奖励的数量。在比特币中，这个数字并没有保存在任何地方，也仅是通过区块的总数计算出来：区块的总数除以**210000**。挖出创世区块价值50个BTC，每210000块区块被挖出，比特币单位产量就会减半（210001块到420000块时，只值25BTC了）。在我们的实现中，我们将会用一个常量来存储这个奖励（目前来说是如此😉）。


## 保存交易

现在，每个区块都必须至少保存一笔交易，并且再也不可能不通过交易而挖出新区块。这意味着我们应该删除**Block**类中的**Data**字段，换成**Transactions**。

```golang
type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}
```

**NewBlock**及**NewGenesisBlock**也要相应作更改。

```golang
func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0}
	...
}

func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}
```

下一个改动的是创建新区块链：

```golang
func CreateBlockchain(address string) *Blockchain {
	...
	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)

		b, err := tx.CreateBucket([]byte(blocksBucket))
		err = b.Put(genesis.Hash, genesis.Serialize())
		...
	})
	...
}
```

**CreateBlockchain**函数使用将存放挖出创世区块的地址**address**

## 工作量证明

“Proof-of-Work”算法必须考虑到存储在区块中的交易，在区块链中，对于存储交易的地方，要保证一致性而可靠性。所以要修改一下**prepareData**方法。

```golang
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransactions(), // This line was changed
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)
	return data
}
```

现在不能使用**pow.block.Data**了，得使用**pow.block.HashTransactions()**：

```golang
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}
```

我们再一次使用hash作为提供数据唯一表现的机制。必须保证所有交易在区块中都有确定唯一的hash值。为了实现这一点，我们计算每一个交易的hash，把它们连接起来，再计算合起来的hash。

> Bitcoin uses a more elaborate technique: it represents all transactions containing in a block as a [Merkle tree][Merkle_tree] and uses the root hash of the tree in the Proof-of-Work system. This approach allows to quickly check if a block contains certain transaction, having only just the root hash and without downloading all the transactions.

> 比特币使用了更加精细的技术：把所有交易都维护在一棵[默克尔树][Merkle_tree]中，并“Proof-of-Work”工作量证明中使用树根的hash值。这样做可以快速检测是否区块包含有指定的交易，仅需要树的根节点而不需要下载整棵树。

## Output结余

现在需要找出交易中output的结余（UTXO， unspent transaction outputs）。Unspent（结余）意思是这些output并没有关联到任何input，在上面的那张图中，有：

1. tx0, output 1;
2. tx1, output 0;
3. tx3, output 0;
4. tx4, output 0.

当然，我们需要检测余额，并不需要检测上面的全部，只需要检测那些我们的私钥能解锁的output（我们目前没有实现密钥，通过使用用户定义的地址作为替代）。现在定义在input和output上增加加锁和解锁方法：

```golang
func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}
```

我们简单地通过比较script的字段来判断是否能解锁。我们会在后面的章节中，等实现了基于私钥创建地址，再实现真正的加解锁。

下一步，找到有结余output的交易，这个比较麻烦：

```golang
func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
  var unspentTXs []Transaction
  spentTXOs := make(map[string][]int)
  bci := bc.Iterator()

  for {
    block := bci.Next()

    for _, tx := range block.Transactions {
      txID := hex.EncodeToString(tx.ID)

    Outputs:
      for outIdx, out := range tx.Vout {
        // Was the output spent?
        if spentTXOs[txID] != nil {
          for _, spentOut := range spentTXOs[txID] {
            if spentOut == outIdx {
              continue Outputs
            }
          }
        }

        if out.CanBeUnlockedWith(address) {
          unspentTXs = append(unspentTXs, *tx)
        }
      }

      if tx.IsCoinbase() == false {
        for _, in := range tx.Vin {
          if in.CanUnlockOutputWith(address) {
            inTxID := hex.EncodeToString(in.Txid)
            spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
          }
        }
      }
    }

    if len(block.PrevBlockHash) == 0 {
      break
    }
  }

  return unspentTXs
}
```
因为交易是被存储在区块中的，我们必须去检测区块链中的每一区块。

我们从output开始：

```golang
if out.CanBeUnlockedWith(address) {
	unspentTXs = append(unspentTXs, tx)
}
```

如果锁住output的地址和我们传进来的一样，那么我们要找的就是该output。但是在这之前，得检测output是否已经被input引用：

```golang
if spentTXOs[txID] != nil {
	for _, spentOut := range spentTXOs[txID] {
		if spentOut == outIdx {
			continue Outputs
		}
	}
}
```

跳过已经被input引用的，因为这些值已经被移动到其它output中，导致我们不能再去计算它。在检测output后，我们收集了所有能解锁对应地址output的input（这里不适用于coinbase交易，因为它不需要解锁output）：

```golang
if tx.IsCoinbase() == false {
    for _, in := range tx.Vin {
        if in.CanUnlockOutputWith(address) {
            inTxID := hex.EncodeToString(in.Txid)
            spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
        }
    }
}
```

为了计算余额，还需要能把**FindUnspentTransactions**返回的transaction中的output剥出来：

```golang
func (bc *Blockchain) FindUTXO(address string) []TXOutput {
       var UTXOs []TXOutput
       unspentTransactions := bc.FindUnspentTransactions(address)

       for _, tx := range unspentTransactions {
               for _, out := range tx.Vout {
                       if out.CanBeUnlockedWith(address) {
                               UTXOs = append(UTXOs, out)
                       }
               }
       }

       return UTXOs
}
```

再给CIL增加**getBalance**指令：

```golang
func (cli *CLI) getBalance(address string) {
	bc := NewBlockchain(address)
	defer bc.db.Close()

	balance := 0
	UTXOs := bc.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

```

账户余额就是有结余的交易中被账户地址锁住的output的value总和。

检测一下挖出创世区块时的余额:

```shell
$ blockchain_go getbalance -address Ivan
```

创世区块给我们带来了10个BTC的收益。

## 发送币

现在，我们要把币送给其它人。为了实现这个，需要创建一笔交易，把它设到区块中，然后挖出这个区块。到目前为止，我们的代码也只是实现了coinbase交易，现在需要一个普通的交易。

```golang
func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	acc, validOutputs := bc.FindSpendableOutputs(from, amount)

	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}

	// Build a list of inputs
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)

		for _, out := range outs {
			input := TXInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	// Build a list of outputs
	outputs = append(outputs, TXOutput{amount, to})
	if acc > amount {
		outputs = append(outputs, TXOutput{acc - amount, from}) // a change
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}
```

在创建新的output前，首先得找到所有有结余的output，并且要有足够的值来消费。**FindSpendableOutputs**方法负责做这事。然后，对于找到的能用的每一个ouput，都会有一个input关联它们。下一步，我们创建两个output：

1. 一个被接收者的地址锁住。这个output是真正的被传送到其它地址的币。
2. 一个被发送者的地址锁住。这个是找零（change）。仅是在进行结余的output的总额大于需要发送给接收者所需值的交易时才会被创建。还有，output是**不可以分隔的**；

**FindSpendableOutputs**基于前面定义的**FindUnspentTransactions**方法：

```golang
func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}
```

该方法遍历所有有结余的交易，汇总它们的值，当汇总的值等于或大于需要传送到其它地址的值时，就会停止查找，立即返回已经汇总到的值和以交易id分组的output索引数组。不需要找到比本次传送额更多的output。

现在修改**Blockchain.MineBlock**方法：

```golang
func (bc *Blockchain) MineBlock(transactions []*Transaction) {
	...
	newBlock := NewBlock(transactions, lastHash)
	...
}
```

最后，实现**Send**方法：

```golang
func (cli *CLI) send(from, to string, amount int) {
	bc := NewBlockchain(from)
	defer bc.db.Close()

	tx := NewUTXOTransaction(from, to, amount, bc)
	bc.MineBlock([]*Transaction{tx})
	fmt.Println("Success!")
}
```

传送币到其它地址，意味着会创建新的交易，然后会通过挖出新的区块，把交易放到该区块中，再把该区块放到区块链的方式让交易得以在区块链中。但是区块链并不会立即做到这一步，相反，它把所有的交易放到存储池中，当矿机准备好挖区块时，它就把存储池中的所有交易拿出来并创建候选的区块。交易只有在包含了该交易的区块被挖出且附加到区块链中时才会被确认。

现在看看传送币的工作是否正常：

```shell
$ blockchain_go send -from Ivan -to Pedro -amount 6
00000001b56d60f86f72ab2a59fadb197d767b97d4873732be505e0a65cc1e37

Success!

$ blockchain_go getbalance -address Ivan
Balance of 'Ivan': 4

$ blockchain_go getbalance -address Pedro
Balance of 'Pedro': 6
```

再创建几笔交易，然后确认多个output在花费过程中是否工作正常：

```golang
$ blockchain_go send -from Pedro -to Helen -amount 2
00000099938725eb2c7730844b3cd40209d46bce2c2af9d87c2b7611fe9d5bdf

Success!

$ blockchain_go send -from Ivan -to Helen -amount 2
000000a2edf94334b1d94f98d22d7e4c973261660397dc7340464f7959a7a9aa

Success!
```

Helen的币被两个output锁（只有自己的地址才能解锁）在了两个output中，一个是Pedro，另一个是Ivan。现在再传给其他人：

```shell
$ blockchain_go send -from Helen -to Rachel -amount 3
000000c58136cffa669e767b8f881d16e2ede3974d71df43058baaf8c069f1a0

Success!

$ blockchain_go getbalance -address Ivan
Balance of 'Ivan': 2

$ blockchain_go getbalance -address Pedro
Balance of 'Pedro': 4

$ blockchain_go getbalance -address Helen
Balance of 'Helen': 1

$ blockchain_go getbalance -address Rachel
Balance of 'Rachel': 3
```

现在Pedro只有4个币了，再尝试把向Ivan传送5个：

```shell
$ blockchain_go send -from Pedro -to Ivan -amount 5
panic: ERROR: Not enough funds

$ blockchain_go getbalance -address Pedro
Balance of 'Pedro': 4

$ blockchain_go getbalance -address Ivan
Balance of 'Ivan': 2
```

正常～


## 本章总结

呼！不是很容易，至少现在有交易了。尽管关键的特性像比特币那样的加密货币还没有实现：

1. 地址。我们没有实现真正的地址，基于私钥的地址。
2. 奖励。现在挖出区块是没有甜头的。
3. UTXO 集合。获取余额需要查找整个区块，如果有很多的区块链时需要花费非常长的时间。并且，要验证后续的交易，也会花费大量的时间。UTXO集合就是为了解决这个问题，让对整个交易的操作更快些。
4. 存储池（Mempool）。这里保存那些等着被打包到区块中的交易。在我们的当前的实现里，一个区块只有一个交易，这很没有效率。




## 相关链接

[本文代码][本文代码]

[bitcoin script][script]

[交易](https://en.bitcoin.it/wiki/Transaction)

[默克尔树][Merkle_tree]

[Coinbase][Coinbase]

本序列文章：

1. [Golang 区块链入门 第一章 基本概念][本序列第一篇]
2. [Golang 区块链入门 第二章 工作量证明][本序列第二篇]
3. [Golang 区块链入门 第三章 持久化和命令行接口][本序列第三篇]
4. [Golang 区块链入门 第四章 交易 第一节][本序列第四篇]
5. [Golang 区块链入门 第五章 地址][本序列第五篇]
6. [Golang 区块链入门 第六章 交易 第二节][本序列第六篇]


<div id="there_is_no_spoon_mean">
<sup>[1]</sup>
<span>这一句是从黑客帝国里借鉴而来，不知道怎么翻译才不失味道，心无外物。大概是说，区块链（或是比特币的交易）并非我们普通的交易那样子</span>
</div>

[原文]: https://jeiwan.cc/posts/building-blockchain-in-go-part-4/
[bitcoin_pdf]: https://bitcoin.org/bitcoin.pdf
[本文代码]: https://github.com/printfcoder/blockchain-abc/tree/part_4

[input与output]: /myblog/blockchain/bitcoin/2018/03/10/how-shall-we-understand-the-input-and-output-of-bitcoin/
[script]: https://en.bitcoin.it/wiki/Script

[first_transaction]: https://blockchain.info/tx/4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b?show_adv=true
[Merkle_tree]: https://en.wikipedia.org/wiki/Merkle_tree
[Coinbase]: https://en.bitcoin.it/wiki/Coinbase

[本序列第一篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/05/abc-building-blockchain-in-go-part-1-basic-prototype/
[本序列第二篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/06/abc-building-blockchain-in-go-part-2-proof-of-work/
[本序列第三篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/07/abc-building-blockchain-in-go-part-3-persistence-and-cli/
[本序列第四篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/09/abc-building-blockchain-in-go-part-4-transactions-1/
[本序列第五篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/14/abc-building-blockchain-in-go-part-5-address/
[本序列第六篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/17/abc-building-blockchain-in-go-part-6-transactions-2/
