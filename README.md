# Golang 区块链入门

## 概述

这个项目是基于Golang的区块链实现，下面章节逐步由浅入深，从原型、工作量证明、持久化、交易、地址等实现了区块链技术。该项目并不是能够接入比特币网络的区块链解决方案，而只是旨在向大家阐述其工作原理。

该项目大体从[原项目](https://github.com/Jeiwan/blockchain_go)翻译及演译过来。会有部分改动，对每一个方法，容易引起疑问的地方增加了详细注释，也把代码与其对应的文章一一比对，并把文章中的重点写到代码注释中，帮助大家基于代码来理解区块链，从而更加了解原理。

master分支是基于第七章而来。

有任何问题可与我邮件沟通，谢谢大家！

## 本序列文章

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