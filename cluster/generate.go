package cluster

//go:generate protoc -I=../actor --proto_path=. --go_out=. --go_opt=paths=source_relative ./*.proto
