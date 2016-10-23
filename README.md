### QLB reader

#### Initial setup (Go)

Make sure GOPATH is set.

```
# in ~/.profile, ~/.bashrc, or ~/.bash_profile
export GOPATH=~/go
```

Fetch the packages referenced by qlb.go.

```
$ go get .
```

#### Running

```
go build
./qlb [-o outdir] indir
```

or

```
go run qlb.go [-o outdir] indir
```
