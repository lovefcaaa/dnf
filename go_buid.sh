cur_path=`pwd`
export GOPATH=$cur_path:$GOPATH

go build -o bin/dnf src/main.go
