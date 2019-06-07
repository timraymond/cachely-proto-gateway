//go:generate protoc --proto_path=../_protos --gogo_out=plugins=grpc:. --grpc-gateway_out=logtostderr=true:. cachely/v1/cache_api.proto
package cachelyv1
