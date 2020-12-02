CODE_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
CWD := ${CURDIR}

COMMON=$(CODE_DIR)/*.go $(CODE_DIR)/internal/*/*.go

FUNCTIONS = \
	$(CODE_DIR)/build/iocRecord \
	$(CODE_DIR)/build/iocDetect \
	$(CODE_DIR)/build/entityRecord \
	$(CODE_DIR)/build/entityDetect \
	$(CODE_DIR)/build/crawlURLHouse

SRC=$(CODE_DIR)/pkg/*/*.go
CDK_STACK=$(CODE_DIR)/lib/remo2cw-stack.js
CDK_SRC=$(CODE_DIR)/lib/remo2cw-stack.ts

all: deploy

$(CODE_DIR)/build/iocRecord: $(SRC) $(CODE_DIR)/lambda/iocRecord/*.go
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CODE_DIR)/build/iocRecord $(CODE_DIR)/lambda/iocRecord/ && cd $(CWD)
$(CODE_DIR)/build/iocDetect: $(SRC) $(CODE_DIR)/lambda/iocDetect/*.go
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CODE_DIR)/build/iocDetect $(CODE_DIR)/lambda/iocDetect/ && cd $(CWD)
$(CODE_DIR)/build/entityRecord: $(SRC) $(CODE_DIR)/lambda/entityRecord/*.go
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CODE_DIR)/build/entityRecord $(CODE_DIR)/lambda/entityRecord/ && cd $(CWD)
$(CODE_DIR)/build/entityDetect: $(SRC) $(CODE_DIR)/lambda/entityDetect/*.go
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CODE_DIR)/build/entityDetect $(CODE_DIR)/lambda/entityDetect/ && cd $(CWD)
$(CODE_DIR)/build/crawlURLHouse: $(SRC) $(CODE_DIR)/lambda/crawlURLHouse/*.go
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CODE_DIR)/build/crawlURLHouse $(CODE_DIR)/lambda/crawlURLHouse/ && cd $(CWD)

build: $(FUNCTIONS)

$(CDK_STACK): $(CDK_SRC)
	cd $(CODE_DIR) && tsc && cd $(CWD)

deploy: $(FUNCTIONS) $(CDK_STACK)
	cdk deploy
