<p align="center">
  <a href="https://github.com/blacktop/graboid"><img alt="graboid Logo" src="https://github.com/blacktop/graboid/raw/master/docs/graboids.jpg" height="200" /></a>
  <h1 align="center">graboid</h1>
  <h4><p align="center"><del>Client</del><b>LESS</b> Docker Image Downloader</p></h4>
  <p align="center">
    <a href="https://github.com/blacktop/graboid/actions?query=workflow%3AGo" alt="Actions">
          <img src="https://github.com/blacktop/graboid/workflows/Go/badge.svg" /></a>
    <a href="https://github.com/blacktop/graboid/releases/latest" alt="Downloads">
          <img src="https://img.shields.io/github/downloads/blacktop/graboid/total.svg" /></a>
    <a href="https://github.com/blacktop/graboid/releases" alt="GitHub Release">
          <img src="https://img.shields.io/github/release/blacktop/graboid.svg" /></a>
    <a href="http://doge.mit-license.org" alt="LICENSE">
          <img src="https://img.shields.io/:license-mit-blue.svg" /></a>
</p>
<br>

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
$ docker load -i blacktop_scifgif.tar.gz
```

### Download with a **Proxy**

``` sh
$ graboid --proxy http://proxy.org:[PORT] blacktop/scifgif:latest
```

### Extract a file from the image's filesystem :construction: :new:

``` sh
$ graboid extract blacktop_ghidra_beta.tar.gz
```

> **NOTE:** Press `<enter>` to expand a layer and press `<space>` to extract file

![extract](https://github.com/blacktop/graboid/raw/master/docs/extract.png)

## TODO

* [ ] parallelize the layer downloads to decrease the total time to download large images
* [ ] add image signature verification ([Notary](https://github.com/docker/notary)?)
* [x] ensure support for long connections for large downloads

## Issues

Find a bug? Want more features? Find something missing in the documentation? Let me know! Please don't hesitate to [file an issue](https://github.com/blacktop/graboid/issues/new)

## License

MIT Copyright (c) 2017-2021 **blacktop**

