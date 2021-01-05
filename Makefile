CODE_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
CWD := ${CURDIR}

FUNCTIONS = \
	$(CODE_DIR)/build/iocRecord \
	$(CODE_DIR)/build/iocDetect \
	$(CODE_DIR)/build/entityRecord \
	$(CODE_DIR)/build/entityDetect \
	$(CODE_DIR)/build/crawlURLHaus

SRC=$(CODE_DIR)/*.go $(CODE_DIR)/pkg/*/*.go

all: build

$(CODE_DIR)/build/iocRecord: $(SRC) $(CODE_DIR)/lambda/iocRecord/*.go
	env GOARCH=amd64 GOOS=linux go build -o $(CODE_DIR)/build/iocRecord $(CODE_DIR)/lambda/iocRecord/
$(CODE_DIR)/build/iocDetect: $(SRC) $(CODE_DIR)/lambda/iocDetect/*.go
	env GOARCH=amd64 GOOS=linux go build -o $(CODE_DIR)/build/iocDetect $(CODE_DIR)/lambda/iocDetect/
$(CODE_DIR)/build/entityRecord: $(SRC) $(CODE_DIR)/lambda/entityRecord/*.go
	env GOARCH=amd64 GOOS=linux go build -o $(CODE_DIR)/build/entityRecord $(CODE_DIR)/lambda/entityRecord/
$(CODE_DIR)/build/entityDetect: $(SRC) $(CODE_DIR)/lambda/entityDetect/*.go
	env GOARCH=amd64 GOOS=linux go build -o $(CODE_DIR)/build/entityDetect $(CODE_DIR)/lambda/entityDetect/
$(CODE_DIR)/build/crawlURLHaus: $(SRC) $(CODE_DIR)/lambda/crawlURLHaus/*.go
	env GOARCH=amd64 GOOS=linux go build -o $(CODE_DIR)/build/crawlURLHaus $(CODE_DIR)/lambda/crawlURLHaus/

build: $(FUNCTIONS)

asset: build
	cp $(CODE_DIR)/build/* /asset-output

clean:
	rm -f $(FUNCTIONS)
