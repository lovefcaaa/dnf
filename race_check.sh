cur_path=`pwd`
export GOPATH=$cur_path:$GOPATH

go run -race src/main.go
