CODE_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
CWD := ${CURDIR}

FUNC_NAMES = iocRecord iocDetect entityRecord entityDetect crawlOTX crawlURLHaus
FUNCTIONS = $(foreach f,$(FUNC_NAMES),$(CODE_DIR)/build/$(f)/bootstrap)

SRC=$(CODE_DIR)/*.go $(CODE_DIR)/pkg/*/*.go

all: build

$(CODE_DIR)/build/%/bootstrap: $(SRC) $(CODE_DIR)/lambda/%/*.go
	env GOARCH=amd64 GOOS=linux go build -o $@ $(CODE_DIR)/lambda/$*/

build: $(FUNCTIONS)

clean:
	rm -rf $(CODE_DIR)/build
