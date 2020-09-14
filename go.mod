module github.com/blacktop/graboid

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20200826222420-2800ab02245e
	github.com/docker/docker => github.com/docker/docker v17.12.0-ce-rc1.0.20200913044305-bf6f0d82bc24+incompatible
)

require (
	github.com/apex/log v1.1.1
	github.com/blacktop/ipsw v0.0.0-20190907012325-eda024ad7908
	github.com/docker/docker v17.12.0-ce-rc1.0.20200913044305-bf6f0d82bc24+incompatible
	github.com/docker/go-units v0.4.0 // indirect
	github.com/dustin/go-humanize v1.0.0
	github.com/gizak/termui/v3 v3.1.0
	github.com/mattn/go-isatty v0.0.9 // indirect
	github.com/mattn/go-runewidth v0.0.4 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/nsf/termbox-go v0.0.0-20190817171036-93860e161317 // indirect
	github.com/opencontainers/go-digest v1.0.0
	github.com/phayes/permbits v0.0.0-20190612203442-39d7c581d2ee // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/urfave/cli v1.22.4
	github.com/wagoodman/dive v0.9.3-0.20200914104330-2db716d1919f
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/cheggaaa/pb.v1 v1.0.16
)

go 1.13
