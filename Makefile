# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

VERSION = 1.0.0
GIT_REV_SHORT = $(shell git rev-parse --short HEAD)

generate-models:
	cd proto && protoc -I. --go_out=../zookeeper/ model/data.proto model/exchange.proto model/quorum.proto model/txn.proto model/persistance.proto

install:
	go install

build:
	go build -ldflags "-X main.GitRevString=$(GIT_REV_SHORT) -X main.Version=$(VERSION)" .
