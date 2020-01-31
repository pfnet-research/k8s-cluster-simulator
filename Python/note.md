protobuf compilation for golang

```
protoc --go_out=plugins=grpc:. k8s_sim.proto
```

protobuf compilation for python

```
python -m grpc_tools.protoc -I../protos --python_out=. --grpc_python_out=. ../protos/k8s_sim.proto
```