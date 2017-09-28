Install `graboid` for macOS
===========================

Using [homebrew](https://brew.sh)
---------------------------------

```sh
$ brew install blacktop/tap/graboid
```

Download latest version from Github [releases](https://github.com/blacktop/graboid/releases)
--------------------------------------------------------------------------------------------

Run this command to download `graboid`, replacing `VERSION` with the specific version of `graboid` you want to use:

```sh
$ VERSION=0.14.0
$ curl -L "https://github.com/blacktop/graboid/releases/download/${VERSION}/graboid_${VERSION}_macOS_amd64.tar.gz" \
  | tar -xzf - -C /usr/local/bin/
```
