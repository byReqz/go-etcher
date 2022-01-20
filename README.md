# go-etcher

go-etcher is a simple tool to create bootable drives that is mainly inspired by the deprecated  <a href="https://github.com/balena-io/etcher-cli">etcher-cli</a>

# Usage
```
go-etcher [image] [device]
```

arguments: 
```
  -d, --device string   target device
  -f, --force           override safety features
  -i, --input string    input file
```

If no image or device is specified, etcher will enter interactive mode and prompt for the missing parameters.

# Planned features
- Windows support
- prettier output
- image OS detection via <a href="https://github.com/byReqz/pt">pt</a>