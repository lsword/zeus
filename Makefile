.PHONY: install clean test save

PROGRAMNAME:=zeus
CURDIR:=$(shell pwd)
OLDGOPATH=$(GOPATH)

all: godep save install

godep:
	@export GOPATH=$(CURDIR); go get -u github.com/tools/godep

save:
	if [ ! -d $(CURDIR)/src/golang.org/x/crypto ]; then git clone https://github.com/golang/crypto.git $(CURDIR)/src/golang.org/x/crypto; fi
	if [ ! -d $(CURDIR)/src/golang.org/x/net ]; then git clone https://github.com/golang/net.git $(CURDIR)/src/golang.org/x/net; fi
	if [ ! -d $(CURDIR)/src/golang.org/x/text ]; then git clone https://github.com/golang/text.git $(CURDIR)/src/golang.org/x/text; fi
	@export GOPATH=$(CURDIR); go get -v ./...; cd $(CURDIR)/src/$(PROGRAMNAME); godep save

restore:
	@export GOPATH=$(CURDIR); cd $(CURDIR)/src/$(PROGRAMNAME); godep restore

install:
	@export GOPATH=$(CURDIR); cd $(CURDIR)/src/$(PROGRAMNAME); go build -o $(CURDIR)/bin/$(PROGRAMNAME)

test:
	@export GOPATH=$(CURDIR); cd $(CURDIR)/src/$(PROGRAMNAME); godep go test

clean:
	rm -rf bin pkg
