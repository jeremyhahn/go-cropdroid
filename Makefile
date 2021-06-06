ORG                     := automatethethingsllc
TARGET_OS               := linux
TARGET_ARCH             := $(shell uname -m)

CORES                   ?= $(shell nproc)
ARCH                    := $(shell go env GOARCH)
OS                      := $(shell go env GOOS)
BUILDDIR                := $(shell pwd)
TIMESTAMP               := $(shell date +"%Y-%m-%d_%H-%M-%S")
NUMPROC                 := $(shell nproc)

GOBIN                   := $(shell dirname `which go`)

ARM_CC                  ?= arm-linux-gnueabihf-gcc-8
ARM_CC_64				?= aarch64-linux-gnu-gcc

APP                     := cropdroid
APPTYPE					?= standalone
ENV             		?= dev
WEBSERVER_USER          ?= www-data

CROPDROID_VERSION       ?= $(shell git describe --tags --abbrev=0)
GIT_TAG                 = $(shell git describe --tags)
GIT_HASH                = $(shell git rev-parse HEAD)
BUILD_DATE              = $(shell date '+%y-%m-%d_%H:%M:%S')

LDFLAGS=-X github.com/jeremyhahn/$(APP)/app.Image=${IMAGE_NAME}
LDFLAGS+= -X github.com/jeremyhahn/$(APP)/app.Environment=${ENV}
LDFLAGS+= -X github.com/jeremyhahn/$(APP)/app.Release=${CROPDROID_VERSION}
LDFLAGS+= -X github.com/jeremyhahn/$(APP)/app.GitHash=${GIT_HASH}
LDFLAGS+= -X github.com/jeremyhahn/$(APP)/app.GitTag=${GIT_TAG}
LDFLAGS+= -X github.com/jeremyhahn/$(APP)/app.BuildUser=${USER}
LDFLAGS+= -X github.com/jeremyhahn/$(APP)/app.BuildDate=${BUILD_DATE}

ROCKSDB_HOME    ?= /rocksdb
ROCKSDB_INCLUDE ?= $(ROCKSDB_HOME)/include

.PHONY: deps 

default: build-standalone

certs:
	mkdir -p keys/
	openssl req -new -newkey rsa:4096 -days 365 -nodes -x509 -keyout keys/key.pem -out keys/cert.pem \
          -subj "/C=US/ST=MA/L=Boston/O=Automate The Things, LLC/CN=localhost"
	openssl genrsa -out keys/rsa.key 2048
	openssl rsa -in keys/rsa.key -pubout -out keys/rsa.pub

benchmark-api:
	gobench -u http://localhost:8091/status -k=true -c 500 -t 20

get-deps:
	go get

rocksdb-deps:
	sudo apt-get install -y libgflags-dev libsnappy-dev zlib1g-dev libbz2-dev libzstd-dev liblz4-dev

rocksdb-deps-arm64:
	sudo apt-get install -y libgflags-dev:arm64 libsnappy-dev:arm64 zlib1g-dev:arm64 libbz2-dev:arm64 libzstd-dev:arm64 liblz4-dev:arm64

cockroachdb-deps:
	sudo apt-get install -y libncurses-dev cmake

gorocksdb:
	CGO_CFLAGS="-I${ROCKSDB_INCLUDE}" \
	CGO_LDFLAGS="-L${ROCKSDB_HOME} -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd -ldl" \
	GOOS=linux $(GOBIN)/go get github.com/tecbot/gorocksdb

#gorocksdb-cc:
#	CGO_CFLAGS="-I${ROCKSDB_INCLUDE}" \
#	CGO_LDFLAGS="-L${ROCKSDB_HOME} -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd -ldl" \
#	env CC=$(ARM_CC) GOARCH=arm64 GOOS=linux $(GOBIN)/go get github.com/tecbot/gorocksdb

arm-deps:
	sudo apt-get install -y gcc-6-arm-linux-gnueabihf

build-deps:
	sudo apt-get install -y build-essential qemu qemu-user-static binfmt-support qemu-user-binfmt cloud-utils


build-standalone:
	$(GOBIN)/go build -o $(APP) -ldflags="-w -s ${LDFLAGS}"

build-standalone-debug:
	$(GOBIN)/go build -gcflags "-N" -o $(APP) -gcflags='all=-N -l' -ldflags="-w -s ${LDFLAGS}"

build-standalone-static:
	CGO_ENABLED=1 $(GOBIN)/go build -o $(APP) --ldflags '-w -s -extldflags -static -v ${LDFLAGS}'

build-standalone-debug-static:
	CGO_ENABLED=1 $(GOBIN)/go build -o $(APP) -gcflags='all=-N -l' --ldflags '-extldflags -static -v ${LDFLAGS}'


build-standalone-arm:
	CC=$(ARM_CC) GOOS=linux GOARCH=arm GOARM=6 \
	$(GOBIN)/go build -o $(APP) -ldflags="-w -s ${LDFLAGS}"

build-standalone-static-arm:
	CC=$(ARM_CC) CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=6 \
	$(GOBIN)/go build --tags="standalone" -v -a -o $(APP) -v --ldflags '-w -s -extldflags -static -v ${LDFLAGS}'

build-standalone-debug-arm:
	GOOS=linux GOARCH=arm GOARM=6 $(GOBIN)/go build -gcflags "all=-N -l" -o $(APP) -ldflags $(LDFLAGS)


build-standalone-arm64:
	CC=$(ARM_CC_64) GOOS=linux GOARCH=arm64 \
	$(GOBIN)/go build -o $(APP) -ldflags="-w -s ${LDFLAGS}"

