.PHONY: require-% install-statik install-pilosa install

export GO111MODULE=on

# Installs statik.
install-statik:
	go get -u github.com/rakyll/statik

# installs demo-taxi binary in $GOPATH/bin
install: require-go require-statik require-pilosa
	go generate
	go install

install-pilosa:
	echo "see more https://www.pilosa.com/docs/latest/installation/"
	brew install pilosa

## COMMON (go)

# Prints instructions for installing go
install-go:
	@echo "Will not install Go automatically. Follow instructions at https://golang.org/doc/install"


fake-data:
	pypy tools/fake.py init
	bash tools/fake.sh

## COMMON
require-%:
	$(if $(shell which $* 2>/dev/null),\
		$(info Verified build dependency "$*" is installed.),\
		$(error Build dependency "$*" not installed. To install, try `make install-$*`))

