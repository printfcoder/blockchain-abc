# Golang 区块链入门 第7章 网络

到目前为止，前面我们构造了拥有全部关键功能的区块链，匿名、安全、随机的地址；区块链数据存储；工作量证明系统；可靠的交易存储。这些特性都很重要，但还不够。能够让这些能力发出耀眼光芒的，让加密货币成为可能，是网络。这样实现的区块链，只能运行在一台计算机上的能有什么卵用？那些基于密码学的特性有什么用，什么时候又只会有一个人使用？是网络让所有的机制运行起来且变得有用。

可以把区块链中的我当成规则，类似于人们在彼此的生活成长中而建立的规则。一种社会秩序。区块链网络是一个遵从相同的规则的程序生态社区，也是因为遵从这些规则而赋予区块链网络生命。类似地，当人们分享相同的想法时，就会变得更强也会一起创建更好的生活。有的人遵从不同的规则时，这些人就会被社会隔离（国家、公社，等等）。同样，如果区块链节点都使用遵守不同的规则，那么它们只会在一个隔离的网络中生长。

**这点非常重要**：不骨网络没有大量的节点共享相同的规则，这些规则一点用也没有。

> 免责声明：非常不幸，我没有足够的时间来实现真正的P2P网络。这篇文章我会阐明最常用的场景，涉及不同类型的节点。改善这一场景并使之成为P2P网络对你来说是一个非常不错的挑战和尝试。并且，我不保证其它非本章的实现方式可以运行。抱歉！
> 
> 本篇文章的代码改动比较大，就不详细解释了。可以到[这里](https://github.com/printfcoder/blockchain-abc//compare/part_6...part_7)来看所有的改变。看他们的区别。

## 区块链网络
区块链网络是去中心化的，也就是说是没有一个中央服务器作为伺服，也没有客户端向服务器获取或传送数据。在区块链网络中有节点，每一个节点都是该网络中完整的成员。一个节点就是一切，即是服务器也是客户端。记住这点很重要，和Web应用是不同的。

区块链网络是P2P（Peer-to-Peer）网络，也就是说节点之间是彼此直接相连的。这个拓扑非常庞大，因为节点角色之间没有层级的之分。下图是P2P网络的图解：
![](https://printfcoder.github.io/myblog/assets/images/blockchain/abc/p2p-network.png)

[使用Freepik画的](https://www.freepik.com/dooder)

节点在这样的网络是非常难实现的，因为他们必须执行很多操作。每个节点都肯定会和其它很多节点交互，也会请求其它节点的状态，与自己的状态比较，如果自己的状态过期时就要更新状态。

## 节点规则
尽管是成熟的，区块链节点在网络中可以充当不同的角色：

### 1. 矿工
有些节点运行强大或特制的硬件（比如ASIC），它们的目的就是尽可能快地挖出新的区块。矿机可能是仅有的在区块链里使用工作量证明的（程序），因为挖矿意味着要解释PoW问题。而举个例子，在权益证明（Proof-of-Stake）区块链中是没有挖矿的。

### 2. 全功能节点
这些节点验证矿工挖出来的区块和核实交易。为了完成这点，它们必须握有整个区块链的副本。并且，这些节点执行路由操作，就像帮助其它节点互相发现。

网络中拥有全功能节点非常重要，这些节点可以执行决策：它们能裁定是否区块或者交易是合法的。

### 3. SPV

SPV代表Simplified Payment Verification，简化交易验证。这些节点并不会去保存区块链的完整副本，但是却能够核实交易（并不是全部，而是子集，比如，那些会发送到特殊地址的交易）。SPV节点依赖全功能节点提供数据，也可以有多个SPV节点连接到一个全功能节点。SPV使得钱包应用成为可能：钱包不需要下载整个区块链，但是能够核实交易。

## 网络简化

为了在我们的区块链中实现网络，我们必须得简化一点东西。问题在于我们没有太多的电脑来模拟有多个节点的网络。我们过去是可以使用虚拟机或者Docker来解决这个问题的，但是这会让每件事都变得复杂，我们得解决虚拟机或者Docker的问题，而我们的目标仅是集中精力到区块链的实现上。所以，我们需要运行多个区块链节点在一台机器上，以此同时，我们还要让它们有不同的地址。为了实现这点，我们使用**端口**作为节点的标识，而不是IP地址，等等。下面将会有节点拥有这些地址，**127.0.0.1:3000**，**127.0.0.1:3001**，**127.0.0.1:3002**等等。我们会调用端口的节点id，然后用**NODE_ID**环境变量设置它们。因此，你可以打开多个终端窗口，设置不同的**NODE_ID**，就可以运行不同的节点了。

这波操作同样需要不同的区块链和钱包文件。它们现在必须依赖节点id，并被命名像：**blockchain_3000.db**, **blockchain_30001.db**和**wallet_3000.db**, **wallet_30001.db**等等

## 实现
那么，当下载时发生了什么，是指下载Bitcoin Core然后第一次运行？答案是必须连接到一些节点下载区块链最后的状态。考虑到你们计算机并不知道全部或者部分的比特币节点，到底这个节点是什么呢。

在Bitcoin Core里使用硬编码地址可能会出错，节点会被攻击或者关掉，会导致新的节点不能加入到网络中。相反，在Bitcoin Core中，有使用[DNS seeds](https://bitcoin.org/en/glossary/dns-seed)硬编码。它们不是节点，而是存放了一些节点地址的DNS服务器。当你开始运行一个纯净的Bitcoin Core时，它会连接到一个seed然后获取上面记录的所有节点列表，根据这个列表下载区块链。

不过，在我们的实现中，还是会中心化。会用到三个节点：

1. 中心节点：这个节点会被其它节点连接。该节点会在其它节点之间发送数据。
2. 矿工节点：这个节点会存储新的交易到缓存池中，当有足够的交易时，它就会挖出新的区块。
3. 钱包节点：这个节点会用来在钱包之间发送钱币。但是和SPV节点不同，它会存储区块链的完整副本。

### 场景
本篇的目标是实现下面的场景：
1. 中心节点生成新的区块链
2. 钱包节点连接到中心节点然后下载区块链
3. 矿工节点连接到中心节点然后下载区块链
4. 钱包节点创建交易
5. 矿工节点接收交易并把它缓存在缓存池中
6. 当缓存池中有足够的交易时，矿工开始挖新的区块
7. 当新的区块被挖出来时，会被发送到中心节点。
8. 钱包节点与中心节点同步
9. 钱包使用者检测他们支付是否成功

这个场景看起来和比特币很像。尽管我们没有构建一个真正的P2P网络，我们准备实现一个真实的，比特币的主要、最重要的使用案例。

[原文][原文]（略有删改）

> 本节的阐述会有重大的代码改变，如果在这里讲就有点麻烦了。请跳到[这里](https://github.com/printfcoder/blockchain-abc//compare/part_5...part_6#files_bucket)来看所有的改变。

### 版本
节点通过消息的含义进行沟通。当新的节点运行时，它会从DNS种子获取节点的信息，然后向它们发送**版本**信息，在我们的实现中，版本的结构如下：
```golang
type version struct {
    Version    int
    BestHeight int
    AddrFrom   string
}
```
我们只有一个区块链版本号，所有**Version**字段不能含有任何重要的信息。**BestHeight**存放节点的区块链长度。**AddFrom**保存发送者的地址。

节点接收**版本**消息做什么呢？它会回复它自己的**版本**信息。这是握手的一种类型，除了先去彼此打招呼别无其它的交互可能。但是这并不仅仅是有礼貌，**版本**用于找到更长的区块链。当一个节点接收到**版本**信息时，它会检测是否节点的区块链比**BestHeight**值要大。如果不是，节点就会请求下载缺失的区块。

为了能接收到消息，我们要有一个服务器：

```golang
var nodeAddress string
var knownNodes = []string{"localhost:3000"}

func StartServer(nodeID, minerAddress string) {
    nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
    miningAddress = minerAddress
    ln, err := net.Listen(protocol, nodeAddress)
    defer ln.Close()

    bc := NewBlockchain(nodeID)

    if nodeAddress != knownNodes[0] {
        sendVersion(knownNodes[0], bc)
    }

    for {
        conn, err := ln.Accept()
        go handleConnection(conn, bc)
    }
}
```
首先，在中央服务器的地址上使用硬编码，因为每一个新的节点都必须要知道从哪里获得初始化数据。**minerAddress**参数指定接收挖出新区块的奖励地址。

```golang
if nodeAddress != knownNodes[0] {
    sendVersion(knownNodes[0], bc)
}
```
就是说当前节点不是中央节点时，它就会发送**version**消息到中央节点判断是否自已的区块链是否过期了。
```golang
func sendVersion(addr string, bc *Blockchain) {
    bestHeight := bc.GetBestHeight()
    payload := gobEncode(version{nodeVersion, bestHeight, nodeAddress})

    request := append(commandToBytes("version"), payload...)

    sendData(addr, request)
}
```
消息是底层的比特序列。前12字节指定了命令名（在这里的情况就是“version”），后面的字节会包含**gob**编码过的消息结构。**commandToBytes**：

```golang
func commandToBytes(command string) []byte {
    var bytes [commandLength]byte

    for i, c := range command {
        bytes[i] = byte(c)
    }

    return bytes[:]
}
```
它创建了12字节的缓存区，使用命令名来填充，把余留的字节置空。上面是它的反向函数：
```golang
func bytesToCommand(bytes []byte) string {
    var command []byte

    for _, b := range bytes {
        if b != 0x0 {
            command = append(command, b)
        }
    }

    return fmt.Sprintf("%s", command)
}
```

当节点接收到命令时，它会运行**bytesToCommand**指令把命令名展开，然后使用正确的处理函数执行命令：
```golang
func handleConnection(conn net.Conn, bc *Blockchain) {
    request, err := ioutil.ReadAll(conn)
    command := bytesToCommand(request[:commandLength])
    fmt.Printf("Received %s command\n", command)

    switch command {
    ...
    case "version":
        handleVersion(request, bc)
    default:
        fmt.Println("Unknown command!")
    }

    conn.Close()
}
```
**version**处理函数如下：
```golang
func handleVersion(request []byte, bc *Blockchain) {
    var buff bytes.Buffer
    var payload verzion

    buff.Write(request[commandLength:])
    dec := gob.NewDecoder(&buff)
    err := dec.Decode(&payload)

    myBestHeight := bc.GetBestHeight()
    foreignerBestHeight := payload.BestHeight

    if myBestHeight < foreignerBestHeight {
        sendGetBlocks(payload.AddrFrom)
    } else if myBestHeight > foreignerBestHeight {
        sendVersion(payload.AddrFrom, bc)
    }

    if !nodeIsKnown(payload.AddrFrom) {
        knownNodes = append(knownNodes, payload.AddrFrom)
    }
}
```
首先要解码请求，展开内部信息。所有的处理函数都是相似的，后面会把篇幅省下来。

然后节点会用它的**BestHeight**与消息中的比较。如果节点区块更长时，那么它会回复**version**消息，相反，它会发送**getBlocks（获取区块）**消息。

### getblocks

```golang
type getblocks struct {
    AddrFrom string
}
```

**getblocks**的意思是**亮出你有的区块**（在比特币中，会更复杂）。注意，不是**扔你所有的区块过来**，相反它是请求区块hash的列表。这么做是为了降低网络负载，因为区块可以从不同的节点下载，我们也不用到一个节点去下载上千兆的数据。

处理这个命令比较简单：
```golang
func handleGetBlocks(request []byte, bc *Blockchain) {
    ...
    blocks := bc.GetBlockHashes()
    sendInv(payload.AddrFrom, "block", blocks)
}
```
我们的实现中，它会返回所有区块的hash

### inv

```golang
type inv struct {
    AddrFrom string
    Type     string
    Items    [][]byte
}
```
比特币中使用**inv**来向其它节点展示当前节点有哪些区块或者交易。再说一遍，它并不包含所有的区块和交易，只保存有它们的hash值。**Type**字段用来声明这里存的是区块还是交易。

处理**inv**就比较复杂些了：
```golang
func handleInv(request []byte, bc *Blockchain) {
    ...
    fmt.Printf("Recevied inventory with %d %s\n", len(payload.Items), payload.Type)

    if payload.Type == "block" {
        blocksInTransit = payload.Items

        blockHash := payload.Items[0]
        sendGetData(payload.AddrFrom, "block", blockHash)

        newInTransit := [][]byte{}
        for _, b := range blocksInTransit {
            if bytes.Compare(b, blockHash) != 0 {
                newInTransit = append(newInTransit, b)
            }
        }
        blocksInTransit = newInTransit
    }

    if payload.Type == "tx" {
        txID := payload.Items[0]

        if mempool[hex.EncodeToString(txID)].ID == nil {
            sendGetData(payload.AddrFrom, "tx", txID)
        }
    }
}
```
当区块的hash转移好后，需要把它们保存到**blocksInTransit**变量中来跟踪下载过的区块。这允许我们可以从不同的节点下载区块。在区块进入传输状态后，发送**getData**指令给**inv**的发送者然后更新**blocksInTransit**。在真正的P2P网络中，得在不同的区块之间传办理区块。

在实现中，还永不会发送**inv**时带上多个hash。这也是为什么当**payload.Type == "tx"**时，只用到数组中获取第一个hash。然后检测是否刚刚的txID是否存在，如果不存在，那么发送**getdata**指令获取这个交易。

### getdata
```golang
type getdata struct {
    AddrFrom string
    Type     string
    ID       []byte
}
```

**getdata**用于请求一个指定的区块或交易，它只能带有一个区块或交易的id。

```golang
func handleGetData(request []byte, bc *Blockchain) {
    ...
    if payload.Type == "block" {
        block, err := bc.GetBlock([]byte(payload.ID))

        sendBlock(payload.AddrFrom, &block)
    }

    if payload.Type == "tx" {
        txID := hex.EncodeToString(payload.ID)
        tx := mempool[txID]

        sendTx(payload.AddrFrom, &tx)
    }
}
```

这个**getdata**处理函数比较简单。当请求的是区块，则返回区块；如果是交易，则返回交易。注意，这里有个缺陷，就是没有去检测是否存在指定的区块或者交易。

### 区块和交易
```golang
type block struct {
    AddrFrom string
    Block    []byte
}

type tx struct {
    AddFrom     string
    Transaction []byte
}
```
就是这些消息完成真正的数据传送

**block**的处理器很简单：
```golang
func handleBlock(request []byte, bc *Blockchain) {
    ...

    blockData := payload.Block
    block := DeserializeBlock(blockData)

    fmt.Println("Recevied a new block!")
    bc.AddBlock(block)

    fmt.Printf("Added block %x\n", block.Hash)

    if len(blocksInTransit) > 0 {
        blockHash := blocksInTransit[0]
        sendGetData(payload.AddrFrom, "block", blockHash)

        blocksInTransit = blocksInTransit[1:]
    } else {
        UTXOSet := UTXOSet{bc}
        UTXOSet.Reindex()
    }
}
```
当我们接收到新的区块时，我们把它放到我们的区块链中。如果有很多区块需要下载，我们从前一个相同的下载过区块的节点下载它们。当完成全部的区块下载时，UTXO就需要更新了。

> 备注：并不是要无条件相信，我们应该在把每一个传来的区块加入区块链之前得验证它们。
>
> 备注：并不需要运行**UTXOSet.Reindex()**方法，应该用**UTXOSet.Update(block)**，因为区块链太大了，重置索引会花费太多时间。

处理**tx**消息的函数稍微复杂些：

```golang
func handleTx(request []byte, bc *Blockchain) {
    ...
    txData := payload.Transaction
    tx := DeserializeTransaction(txData)
    mempool[hex.EncodeToString(tx.ID)] = tx

    if nodeAddress == knownNodes[0] {
        for _, node := range knownNodes {
            if node != nodeAddress && node != payload.AddFrom {
                sendInv(node, "tx", [][]byte{tx.ID})
            }
        }
    } else {
        if len(mempool) >= 2 && len(miningAddress) > 0 {
        MineTransactions:
            var txs []*Transaction

            for id := range mempool {
                tx := mempool[id]
                if bc.VerifyTransaction(&tx) {
                    txs = append(txs, &tx)
                }
            }

            if len(txs) == 0 {
                fmt.Println("All transactions are invalid! Waiting for new ones...")
                return
            }

            cbTx := NewCoinbaseTX(miningAddress, "")
            txs = append(txs, cbTx)

            newBlock := bc.MineBlock(txs)
            UTXOSet := UTXOSet{bc}
            UTXOSet.Reindex()

            fmt.Println("New block is mined!")

            for _, tx := range txs {
                txID := hex.EncodeToString(tx.ID)
                delete(mempool, txID)
            }

            for _, node := range knownNodes {
                if node != nodeAddress {
                    sendInv(node, "block", [][]byte{newBlock.Hash})
                }
            }

            if len(mempool) > 0 {
                goto MineTransactions
            }
        }
    }
}
```

第一件要做的事就是把新的交易放到缓存池中（再强调一次，交易在被放到缓存池前一定要核实），下一块代码：
```golang
if nodeAddress == knownNodes[0] {
    for _, node := range knownNodes {
        if node != nodeAddress && node != payload.AddFrom {
            sendInv(node, "tx", [][]byte{tx.ID})
        }
    }
}
```
检测是否当前的节点是中央节点，在我们的实现当中，中央节点并不会挖矿，相反，它只是把新的交易传送给网络中的其它节点。

接下来这块代码只是给矿机节点用的，把它分成两小片：
```golang
if len(mempool) >= 2 && len(miningAddress) > 0 {
```
**miningAddress**只有矿机节点才会被设置。当前的节点中有2个或多个交易在缓存池中时，挖矿就开始。
```golang
for id := range mempool {
    tx := mempool[id]
    if bc.VerifyTransaction(&tx) {
        txs = append(txs, &tx)
    }
}

if len(txs) == 0 {
    fmt.Println("All transactions are invalid! Waiting for new ones...")
    return
}
```

首先，缓存池中所有的交易都是核实过的。不合法的交易会被忽略掉，如果没有合法的交易，挖坑就会中断。

```golang
cbTx := NewCoinbaseTX(miningAddress, "")
txs = append(txs, cbTx)

newBlock := bc.MineBlock(txs)
UTXOSet := UTXOSet{bc}
UTXOSet.Reindex()

fmt.Println("New block is mined!")
```
核实过的交易正被放到区块中，还有带有奖励的coinbase交易。当挖出区块后，UTXO集合就会被重置索引。

>备忘：再一次说明，要用UTXOSet.Update而不是UTXOSet.Reindex

```golang
for _, tx := range txs {
    txID := hex.EncodeToString(tx.ID)
    delete(mempool, txID)
}

for _, node := range knownNodes {
    if node != nodeAddress {
        sendInv(node, "block", [][]byte{newBlock.Hash})
    }
}

if len(mempool) > 0 {
    goto MineTransactions
}
```

在交易被挖时，它就会从缓存池中移除。其它被当前节点通知到的节点都会收到带有新区块hash的**inv**消息。在收到消息后，它们可以请求该刚被挖出的新区块。

## 成果

现在演示上面定义的场景。

首先，在第一个终端窗口中设置环境变量**NODE_ID**为3000（**export NODE_ID=3000**）。我们在下一段中会使用像**NODE 3000**或者**NODE 3001**这样的标识，以便在大家能知道打印出的活动是哪个节点的。

下面的分段title出是切到指定的窗口或打开新窗口

### 节点 3000

创建新的钱包和新的区块链
```shell
$ blockchain_go createblockchain -address CENTREAL_NODE
```
（这里使用假的地址，这样可以简单明了些）

然后，这个区块链只包含有创世区块。我们需要去保存这个区块然后在其它节点中使用它。创世区块作为区块链的标识（在比特币中，创世区块是硬编码的）。

```shell
$ cp blockchain_3000.db blockchain_genesis.db 
```

### 节点 3001

下一步，打开新的终端窗口，把node ID设置为3001。这个节点是钱包节点。用**blockchain_go createwallet**来生成几个地址，定义这些地址为**WALLET_1**、**WALLET_2**、**WALLET_3**。

### 节点 3000

发送一些币到钱包地址中

```shell
$ blockchain_go send -from CENTREAL_NODE -to WALLET_1 -amount 10 -mine
$ blockchain_go send -from CENTREAL_NODE -to WALLET_2 -amount 10 -mine
```

**-mine**指令是说区块会在相同的节点中立马挖出来。我们加个标记是因为在初始时网络中没有矿机节点。

运行这个节点：

```shell
$ blockchain_go startnode
```
这个节点会一直运行直到场景结束。

### 节点 3001

开始这个节点的区块链，带着上面说到的创世区块。

```shell
$ cp blockchain_genesis.db blockchain_3001.db
```

运行节点：
```shell
$ blockchain_go startnode
```
它会去中央节点里下载所有的区块。检测所有事情都好了之后，停止节点然后检测余额。

```shell
$ blockchain_go getbalance -address WALLET_1
Balance of 'WALLET_1': 10

$ blockchain_go getbalance -address WALLET_2
Balance of 'WALLET_2': 10
```
当然，也可以检测**CENTRAL_NODE中央节点**的余额，因为3001节点已经有它自己的区块链了：

```shell
$ blockchain_go getbalance -address CENTRAL_NODE
Balance of 'CENTRAL_NODE': 10
```

### 节点 3002

打开新的窗口，设置ID为3002，然后生成钱包，这个是个矿机节点。初始化区块链：
```shell
$ cp blockchain_genesis.db blockchain_3002.db
```

启动节点
```shell
$ blockchain_go startnode -miner MINER_WALLET
```

### 节点 3001

发送币：
```shell
$ blockchain_go send -from WALLET_1 -to WALLET_3 -amount 1
$ blockchain_go send -from WALLET_2 -to WALLET_4 -amount 1
```

### 节点 3002
很快，转到矿机节点后，可以看到它挖出了新的区块。也检测了中央节点的output。

### 节点 3001
选择钱包节点，然后启动：

```shell
$ blockchain_go startnode
```
它会下载新挖出的区块。

停下来，然后检测余额：

```shell
$ blockchain_go getbalance -address WALLET_1
Balance of 'WALLET_1': 9

$ blockchain_go getbalance -address WALLET_2
Balance of 'WALLET_2': 9

$ blockchain_go getbalance -address WALLET_3
Balance of 'WALLET_3': 1

$ blockchain_go getbalance -address WALLET_4
Balance of 'WALLET_4': 1

$ blockchain_go getbalance -address MINER_WALLET
Balance of 'MINER_WALLET': 10
```

这里就是全部了！

## 结论
这是我们这个系列文章的最后一篇了。应该要实现真正的P2P原型网络，但是确实没有这么多的时间。希望这篇文章能解答一些你关于比特币技术的疑问，并且获得新姿势，你也可以自己去找到答案。还有很多有趣的知识藏在比特币技术中。祝你开车愉快！

P.S. 你可以开始通过实现**addr**消息来完善网络，就如比特币网络协议中描述的一样。这是消息非常重要，因为它可以让节点互相发现彼此。我已经开始实现它了，但是还没有完成。

## 相关链接

### 延伸与代码

1. [本文代码][本文代码]
2. [比特币协议](https://en.bitcoin.it/wiki/Protocol_documentation)
3. [比特币网络](https://en.bitcoin.it/wiki/Network)

### 本序列文章

1. [Golang 区块链入门 第一章 基本概念][本序列第一篇]
2. [Golang 区块链入门 第二章 工作量证明][本序列第二篇]
3. [Golang 区块链入门 第三章 持久化和命令行接口][本序列第三篇]
4. [Golang 区块链入门 第四章 交易 第一节][本序列第四篇]
5. [Golang 区块链入门 第五章 地址][本序列第五篇]
6. [Golang 区块链入门 第六章 交易 第二节][本序列第六篇]
7. [Golang 区块链入门 第七章 网络][本序列第七篇]

[本序列第一篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/05/abc-building-blockchain-in-go-part-1-basic-prototype/
[本序列第二篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/06/abc-building-blockchain-in-go-part-2-proof-of-work/
[本序列第三篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/07/abc-building-blockchain-in-go-part-3-persistence-and-cli/
[本序列第四篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/09/abc-building-blockchain-in-go-part-4-transactions-1/
[本序列第五篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/14/abc-building-blockchain-in-go-part-5-address/
[本序列第六篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/17/abc-building-blockchain-in-go-part-6-transactions-2/
[本序列第七篇]: https://printfcoder.github.io/myblog/blockchain/abc/2018/03/20/abc-building-blockchain-in-go-part-7-network/

[原文]: https://jeiwan.cc/posts/building-blockchain-in-go-part-7/
[本文代码]: https://github.com/printfcoder/blockchain-abc/tree/part_7