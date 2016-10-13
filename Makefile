ifeq ($(TMPDIR),)
	TMPDIR := /tmp
endif

export PKGNAME := ip-out
export TMPGOPATH := $(TMPDIR)/$(PKGNAME)-gopath
export PKGPATH := $(TMPGOPATH)/src/github.com/open-horizon
export PATH := $(TMPGOPATH)/bin:$(PATH)
export VERSION := $(shell cat VERSION)
export ARCH := $(shell uname -m)

SHELL := /bin/bash
ARCH = $(shell uname -m)
PKGS=$(shell cd $(PKGPATH)/$(PKGNAME); GOPATH=$(TMPGOPATH) go list ./... | gawk '$$1 !~ /vendor\// {print $$1}')

COMPILE_ARGS := CGO_ENABLED=0 GOOS=linux
# TODO: handle other ARM architectures on build boxes too
ifeq ($(ARCH),armv7l)
	COMPILE_ARGS +=  GOARCH=arm GOARM=7
endif

all: $(PKGNAME)-$(ARCH)-$(VERSION)

# will always run b/c deps target is PHONY
$(PKGNAME)-$(ARCH)-$(VERSION): gopathlinks
	cd $(PKGPATH)/$(PKGNAME) && \
	  export GOPATH=$(TMPGOPATH); \
	    $(COMPILE_ARGS) go build -o $(PKGNAME)-$(ARCH)-$(VERSION)
			ln -s $(PKGNAME)-$(ARCH)-$(VERSION) $(PKGNAME)

clean:
	find ./vendor -maxdepth 1 -not -path ./vendor -and -not -iname "vendor.json" -print0 | xargs -0 rm -Rf
	rm -f $(PKGNAME)*
	rm -rf $(TMPGOPATH)

# this is a symlink to facilitate building outside of user's GOPATH
gopathlinks:
	mkdir -p $(PKGPATH)
	rm -f $(PKGPATH)/$(PKGNAME)
	ln -s $(CURDIR) $(PKGPATH)/$(PKGNAME)

install: $(PKGNAME)
	mkdir -p $(DESTDIR)/bin
	cp -L $(PKGNAME) $(DESTDIR)/bin/$(PKGNAME)

lint:
	-golint ./... | grep -v "vendor/"
	-go vet ./... 2>&1 | grep -vP "exit\ status|vendor/"

.PHONY: clean deps gopathlinks install lint
