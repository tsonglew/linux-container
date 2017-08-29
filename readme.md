# xperiMoby

An Experiment with Linux Container

## Install

```shell
$ go get https://github.com/kasheemlew/xperiMoby; cd $GOPATH/src/github.com/kasheemlew/xperiMoby; go build .
```

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
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```
