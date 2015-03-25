cur_path=`pwd`
export GOPATH=$cur_path:$GOPATH

go test attribute
go build -o bin/dnf src/main.go
