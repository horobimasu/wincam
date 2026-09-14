# go-cam
pure golang utility to take webcam pictures on windows with no other utilities

literally every single other golang webcam package for windows requires some other retarded dependencies to be installed like opencv or electron or wtv else, this dosnt need any of that as it just uses windows syscalls

this is not a package you can just require in your project btw, you need to copy the code in `main.go` and paste it in your project

(only tested on windows 11)
