[前面的章节][本序列第四篇]我们开始实现交易。也介绍了和人与人交易的不同：没有账户，也不需要个人信息（名字，户照，社保卡等等），也没有保存在比特币中。但是要有机制能确定你就是交易的output拥有者（能解锁这些被锁住的output的人)，address地址就是干这个的。前面我们使用了比较随意的用户定义字符串作为地址，现在我们实现真正的地址，和在比特币中实现的那样。

[原文][原文]（略有删改）

# 比特币地址

这里有一份比特币地址的[例子1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa][比特币地址样例]。这是第一个比特币的地址，传言就是中本聪本人的比特币。比特币地址是公开的。如果要发送比特币给其他人，就得知道他的地址。但是地址（尽管都是唯一的）不是区分你是钱包的拥有者的东西。事实上，这些地址都是人眼可读的公钥文本。在比特币中，你的标识是一组（存放在你电脑或者其它你指定的地方）上的私钥和公钥对。比特币依赖加密算法的组合创建的密钥，用以保证世界上没有任何一个人可以绕开你物理机上的实体私钥就可以操控你的比特币。现在探讨一下这个算法机制。

# 公钥加密

公钥加密算法使用密钥对--公钥和私钥。公钥不敏感，可以公开给任何人。与此相反，私钥则不应该暴露出来，除了私钥持有者其它人都不能访问私钥，因为这是持有者身份标识。可以这么说，在加密了的世界里，你就是你的私钥。

比特币的钱包本质上就是上面那些密钥对。当你安装钱包应用或者比特币客户端生成新的地址时，一对密钥就生成了。谁控制了私钥就控制了所有的发送到这个密钥（公钥地址文本）的所有币。

私钥与公钥是随机的byte序列，因此打印出来也不是人可读的。所以比特币使用了另一个算法来把公钥转成字符串，让人类可读。

> If you’ve ever used a Bitcoin wallet application, it’s likely that a mnemonic pass phrase was generated for you. Such phrases are used instead of private keys and can be used to generate them. This mechanism is implemented in [BIP-039][bip-0039].

