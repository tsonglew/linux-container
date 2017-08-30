# xperiMoby

An Experiment with Linux Container

## Get Ubuntu

[Download](https://www.ubuntu.com/download) Ubuntu distribution 

## Get Go

```shell
$ wget https://storage.googleapis.com/golang/go1.9.linux-amd64.tar.gz
$ tar -C /usr/local -xzf go1.9.linux-amd64.tar.gz
```

## Set Go Environment

```shell
export PATH=$PATH:/usr/local/go/bin
export GOPATH=</path/to/go/workspace>
export PATH=$PATH:$GOPATH/bin
export GOROOT="/usr/local/go"
export GOROOT_BOOTSTRAP=$GOROOT
```

## Get dep (golang dependence manager)

```
$ go get -u github.com/golang/dep/cmd/dep
$ cd $GOPATH/src/golang/dep/
$ go build
```

## Install

```shell
$ go get https://github.com/kasheemlew/xperiMoby
$ cd $GOPATH/src/github.com/kasheemlew/xperiMoby
$ dep init
$ go build
$ ln xperiMoby /usr/local/sbin/xm
```

## Set xperiMoby working directory

```shell
$ mkdir -p /root/xperi
```

## Get busybox image

download busybox image and put it into /root/xperi as `busybox.tar`

## Usage

```shell
NAME:
   xperiMoby - LXC runtime implemention

USAGE:
   xperiMoby [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
     init     Init container process run user's process in container
     run      Create container with namespace and cgroup limit
                  xperiMoby run -ti [command]
     commit   commit a container into image
     ps       list all the containers
     logs     print logs of a container
     exec     exec a command into container
     stop     stop a container
     rm       remove unused containers
     network  container network commands
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

## Example

```shell
$ xm run -ti --name containerName -v /root/test:/innercontainer busybox sh
{"level":"info","msg":"command all is sh","time":"2017-08-30T11:39:39+08:00"}
{"level":"info","msg":"init come on","time":"2017-08-30T11:39:39+08:00"}
{"level":"info","msg":"Current location is /root/xperi/mnt/x7ljx8res5","time":"2017-08-30T11:39:39+08:00"}
{"level":"info","msg":"Find path /bin/sh","time":"2017-08-30T11:39:39+08:00"}
/ #
```

## Help

Get help of `xm` command

```shell
$ xm --help
```

Get help of subcommand

```shell
$ xm help run
NAME:
   xperiMoby run - Create container with namespace and cgroup limit
      xperiMoby run -ti [command]

USAGE:
   xperiMoby run [command options] [arguments...]

OPTIONS:
   --ti              enable tty
   -d                detach container
   -m value          memory limit
   --CPUshare value  CPUshare limit
   --CPUset value    CPUset limit
   -v value          volume
   --name value      container name
   -e value          set environment
   --net value       container network
   -p value          port mapping

``