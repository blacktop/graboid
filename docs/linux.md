Install `graboid` for Linux
===========================

Download latest version from Github [releases](https://github.com/blacktop/graboid/releases)
--------------------------------------------------------------------------------------------

Run this command to download `graboid`, replacing `VERSION` with the specific version of `graboid` you want to use:

```sh
$ VERSION=0.7.0
$ curl -L "https://github.com/blacktop/graboid/releases/download/${VERSION}/graboid_${VERSION}_linux_amd64.tar.gz" \
  | tar -xzf - -C /usr/local/bin/
```
