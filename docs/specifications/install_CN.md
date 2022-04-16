
<h1 align="center">Ontology </h1>
<p align="center" class="version">Version 0.7.0 </p>

[![GoDoc](https://godoc.org/github.com/cntmio/cntmology?status.svg)](https://godoc.org/github.com/cntmio/cntmology)
[![Go Report Card](https://goreportcard.com/badge/github.com/cntmio/cntmology)](https://goreportcard.com/report/github.com/cntmio/cntmology)
[![Travis](https://travis-ci.org/cntmio/cntmology.svg?branch=master)](https://travis-ci.org/cntmio/cntmology)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/cntmio/cntmology?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

[English](install.md) | 中文

## 构建开发环境
成功编译cntmology需要以下准备：

* Golang版本在1.9及以上
* 安装第三方包管理工具glide
* 正确的Go语言开发环境
* Golang所支持的操作系统

## 部署|获取cntmology
### 从源码获取
克隆cntmology仓库到 **$GOPATH/src/github.com/cntmio** 目录

```shell
$ git clone https://github.com/cntmio/cntmology.git
```
或者
```shell
$ go get github.com/cntmio/cntmology
```

用第三方包管理工具glide拉取依赖库

````shell
$ cd $GOPATH/src/github.com/cntmio/cntmology
$ glide install
````

用make编译源码

```shell
$ make
```

成功编译后会生成两个可以执行程序

* `cntmology`: 节点程序/以命令行方式提供的节点控制程序

### 从release获取
You can download at [release page](https://github.com/cntmio/cntmology/releases).