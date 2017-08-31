![logo](https://github.com/blacktop/graboid/raw/master/graboids.jpg)

graboid
=======

[![Circle CI](https://circleci.com/gh/blacktop/graboid.png?style=shield)](https://circleci.com/gh/blacktop/graboid) [![GitHub release](https://img.shields.io/github/release/blacktop/graboid.svg)](https://github.com/https://github.com/blacktop/graboid/releases/releases) [![License](http://img.shields.io/:license-mit-blue.svg)](http://doge.mit-license.org)

> ~~Client~~**LESS** Docker Image Downloader

---

Install
-------

-	On [Mac](https://github.com/blacktop/graboid/blob/master/docs/macos.md)
-	On [Linux](https://github.com/blacktop/graboid/blob/master/docs/linux.md)
-	On [Windows](https://github.com/blacktop/graboid/blob/master/docs/windows.md)

Why
---

This project was created for people whom can't install docker on their desktops, but still need to be able to download docker images from [DockerHUB](https://hub.docker.com) and then transfer them to another machine running docker.

Getting Started
---------------

```
Usage: graboid [OPTIONS] COMMAND [arg...]

Docker Image Downloader

Version: 0.12.0, BuildTime: 20170816
Author: blacktop - <https://github.com/blacktop>

Options:
  --verbose, -V     verbose output
  --index value     override index endpoint (default: "https://index.docker.io") [$GRABOID_INDEX]
  --registry value  override registry endpoint [$GRABOID_REGISTRY]
  --proxy value     HTTP/HTTPS proxy [$GRABOID_PROXY]
  --insecure        do not verify ssl certs
  --user value      registry username [$GRABOID_USERNAME]
  --password value  registry password [$GRABOID_PASSWORD]
  --help, -h        show help
  --version, -v     print the version

Commands:
  tags  List image tags
  help  Shows a list of commands or help for one command

Run 'graboid COMMAND --help' for more information on a command.
```

### List available image tags

```sh
$ graboid tags blacktop/scifgif
```

```sh
- Repository: blacktop/scifgif
- Tags:
    0.2.0
    latest
```

### Download the docker image `blacktop/scifgif`

```sh
$ graboid blacktop/scifgif:latest
```

Import image into docker

```sh
$ docker load -i blacktop_scifgif.tar
```

### Download with a **Proxy**

```sh
$ graboid --proxy http://proxy.org:[PORT] blacktop/scifgif:latest
```

### TODO

-	[ ] parallelize the layer downloads to decrease the total time to download large images
-	[ ] add image signature verification ([Notary](https://github.com/docker/notary)?)
-	[ ] ensure support for long connections for large downloads

### Issues

Find a bug? Want more features? Find something missing in the documentation? Let me know! Please don't hesitate to [file an issue](https://github.com/blacktop/graboid/issues/new)

### License

MIT Copyright (c) 2017 **blacktop**
