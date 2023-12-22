CODE_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
CWD := ${CURDIR}

FUNCTIONS = \
	$(CODE_DIR)/build/iocRecord/bootstrap \
	$(CODE_DIR)/build/iocDetect/bootstrap \
	$(CODE_DIR)/build/entityRecord/bootstrap \
	$(CODE_DIR)/build/entityDetect/bootstrap \
	$(CODE_DIR)/build/crawlOTX/bootstrap \
	$(CODE_DIR)/build/crawlURLHaus/bootstrap

SRC=$(CODE_DIR)/*.go $(CODE_DIR)/pkg/*/*.go

all: build

$(CODE_DIR)/build/iocRecord/bootstrap: $(SRC) $(CODE_DIR)/lambda/iocRecord/*.go
	env GOARCH=amd64 GOOS=linux go build -o $(CODE_DIR)/build/iocRecord/bootstrap $(CODE_DIR)/lambda/iocRecord/
$(CODE_DIR)/build/iocDetect/bootstrap: $(SRC) $(CODE_DIR)/lambda/iocDetect/*.go
	env GOARCH=amd64 GOOS=linux go build -o $(CODE_DIR)/build/iocDetect/bootstrap $(CODE_DIR)/lambda/iocDetect/
$(CODE_DIR)/build/entityRecord/bootstrap: $(SRC) $(CODE_DIR)/lambda/entityRecord/*.go
	env GOARCH=amd64 GOOS=linux go build -o $(CODE_DIR)/build/entityRecord/bootstrap $(CODE_DIR)/lambda/entityRecord/
$(CODE_DIR)/build/entityDetect/bootstrap: $(SRC) $(CODE_DIR)/lambda/entityDetect/*.go
	env GOARCH=amd64 GOOS=linux go build -o $(CODE_DIR)/build/entityDetect/bootstrap $(CODE_DIR)/lambda/entityDetect/
$(CODE_DIR)/build/crawlURLHaus/bootstrap: $(SRC) $(CODE_DIR)/lambda/crawlURLHaus/*.go
	env GOARCH=amd64 GOOS=linux go build -o $(CODE_DIR)/build/crawlURLHaus/bootstrap $(CODE_DIR)/lambda/crawlURLHaus/
$(CODE_DIR)/build/crawlOTX/bootstrap: $(SRC) $(CODE_DIR)/lambda/crawlOTX/*.go
	env GOARCH=amd64 GOOS=linux go build -o $(CODE_DIR)/build/crawlOTX/bootstrap $(CODE_DIR)/lambda/crawlOTX/

build: $(FUNCTIONS)

clean:
	rm -f $(FUNCTIONS)
