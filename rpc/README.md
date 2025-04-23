## gRPC API for Horizon

## To generate *.pb.go files
```bash
protoc -I rpc/protos rpc/protos/*.proto --go_out=rpc/protos_pb/ --go-grpc_out=rpc/protos_pb/ --go-grpc_opt=paths=source_relative --experimental_allow_proto3_optional
```

## To generate gRPC documentation
```bash
protoc -I rpc/protos rpc/protos/*.proto --doc_out=./rpc/docs --doc_opt=html,index.html  --experimental_allow_proto3_optional
protoc -I rpc/protos rpc/protos/*.proto --doc_out=./rpc/docs --doc_opt=markdown,index.md  --experimental_allow_proto3_optional
```