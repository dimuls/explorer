package explorer

//go:generate protoc --proto_path=. --proto_path=./.third_party/googleapis --go_out=. entity.proto
//go:generate protoc-go-inject-tag -input=./entity.pb.go

//go:generate protoc --proto_path=. --proto_path=./.third_party/googleapis --proto_path=./.third_party/protocolbuffers/src --proto_path=./.third_party/googleapis --go_out=. --go-grpc_out=. --grpc-gateway_out=. --openapiv2_out=. service.proto
