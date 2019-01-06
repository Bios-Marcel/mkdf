# mkdf

This little tool creates a `.desktop` file at `~/.local/share/applications`
which contains only some Information.

Currently the user can decide on the following content:

* Visible name
* Executable path
* Optional Icon

## Why

I was annoyed by manually creating those files, since some applications don't
do it themselves.

## Installation

```shell
go get github.com/Bios-Marcel/mkdf
```

Make sure that `$GOPATH/bin` is part of your `PATH` variable or manually add
the binary to your path.

## Usage

Simply invoke the tool, it will guide you!