build-standalone-static-arm64:
	CC=$(ARM_CC_64) CGO_ENABLED=1 GOOS=linux GOARCH=arm64 \
	$(GOBIN)/go build -o $(APP) --ldflags '-w -s -extldflags -static -v ${LDFLAGS}'

build-standalone-debug-arm64:
	CC=$(ARM_CC_64) GOOS=linux GOARCH=arm64 $(GOBIN)/go build -gcflags "all=-N -l" -o $(APP) --ldflags="$(LDFLAGS)"


build-cluster: build-cluster-pebble

build-cluster-pebble:
	# Uses default dragonboat pebble database (remove rocksdb config.ExpertConfig in cluster/raft_pebble.go)
	$(GOBIN)/go build --tags="cluster pebble" -o $(APP) -ldflags="-w -s ${LDFLAGS}"

build-cluster-pebble-debug:
	# Uses default dragonboat pebble database (remove rocksdb config.ExpertConfig in cluster/raft_pebble.go)
	$(GOBIN)/go build --tags="cluster pebble" -gcflags "all=-N -l" -o $(APP) -ldflags="$(LDFLAGS)"

build-cluster-pebble-static:
	# Uses default dragonboat pebble database (remove rocksdb config.ExpertConfig in cluster/raft_pebble.go)
	CGO_ENABLED=1 GOOS=linux $(GOBIN)/go build --tags="cluster pebble" -o $(APP) --ldflags '-w -s -extldflags -static -v ${LDFLAGS}'

build-cluster-pebble-debug-static:
	# Uses default dragonboat pebble database (remove rocksdb config.ExpertConfig in cluster/raft_pebble.go)
	$(GOBIN)/go build --tags="cluster pebble" -gcflags "all=-N -l" -o $(APP) --ldflags '-extldflags -static -v ${LDFLAGS}'

build-cluster-pebble-static-arm64:
	# Uses default dragonboat pebble database (remove rocksdb config.ExpertConfig in cluster/raft_pebble.go)
	CC=$(ARM_CC_64) CGO_ENABLED=1 GOOS=linux GOARCH=arm64 \
	$(GOBIN)/go build --tags="cluster pebble" -o $(APP) --ldflags '-w -s -extldflags -static -v ${LDFLAGS}'


build-cluster-rocksdb:
	# Uses dragonboat rocksdb database (rocksdb config.ExpertConfig in cluster/raft_rocksdb.go)
	CGO_CFLAGS="-I${ROCKSDB_INCLUDE}" \
	CGO_LDFLAGS="-L${ROCKSDB_HOME} -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd -ldl" \
	GOOS=linux CGO_ENABLED=1 $(GOBIN)/go build --tags="cluster rocksdb" -o $(APP) -ldflags="-w -s ${LDFLAGS}"

build-cluster-rocksdb-debug:
	# Uses dragonboat rocksdb database (rocksdb config.ExpertConfig in cluster/raft_rocksdb.go)
	CGO_CFLAGS="-I${ROCKSDB_INCLUDE}" \
	CGO_LDFLAGS="-L${ROCKSDB_HOME} -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd -ldl" \
	GOOS=linux CGO_ENABLED=1 $(GOBIN)/go build --tags="cluster rocksdb" -gcflags "all=-N -l" -o $(APP) -ldflags="$(LDFLAGS)"

build-cluster-rocksdb-debug-static:
	# Uses dragonboat rocksdb database (rocksdb config.ExpertConfig in cluster/raft_rocksdb.go)
	CGO_CFLAGS="-I${ROCKSDB_INCLUDE}" \
	CGO_LDFLAGS="-L${ROCKSDB_HOME} -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd -ldl" \
	GOOS=linux CGO_ENABLED=1 $(GOBIN)/go build -a --tags="cluster rocksdb" -gcflags "all=-N -l" -o $(APP) -v --ldflags '-extldflags -static ${LDFLAGS}'

# build-cluster-rocksdb-static-arm64:
# 	CC=$(ARM_CC_64) CGO_ENABLED=1 GOOS=linux GOARCH=arm64 \
# 	CGO_CFLAGS="-I${ROCKSDB_INCLUDE}" \
# 	CGO_LDFLAGS="-L${ROCKSDB_HOME} -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd -ldl" \
# 	CGO_ENABLED=1 GOOS=linux $(GOBIN)/go build --tags="cluster pebble" -o $(APP) --ldflags '-w -s -extldflags -static ${LDFLAGS}'

clean:
	$(GOBIN)/go clean
	rm -rf $(APP) \
		$(APP).log \
		/usr/local/bin/$(APP) \
		vendor \
		db/$(APP).db \
		db/cluster

unittest:
	cd test && $(GOBIN)/go test -v

integrationtest:
	cd datastore/gorm && $(GOBIN)/go test -v -tags integration

standalone:
	./$(APP) standalone --debug --ssl=false --port 8091

standalone-sqlite:
	./$(APP) standalone --debug --ssl=false --port 8091 --datastore sqlite

standalone-cockroach:
	./$(APP) standalone --debug --ssl=false --port 8091 --datastore cockroach

initlog:
	sudo touch /var/log/cropdroid.log && sudo chmod 777 /var/log/cropdroid.log
	sudo mkdir -p /var/log/cropdroid/cluster
	sudo touch /var/log/cropdroid/cluster/node-1.log && sudo chmod 777 /var/log/cropdroid/cluster/node-1.log
	sudo touch /var/log/cropdroid/cluster/node-2.log && sudo chmod 777 /var/log/cropdroid/cluster/node-2.log
	sudo touch /var/log/cropdroid/cluster/node-3.log && sudo chmod 777 /var/log/cropdroid/cluster/node-3.log