> 如果你有使用过比特币钱包应用，那么就好比给你生成了一个助记密文短语。这些短语可以用到代替私钥，也可以生成私钥<sup>[?](https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki#from-mnemonic-to-seed)</sup>。这一机制基于[BIP-039][bip-0039]实现。

现在知道比特币中是什么区分人员的。但是比特币中是怎么检测交易中output（及存在output中的币）所有权的？

# 数字签名

数学和密码学里，有一个概念叫数字加密。该算法保证：

1. 数据在从发送者传输到接收者过程中不会被更改
2. 数据由确定的发送者创建
3. 发送者不能否认发送过这笔数据

通过一个给数据的签名算法，获得签名，过后可以用来验证。数字签名和私钥一起使用，然后用公钥来验证。（公钥好比锁，大家可以都有，私钥就是钥匙，只有这个锁能证明是这个私钥能打开它，同样反过来，只有这个私钥能打开这个锁证明它是数据拥有者）。

要签名数据得有两个东西：

1. 需要签名的数据
2. 签名的私钥

签名操作产生签名，这个签名就存放在交易的input中。为了验证签名，还需要：

1. 刚被签名的数据 
2. 签名
3. 公钥

简单来说，验证过程可以这么描述：检测签名是从这笔数据与私钥一起计算得来的，而这个公钥也是由该私钥生成的。

> Digital signatures are not encryption, you cannot reconstruct the data from a signature. This is similar to hashing: you run data through a hashing algorithm and get a unique representation of the data. The difference between signatures and hashes is key pairs: they make signature verification possible.
But key pairs can also be used to encrypt data: a private key is used to encrypt, and a public key is used to decrypt the data. Bitcoin doesn’t use encryption algorithms though.

> 数字签名不是加密，你不能在签名中重构出数据。这和hash有点像，你通过hash算法计算数据然后返回一个唯一的数据描述。区别签名和hash的不同是密钥对，密钥对使得验证签名成为可能。但是密钥对也可用于加密数据，私钥用于加密，公钥则用于解密数据。不过，比特币没有用加密算法。

在比特币中，每一笔交易的input都是由创建了这笔交易的人签名。交易在被塞到区块前必须通过验证。验证意味着（除去了一些步骤）：

1. 检测input拥有权限使用前一交易中其关联的output
2. 检测交易签名是正确的

如图，签名和验证的过程：

![](/myblog/assets/images/blockchain/abc/signing-scheme.png)

现在我们重新过一下整个交易的生命周期

1. 首先，有包含coinbase交易的创世区块，此时并没有真正的input在coinbase交易里，所以签名在这一步是不需要的。而coinbase交易的output含有一个使用（**RIPEMD16(SHA256(PubKey)**）算法的hash公钥。
2. 当有人发送币时，会创建一笔交易。交易的input会关联前面交易（可能会关联多个交易）中的output。每回input都会存储一个公钥（没有经过hash处理）和一个用整个交易算出的签名。
3. 比特币网络中其它的节点会收到这个交易然后验证它。它们会检测：input里公钥的hash值是否匹配引用的output的hash值（这一步用于确认发送者只发送了归属他的币）；签名是否正确（证明交易是由币的持有者发起的）。
4. 当矿工节点准备去挖新的区块时，它会把交易塞到区块中，然后开始挖矿。
5. 当区块被挖出来时，网络中每一个其它的节点都会收到该区块被挖出来并被加到区块链中的消息
6. 在区块加入区块链中后，交易就完成了，它的output就可以被新的交易引用（消费）。

# 椭圆曲线加密

前面说到公钥和私钥是两个随机的byte数组序列。因为私钥用于区分持币者，所以就需要满足条件：随机算法必须产生真正的随机bytes。不能让生成其他人已经撑有的私钥。

比特币使用椭圆曲线来生成私钥。椭圆曲线是一个复杂的数学概念，这里就不详细解说了，有兴趣可以查看这篇[文章][椭圆曲线]（警告：很多数学公式）。我们只需要记住这个算法可以生成足够大和随机的数字。比特币中用的椭圆能随机挑出一个介于0到2²⁵⁶（近似于10⁷⁷，要知道，宇宙中有10⁷⁸到10⁸²个原子）。这么大的数字意味着几乎任意两次计算都是不可能产生相同的数字的。

另外，比特币使用（我们也将用）ECDS（Elliptic Curve Digital Signature Algorithm）算法来签名交易。

# Base58

现在我们把注意力回到比特币地址上来：

前面说的1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa这个地址。这是人类可读的公钥表现形式，如果我们把它给解码了，公钥看上去就像这样（转成了16进制系统中的byte串）：

```log
0062E907B15CBF27D5425399EBF6F0FB50EBB88F18C29B7D93
```

比特币使用基于Base58的算法来把公钥转成人眼可读的格式，这个算法与Base64很像，但用于更短的字母表，有些字母被移除以避免某些利用字符相似的攻击。因此，这些符号是没有的：0（数字0）、O（大写字母o）、I（大写字母i）、l（小写的L），因为他们实在太像了。当然，也没有+-（加减）符号。

下图展示从公钥算出地址的过程：

![](/myblog/assets/images/blockchain/abc/address-generation-scheme.png)

所以看到上述的公钥解码后由3个部分组成：

```log
Version  Public key hash                           Checksum
00       62E907B15CBF27D5425399EBF6F0FB50EBB88F18  C29B7D93
```

因为使用哈希函数是单向的（即不能被逆转），也不能从hash值里找出公钥。通过执行同样的hash函数然后再比较这个解码后的hash值是否一致来验证该公钥用于生成该hash，如果一致，则公钥用于计算该hash。

重要的都说完了，开始撸代码吧。有些概念在写代码前要弄清楚。

# 实现地址

先定义**钱包（wallet）**的结构：

```golang
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

type Wallets struct {
	Wallets map[string]*Wallet
}

func NewWallet() *Wallet {
	private, public := newKeyPair()
	wallet := Wallet{private, public}

	return &wallet
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pubKey
}
```

钱包除了密钥对，其它什么也没有。现在需要**Wallets**类型来维护钱包集合，把它们的数据落地。**Wallet**的构造函数，有新的密钥对生成。**newKeyPair**函数比较简单，“ECDSA”算法基于椭圆曲线，是我们需要的。下一步，使用椭圆算法生成私钥，然后用私钥生成公钥。注意一点，在椭圆曲线算法中，公钥是在椭圆上的点集合。因此，公钥是直角坐标系的坐标集合，比特币中，这些坐标串起来构成公钥。

现在生成地址：

```golang
func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey)

	versionedPayload := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versionedPayload)

	fullPayload := append(versionedPayload, checksum...)
	address := Base58Encode(fullPayload)

	return address
}

func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}

func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}

```

下面是把公钥转成Base58规范的地址步骤：

1. 获取公钥，使用**RIPEMD160(SHA256(PubKey))**执行两次hash算法。
2. 给hash加上地址生成算法版本
3. 使用**SHA256(SHA256(payload))**hash计算第2步的结果，得到的hash值前4bytes就是校验码。
4. 把校验码附加到**version+PubKeyHash**组合。
5. 使用Base58编码**version+PubKeyHash+checksum**组合

结果，你就算出了一个**真正的比特币地址**，你甚至可以在[blockchain.info][https://blockchain.info/]上找到它的余额。但是刚保证余额肯定是0，不论你生成多少次新的地址，再检测也是0。这也就是为什么选择合适的公钥加密算法非常重要：考虑到私钥是随机的数字 ，生成相同的数字几乎是不可能的，事实上，这个可能性会低到“永不发生”。

还有，要注意你不需要连接到任何的比特币节点去获取它的地址。地址生成算法使用的是开源的算法组合，这些算法在很多编程语言和库中都有实现。

现在修改Input和Output结构，让其能使用地址。

```golang
type TXInput struct {
	Txid      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}

func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

type TXOutput struct {
	Value      int
	PubKeyHash []byte
}

func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

```

注意，我们没有再使用**ScriptPubKey**和**ScriptSig**，因为我们并不准备去实现一个脚本语言。相反，**ScriptSig**分成了**Signature**和**PubKey**。**ScriptPubKey**也改名成**PubKeyHash**。我们会实现与比特币中相同的output锁/解锁和input签名逻辑，但是我们通过使用方法（method）来实现。

**UsesKey**方法负责检测input使用了特别的密钥来解锁output。注意input里存放的是原生的未进行hash处理过的公钥，但是这个函数得到的是经过hash处理过的公钥。

**IsLockedWithKey**负责检测提供的公钥hash是否能用于给output加锁。它是**UsesKey**函数的补充，它们都用在**FindUnspentTransactions**中，用于创建交易之间的连接。

**Lock**就是简单地把output锁上。当把币发送经其它人时，我们是知道他们的地址的，因此这个函数会要求传入这个地址，然后会被解码，把公钥的哈希抽出来再保存到**PubKeyHash**字段中。

现在看看是能工作：

```shell
$ blockchain_go createwallet
Your new address: 13Uu7B1vDP4ViXqHFsWtbraM3EfQ3UkWXt

$ blockchain_go createwallet
Your new address: 15pUhCbtrGh3JUx5iHnXjfpyHyTgawvG5h

$ blockchain_go createwallet
Your new address: 1Lhqun1E9zZZhodiTqxfPQBcwr1CVDV2sy

$ blockchain_go createblockchain -address 13Uu7B1vDP4ViXqHFsWtbraM3EfQ3UkWXt
0000005420fbfdafa00c093f56e033903ba43599fa7cd9df40458e373eee724d

Done!

$ blockchain_go getbalance -address 13Uu7B1vDP4ViXqHFsWtbraM3EfQ3UkWXt
Balance of '13Uu7B1vDP4ViXqHFsWtbraM3EfQ3UkWXt': 10

$ blockchain_go send -from 15pUhCbtrGh3JUx5iHnXjfpyHyTgawvG5h -to 13Uu7B1vDP4ViXqHFsWtbraM3EfQ3UkWXt -amount 5
2017/09/12 13:08:56 ERROR: Not enough funds

$ blockchain_go send -from 13Uu7B1vDP4ViXqHFsWtbraM3EfQ3UkWXt -to 15pUhCbtrGh3JUx5iHnXjfpyHyTgawvG5h -amount 6
00000019afa909094193f64ca06e9039849709f5948fbac56cae7b1b8f0ff162

Success!

$ blockchain_go getbalance -address 13Uu7B1vDP4ViXqHFsWtbraM3EfQ3UkWXt
Balance of '13Uu7B1vDP4ViXqHFsWtbraM3EfQ3UkWXt': 4

$ blockchain_go getbalance -address 15pUhCbtrGh3JUx5iHnXjfpyHyTgawvG5h
Balance of '15pUhCbtrGh3JUx5iHnXjfpyHyTgawvG5h': 6

$ blockchain_go getbalance -address 1Lhqun1E9zZZhodiTqxfPQBcwr1CVDV2sy
Balance of '1Lhqun1E9zZZhodiTqxfPQBcwr1CVDV2sy': 0
```

不错，现在实现交易签名。

# 签名

交易必须被签名，这是比特币中，唯一能保证没有人可以消费别人的币的机制。如果签名不合法，交易也是不合法的。因此，交易也不会被加到区块链中。

除了数据签名，交易中有关签名的所有点都实现了。交易中有哪几部分是真正要签名的？或者整个交易都要被签名？选择需要加密的数据是非常重要的。问题是被签名的数据有独特的方式区别数据信息。举个栗子，仅签名output的值一点也没用，因为这样签名并没有考虑发送者和接收者。

考虑到交易解锁前面交易的output,重新分配它们的值，然后锁到新的output中，下列的数据必须是加密的：

1. 保存在解锁了的output公钥的hash值。这可以辨别交易的发送者。
2. 保存在新的、加锁了的output公钥的hash值。这可以辨别交易的接收者。
3. 新output的值。

> In Bitcoin, locking/unlocking logic is stored in scripts, which are stored in ScriptSig and ScriptPubKey fields of inputs and outputs, respectively. Since Bitcoins allows different types of such scripts, it signs the whole content of ScriptPubKey.

> 在比特币中，加锁/解锁逻辑是存储在脚本中的，分别存储在input和output的ScriptSig、ScriptPubKey字段中。因为比特币允许不同的类型脚本，所以会对整个ScriptPubKey的内容进行签名。

可以看到，我们并不需要去签名input中的公钥，因此，比特币里并不是对整个交易签名的，但是对input从output引用的**ScriptPubKey**进行了适度修剪。

> A detailed process of getting a trimmed transaction copy is described here. It’s likely to be outdated, but I didn’t manage to find a more reliable source of information.

> 更详细的获取裁剪过的交易备份[描述](https://en.bitcoin.it/wiki/File:Bitcoin_OpCheckSig_InDetail.png)，可能比较老了，但是我找不到更可靠的资源了

看上去比较复杂，先从**Sign**开始编写：

```golang
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	txCopy := tx.TrimmedCopy()

	for inID, vin := range txCopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vin[inID].Signature = signature
	}
}
```
方法接收私钥和前面交易的map。前面提到，为了签名交易，需要访问交易中的input引用的ouput，因此需要存放了这些output的交易。

一步一步分析这段代码：

```golang
if tx.IsCoinbase() {
	return
}
```
coinbase 交易是不需要签名的，因为没有真正的input。

```golang
txCopy := tx.TrimmedCopy()
```

签名修剪后的交易副本，而不是整个交易：

```golang
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
}
```
该拷贝含有所有的input和output，但是**TXInput.Signature**和**TXInput.PubKey**设置成了空值

下一步是遍历拷贝中的每一个input：

```golang
for inID, vin := range txCopy.Vin {
	prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
	txCopy.Vin[inID].Signature = nil
	txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
```
遍历中，**Signature**会设置成空（只是复核，确定为空），**PubKey**设成引用的output的**PubKeyHash**值。此时，所有的交易除了当前的都是“空的”，也即是**Signature**和**PubKey**被设置成了空值。因此，input是被**分开（separately）**签名的，尽管在我们的应用里是不必要的，但是比特币允许交易中的input引用不同的地址。

```golang
	txCopy.ID = txCopy.Hash()
	txCopy.Vin[inID].PubKey = nil
```

**Hash**方法把交易序列化再用SHA-256算法算出交易的hash值。得到的结果就是用来签名的数据。算出这个hash值后，还需要把**PubKey**字段重置，所以它不会影响后面的迭代。

核心代码：

```golang
    r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
	signature := append(r.Bytes(), s.Bytes()...)

	tx.Vin[inID].Signature = signature
```

我们用**privKey**给**txCopy.ID**签名。ECDSA算法签名就是一对数字，我们把它们连起来并存放到input的**Signature**里。

验证函数：

```golang
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inID, vin := range tx.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil

		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
			return false
		}
	}

	return true
}
```
这个方法比较简单，首先我们拷贝一份交易的副本：

```golang
 	txCopy := tx.TrimmedCopy()
```

然后我们需要相同的椭圆来生成密钥对：

```golang
    curve := elliptic.P256()
```

给每一个input签名：

```golang
for inID, vin := range tx.Vin {
	prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
	txCopy.Vin[inID].Signature = nil
	txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
	txCopy.ID = txCopy.Hash()
	txCopy.Vin[inID].PubKey = nil
```
这块代码与**Sign**中的方法一样，因为在验证过程中，我们需要的数据，得与被签名的相同。

```golang
	r := big.Int{}
	s := big.Int{}
	sigLen := len(vin.Signature)
	r.SetBytes(vin.Signature[:(sigLen / 2)])
	s.SetBytes(vin.Signature[(sigLen / 2):])

	x := big.Int{}
	y := big.Int{}
	keyLen := len(vin.PubKey)
	x.SetBytes(vin.PubKey[:(keyLen / 2)])
	y.SetBytes(vin.PubKey[(keyLen / 2):])
```
这一步我们把**TXInput.Signature**和**TXInput.PubKey**中的值抽出来，因为签名是一个数字对，公钥是X,Y坐标。在此前为了存储而把它们给组合起来，现在需要拆开得到值来计算**crypto/ecdsa**计算。

```golang
rawPubKey := ecdsa.PublicKey{curve, &x, &y}
	if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
		return false
	}
}
return true
```

我们使用从input中抽出来的公钥创建了一个**ecdsa.PublicKey**公钥，把input中抽出来的签名用**ecdsa.Verify**验证。如果所有input都验证通过了，就返回true。如果有一个失败，都返回false。

现在我们需要一个函数可以获取此前的交易。因为这一操作需要与整个区块链交互，所以得在**Blockchain**区块链上加一个方法：
```golang
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found")
}

func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}
```
这些函数都比较简单：**FindTransaction**用于通过ID找到交易（这需要遍历整个区块链中的区块）；**SignTransaction**则负责给传进来的交易找到其引用的其它交易，并给它签名。**VerifyTransaction**和**SignTransaction**差不多，只是它不是负责签名，而是验证签名。

现在需要签名与验证签名。签名过程在**NewUTXOTransaction**函数中执行。

```golang
func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	...

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	bc.SignTransaction(&tx, wallet.PrivateKey)

	return &tx
}
```

在交易被塞到区块之前，需要验证它：

```golang
func (bc *Blockchain) MineBlock(transactions []*Transaction) {
	var lastHash []byte

	for _, tx := range transactions {
		if bc.VerifyTransaction(tx) != true {
			log.Panic("ERROR: Invalid transaction")
		}
	}
	...
}
```

OK，再运行一下程序看是否正常：

```shell
$ blockchain_go createwallet
Your new address: 1AmVdDvvQ977oVCpUqz7zAPUEiXKrX5avR

$ blockchain_go createwallet
Your new address: 1NE86r4Esjf53EL7fR86CsfTZpNN42Sfab

$ blockchain_go createblockchain -address 1AmVdDvvQ977oVCpUqz7zAPUEiXKrX5avR
000000122348da06c19e5c513710340f4c307d884385da948a205655c6a9d008

Done!

$ blockchain_go send -from 1AmVdDvvQ977oVCpUqz7zAPUEiXKrX5avR -to 1NE86r4Esjf53EL7fR86CsfTZpNN42Sfab -amount 6
0000000f3dbb0ab6d56c4e4b9f7479afe8d5a5dad4d2a8823345a1a16cf3347b

Success!

$ blockchain_go getbalance -address 1AmVdDvvQ977oVCpUqz7zAPUEiXKrX5avR
Balance of '1AmVdDvvQ977oVCpUqz7zAPUEiXKrX5avR': 4

$ blockchain_go getbalance -address 1NE86r4Esjf53EL7fR86CsfTZpNN42Sfab
Balance of '1NE86r4Esjf53EL7fR86CsfTZpNN42Sfab': 6
```

本篇也快搞完了

把**NewUTXOTransaction**中调用的**bc.SignTransaction(&tx, wallet.PrivateKey)**给注释掉，确定未签名的交易不能被挖出来。

```golang
func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
   ...
	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	// bc.SignTransaction(&tx, wallet.PrivateKey)

	return &tx
}
```

```shell
$ go install
$ blockchain_go send -from 1AmVdDvvQ977oVCpUqz7zAPUEiXKrX5avR -to 1NE86r4Esjf53EL7fR86CsfTZpNN42Sfab -amount 1
2017/09/12 16:28:15 ERROR: Invalid transaction
```

# 总结

我们从前面几章开始讲了这么久来实现比特币中的各种关键特性。我们实现了大多数，除了网络连通，下一章，我们把交易弄完。

# 相关链接

[本文代码][本文代码]


本序列文章：

1. [Golang 区块链入门 第一章 基本概念][本序列第一篇]
2. [Golang 区块链入门 第二章 工作量证明][本序列第二篇]
3. [Golang 区块链入门 第三章 持久化和命令行接口][本序列第三篇]
4. [Golang 区块链入门 第四章 交易 第一节][本序列第四篇]
5. [Golang 区块链入门 第五章 地址][本序列第五篇]
6. [Golang 区块链入门 第六章 交易 第二节][本序列第六篇]

[本序列第一篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/05/abc-building-blockchain-in-go-part-1-basic-prototype/
[本序列第二篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/06/abc-building-blockchain-in-go-part-2-proof-of-work/
[本序列第三篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/07/abc-building-blockchain-in-go-part-3-persistence-and-cli/
[本序列第四篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/09/abc-building-blockchain-in-go-part-4-transactions-1/
[本序列第五篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/14/abc-building-blockchain-in-go-part-5-address/
[本序列第六篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/17/abc-building-blockchain-in-go-part-6-transactions-2/

[原文]: https://jeiwan.cc/posts/building-blockchain-in-go-part-6/
[本文代码]: https://github.com/printfcoder/blockchain-abc/tree/part_5
[比特币地址样例]: https://blockchain.info/address/1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa
[bip-0039]: https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki
[椭圆曲线]: http://andrea.corbellini.name/2015/05/17/elliptic-curve-cryptography-a-gentle-introduction/
