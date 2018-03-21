# Golang 区块链入门 第二节 工作量证明

在[前面一节][本序列第一篇]中我们弄了一个简单的区块链结构，大略阐述了区块与链的诞生过程，每一个区块都与前一块连接，不过，上一节中的区块链有一个瑕疵--增加区块过于快速和廉价了。大家知道，创建一个区块在区块链和比特币里是需要大量运算的，也就是`工作量证明`，英文叫`Proof-of-Work`。今天我们完善这个缺点。

[原文][原文]（略有删改）

## Proof-of-Work 工作量证明

区块里有一个非常关键的点，就是节点必须执行足够多且困难的运算才能将数据新增在区块中。这一困难的运算保证了区块链安全、一致。而为了奖励这一运算，该节点会获得数字货币（如比特币）的奖励（从运算到收到奖励的过程，也叫作挖矿）。

这一机制和现实生活中也是相似的：人们辛苦工作获取报酬来维持生活，在区块链中，链络中的参与者（比如矿工）辛苦运算来维系这个区块链网络，不断增加新的区块到链络中，然后获取回报。正是因为这些运算，新的区块基于安全的方式加到区块链中，保证了区块链数据库的稳定。

是不是发现了什么问题呢？大家都在计算，凭什么怎么证明你做的运算就是对的，且是你的。

`努力工作并证明（do hard work and prove）`，这一机制被称为`工作量证明`。需要多努力呢，需要大量计算机资源，即使使用高速计算机也不能做得快多少。而且，随着时间的推移，难度会越来越大，因为要保证每小时有6个区块的诞生，越到后面，区块越来越少，要保证这个速率，只能运算更多，提高难度。在`比特币`中，运算的目标是计算出一串符合要求的`hash`值。而这个hash就是证明。所以说，找到证明（符合要求的hash值）才是实际意义上的工作。


工作量证明还有一个重要知识点。也即`工作困难，而证明容易`。因为如果你的工作困难，而证明也困难，那么你的工作在圈子效率意义就不大，对于需要提供给别人证明的工作，别人证明起来越简单就越好。

## Hash 哈希

Hash运算是区块链最主要使用的工作算法。哈希运算是指给特殊数据计算一串hash字符的过程。对于一笔数据而言，它的hash值是唯一的。hash函数就是可以把任意大小的数据计算出指定大小hash值的函数。

Hash运算有以下几个主要的特点：
<ul>
<li>原始数据不能从hash值中逆向计算得到</li>
<li>确定的数据只有一个hash值且这个hash值是唯一的</li>
<li>改变数据的任一byte都会造成hash值变动</li>
</ul>

