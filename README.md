![logo](https://github.com/blacktop/graboid/raw/master/docs/graboids.jpg)

# graboid

[![Circle CI](https://circleci.com/gh/blacktop/graboid.png?style=shield)](https://circleci.com/gh/blacktop/graboid) [![Build status](https://ci.appveyor.com/api/projects/status/go99ieg0mqpmyi7g?svg=true)](https://ci.appveyor.com/project/blacktop/graboid) [![Github All Releases](https://img.shields.io/github/downloads/blacktop/graboid/total.svg)](https://github.com/blacktop/graboid/releases/latest) [![GitHub release](https://img.shields.io/github/release/blacktop/graboid.svg)](https://github.com/blacktop/graboid/releases) [![License](http://img.shields.io/:license-mit-blue.svg)](http://doge.mit-license.org)

> ~~Client~~**LESS** Docker Image Downloader

---

## Install

### macOS

``` bash
$ brew install blacktop/tap/graboid
```

### linux/windows

Download from [releases](https://github.com/blacktop/graboid/releases/latest)

## Why

This project was created for people whom can't install docker on their desktops, but still need to be able to download docker images from [DockerHUB](https://hub.docker.com) and then transfer them to another machine running docker.

## Getting Started

``` sh
$ graboid --help

Docker Image Downloader

Usage:
  graboid [flags]
  graboid [command]

Available Commands:
  extract     Extract files from image
  help        Help about any command
  tags        List image tags

Flags:
      --config string     config file (default is $HOME/.graboid.yaml)
  -h, --help              help for graboid
      --index string      override index endpoint (default "https://index.docker.io")
      --insecure          do not verify ssl certs
      --proxy string      HTTP/HTTPS proxy
      --registry string   override registry endpoint
  -V, --verbose           verbose output

Use "graboid [command] --help" for more information about a command.
```

### List available image tags

``` sh
$ graboid tags blacktop/scifgif

   • Querying Registry image=blacktop/scifgif
   • Tags:
      • 0.2.0
      • 0.3.0
      • 1.0
      • latest
```

### Download the docker image `blacktop/scifgif` 

``` sh
$ graboid blacktop/scifgif:latest
```

Import image into docker

``` sh
$ docker load -i blacktop_scifgif.tar
```

### Download with a **Proxy**

``` sh
$ graboid --proxy http://proxy.org:[PORT] blacktop/scifgif:latest
```

### Extract a file from the image's filesystem :construction: :new:

``` sh
$ graboid extract blacktop/ghidra:beta
```

> **NOTE:** Press `<enter>` to expand a layer and press `<space>` to extract file *(will also make surrounding folders)*

![extract](https://github.com/blacktop/graboid/raw/master/docs/extract.png)

## TODO

* [ ] parallelize the layer downloads to decrease the total time to download large images
* [ ] add image signature verification ([Notary](https://github.com/docker/notary)?)
* [x] ensure support for long connections for large downloads

## Issues

Find a bug? Want more features? Find something missing in the documentation? Let me know! Please don't hesitate to [file an issue](https://github.com/blacktop/graboid/issues/new)

## License

MIT Copyright (c) 2017 **blacktop**

