![logo](https://github.com/blacktop/graboid/raw/master/graboids.jpg)

graboid
=======

[![Circle CI](https://circleci.com/gh/blacktop/graboid.png?style=shield)](https://circleci.com/gh/blacktop/graboid) [![License](http://img.shields.io/:license-mit-blue.svg)](http://doge.mit-license.org)

> Client-less Docker Image Downloader

---

Getting Started
---------------

Install on macOS

```sh
$ brew install blacktop/tap/graboid
```

Download docker image `blacktop/scifgif`

```sh
$ graboid blacktop/scifgif
```

Import into docker

```sh
$ docker load -i scifgif.tar
```

### Issues

Find a bug? Want more features? Find something missing in the documentation? Let me know! Please don't hesitate to [file an issue](https://github.com/blacktop/graboid/issues/new)

### License

MIT Copyright (c) 2017 **blacktop**