![](https://printfcoder.github.io/myblog/assets/images/blockchain/abc/hashing-example.png)

Hash运算广泛应用于数据一致性的验证。很多线上软件商店都会把软件包的hash值公开，用户下载后自行计算hash值后验证和供应商的是否一致，就可判断软件是否被篡改过。

在区块链中，hash也是用于保证数据的一致性。区块的数据，还有区块的前一区块的hash值也会被用于计算本区块的hash值，这样就保证每个区块是不可变：如果有人要改动自己区块的hash值，那么连他后面的区块hash也要跟着改，这显然是不可能的或者极其困难的（要说服不是自己的区块一同更改很困难）。

## Hashcash 哈希现金

比特币中的工作证明使用的是[Hashcash][Hashcash]技术，起初，这一算法开发出来就是用于防止垃圾电子邮件。它的工作主要有以下几个步骤：

1. 获取公开的信息，比如邮件的收件人地址或者比特币的头部信息
2. 增加一个起始值为0的计数器
3. 计算出`第1步中的信息+计数器值`组合的hash值
4. 按规则检测hash值是否满足需求（一般是指定前20位是0）
   1. 满足
   2. 不满足则重复3-4步骤

这个算法看上去比较暴力：改变计数器值，计算新的hash，检测，增加计数器值，计算新的hash...，所以这个算法比较昂贵。

邮件发送者预备好邮件头部信息然后附加用于计算随机数字计数值。然后计算160-bit长的hash头，如果前20bits

现在进一步分析区块链hash运算的要求。在原始的hashcash实现中，必须根据头信息算出前20位为0的hash值。而在比特币中，这一规则则是根据时间的推移变动的，因为比特币的设计就是10分钟出一块新区块，即使计算机算力提升或者更多的矿工加入挖矿行列中也不会改变，也就是说，算出hash值，会越来越困难。

下面演示这一算法，和上面第一张图一样使用“I like donuts”作为数据，再在数据后加面附加计数器，然后使用SHA256算法找出前面6位为0的hash值。

![](https://printfcoder.github.io/myblog/assets/images/blockchain/abc/hashcash-example.png)

而`ca07ca`就是不停运算找到的计数器值，转换成10进制就是`13240266`，换言之大概执行了13240266次SHA256运算才找到符合条件的值。

## 实现

上面花了点篇幅介绍了工作量证明的原理。现在我们用Golang来实现。先定24位0的作为挖矿的难度：

```golang
const targetBits = 24
```

`注`：在比特币挖矿中，头部中的`target bits`存储该区块的挖矿难度，但是上面说过随着时间推移难度越来越大，所以这个target大小是会变的，这里不实现target的适配算法，这不影响我们理解挖矿。现在只定义一个常量作为全局的难度。

当然，24也是比较随意的，我们只用在内存中占用少于256bits的的空间。差异也要大些，但是也不要太大，太大了就就很难找出来一个合规的hash。

然后定义`ProofOfWork`结构：

```golang
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{b, target}

	return pow
}
```

`ProofOfWork` 有“block”和“target”两个成员。“target”就是上面段落中描述的hashcash的规则信息。使用big.Int是因为要把hash转成大整数，然后检测是否比target要小。

`NewProofOfWork` 函数负责初始化target，将1往左偏移(256-targetBits)位。256是我们要使用的SHA-256标准的hash值长度。转换成16进制就是：

```log
0x10000000000000000000000000000000000000000000000000000000000
```

这会占用29位大小的内存空间。

现在准备创建hash的函数：

```golang
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
```

这段代码做了简化，直接把`block`的信息和`target`、`nonce`合并在一起。`nonce`就是Hashcash中的counter，nonce（现时标志）是加密的术语。

准备工作都OK了，现在实现`PoW`的核心算法：

```golang
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)
	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")

	return nonce, hash[:]
}
```
hashInt是hash值的int形式，nonce是计数器。然后执行`math.MaxInt64`次循环直到找符合target的hash值。为什么是`math.MaxInt64`次，其实这个例子中是不用的考虑这么大的，因为我们示例中的PoW太小了以致于还不会造成溢出，只是编程上要考虑一下，当心为上。

循环体内工作主要是：

1. 准备块数据
2. 计算SHA-256值
3. 转成big int
4. 与target比较

将[上一篇][本序列第一篇]中的`NewBlock`方法改造一下，扔掉`SetHash方法`：

```golang
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}
```

新增了`nonce`作为`Block`的特性。`nonce`作为证明是必须要带的。现在`Block`结构如下：

```golang
type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}
```
然后执行 `go run *.go`

```log
start: 2018-03-07 14:43:31.691959 +0800 CST m=+0.000721510
Mining the block containing "Genesis Block"

0000006f3387b588739cbcfe2cce521fcce27d4306776039e02c2904b116ab9a
end:  2018-03-07 14:44:32.488829 +0800 CST m=+60.798522580
elapsed time:  1m0.797798933s


start: 2018-03-07 14:44:32.489057 +0800 CST m=+60.798750578
Mining the block containing "Send 1 BTC to Ivan"

000000e9d5a266faa6a86f56a36ea09212ecad28e524a8b0599589fd5b800d13
end:  2018-03-07 14:46:32.996032 +0800 CST m=+181.307571128
elapsed time:  2m0.508818203s


start: 2018-03-07 14:46:32.996498 +0800 CST m=+181.308036527
Mining the block containing "Send 2 more BTC to Ivan"

0000001927685501c59f28c0bda3fdd0472e88d6eec3822b0ab98e5b1c28c676
end:  2018-03-07 14:46:53.90008 +0800 CST m=+202.211938702
elapsed time:  20.903900066s
```

可以看到生成了前6位都是0的hash字符串，因为是16进制的，就是2<sup>4</sup>一位，共4*6=24位，也就是我们设置的`targetBits`。从时间上可以看到，计算出hashcash的时间有一定的随机性，多着2分，少则20秒。

现在还需要去验证是否是正确的。

```golang
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int
	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	isValid := hashInt.Cmp(pow.target) == -1
	return isValid
}
```
可以看到，`nonce`在验证是要用到的，告诉对方也用我们计算出来的数据进行二次计算，如果对方计算的符合要求，说明我们的计算合法。

把验证方法加到main函数中：
```golang
func main() {
	...

	for _, block := range bc.blocks {
		...
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}
}
```
结果：

```log
Prev. hash:
Data: Genesis Block
Hash: 0000006f3387b588739cbcfe2cce521fcce27d4306776039e02c2904b116ab9a
PoW: true

Prev. hash: 0000006f3387b588739cbcfe2cce521fcce27d4306776039e02c2904b116ab9a
Data: Send 1 BTC to Ivan
Hash: 000000e9d5a266faa6a86f56a36ea09212ecad28e524a8b0599589fd5b800d13
PoW: true

Prev. hash: 000000e9d5a266faa6a86f56a36ea09212ecad28e524a8b0599589fd5b800d13
Data: Send 2 more BTC to Ivan
Hash: 0000001927685501c59f28c0bda3fdd0472e88d6eec3822b0ab98e5b1c28c676
PoW: true
```

## 本章总结

本章我们的区块链进一步接近实际的结构：增加了计算难度，这意味着挖矿成为可能。不过还是欠缺了一些特性，比如没有把计算出来的数据持久化，没有钱包（wallet）、地址、交易，以及实现一致性机制。接下来的几篇abc中，我们会持续完善。

## 相关链接

[本文代码][本文代码]
[一二章代码差异](https://github.com/printfcoder/blockchain-abc/compare/part_1...part_2)

本序列文章：

1. [Golang 区块链入门 第一章 基本概念][本序列第一篇]
2. [Golang 区块链入门 第二章 工作量证明][本序列第二篇]
3. [Golang 区块链入门 第三章 持久化和命令行接口][本序列第三篇]
4. [Golang 区块链入门 第四章 交易 第一节][本序列第四篇]
5. [Golang 区块链入门 第五章 地址][本序列第五篇]
6. [Golang 区块链入门 第六章 交易 第二节][本序列第六篇]

[原文]: https://jeiwan.cc/posts/building-blockchain-in-go-part-2/
[hash算法]: https://en.bitcoin.it/wiki/Block_hashing_algorithm
[Hashcash]: https://printfcoder.github.io/myblog/algorithms/security/cipher/2018/02/09/notes-on-hashcash/
[本文代码]: https://github.com/printfcoder/blockchain-abc/tree/part_2
[本序列第一篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/05/abc-building-blockchain-in-go-part-1-basic-prototype/
[本序列第二篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/06/abc-building-blockchain-in-go-part-2-proof-of-work/
[本序列第三篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/07/abc-building-blockchain-in-go-part-3-persistence-and-cli/
[本序列第四篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/09/abc-building-blockchain-in-go-part-4-transactions-1/
[本序列第五篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/14/abc-building-blockchain-in-go-part-5-address/
[本序列第六篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/17/abc-building-blockchain-in-go-part-6-transactions-2/