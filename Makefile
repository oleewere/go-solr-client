generate-models:
	protoc -I. --go_out=./ model/zookeeper-data.proto model/zookeeper-proto.proto model/zookeeper-quorum.proto model/zookeeper-txn.proto model/zookeeper-persistance.proto

install:
	go install