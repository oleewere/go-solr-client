generate-models:
	cd proto && protoc -I. --go_out=../zookeeper/ model/data.proto model/exchange.proto model/quorum.proto model/txn.proto model/persistance.proto

install:
	go install
