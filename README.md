# AeroInfoMonitor

Depending on whether you want to run this on linux, windows or Apple OSX, grab the correct executable binary from one of the directories (linux/, windows/, osx/).

All source code for inspection is in main.go

## Usage intro

This is a self-contained binary. It means, to run it e.g. on linux, download linux/AeroInfoMonitor to e.g. /root/, `chmod 755 /root/AeroInfoMonitor`, and you can then run and use it. It is a complied binary.

## Usage

```
Usage: ./osx/AeroInfoMonitor NodeIP NodePORT NamespaceName [username] [password]
```

All this does is query aerospike database for migrations and cluster size and prints it every 50 milliseconds, per node.
