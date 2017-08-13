![logo](https://github.com/blacktop/graboid/raw/master/graboids.jpg)

graboid
=======

[![Circle CI](https://circleci.com/gh/blacktop/graboid.png?style=shield)](https://circleci.com/gh/blacktop/graboid) [![GoDoc](https://godoc.org/github.com/blacktop/graboid?status.svg)](https://godoc.org/github.com/blacktop/graboid) [![License](http://img.shields.io/:license-mit-blue.svg)](http://doge.mit-license.org)

> Client-less Docker Image Downloader

---

Getting Started
---------------

### Install on macOS

```sh
$ brew install blacktop/tap/graboid
```

```sh
Usage: graboid [OPTIONS] COMMAND [arg...]

Docker Image Downloader

Version: 0.4.0, BuildTime: 20170812
Author: blacktop - <https://github.com/blacktop>

Options:
  --verbose, -V               verbose output
  --timeout value             elasticsearch timeout (in seconds) (default: 60) [$TIMEOUT]
  --index value, -i value     override index endpoint (default: "https://index.docker.io") [$GRABOID_INDEX]
  --registry value, -r value  override registry endpoint [$GRABOID_REGISTRY]
  --user value, -u value      registry username [$GRABOID_USERNAME]
  --password value, -p value  registry password [$GRABOID_PASSWORD]
  --help, -h                  show help
  --version, -v               print the version

Commands:
  tags  List image tags
  help  Shows a list of commands or help for one command

Run 'graboid COMMAND --help' for more information on a command.
```

### Download docker image `blacktop/scifgif`

```sh
$ graboid blacktop/scifgif:latest
```

### Import into docker

```sh
$ docker load -i blacktop_scifgif.tar
```

### Issues

Find a bug? Want more features? Find something missing in the documentation? Let me know! Please don't hesitate to [file an issue](https://github.com/blacktop/graboid/issues/new)

### License

MIT Copyright (c) 2017 **blacktop**
