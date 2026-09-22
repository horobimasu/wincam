# win-cam
pure golang utility to take webcam pictures on windows with no other utilities (only tested on windows 11)

literally every single other golang webcam package for windows requires some other retarded dependencies to be installed like opencv or electron or wtv else, this doesnt need any of that as it just uses windows syscalls

# usage
install:
```bash
go get github.com/horobimasu/wincam
```

capture webcam:
```go
wincam.CaptureWebcam() (*bytes.Buffer, error)
```
