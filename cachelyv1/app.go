//go:generate protoc -I/usr/local/include -I/usr/local/go-global/1.12/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --proto_path=../_protos --gogo_out=plugins=grpc:. --grpc-gateway_out=logtostderr=true:. cachely/v1/cache_api.proto
package cachelyv1
