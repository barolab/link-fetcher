PWD:=$(shell pwd)

test:
	@dagger call test

fmt:
	@dagger call fmt export --path=.

build:
	@dagger call build

scan:
	@dagger call scan

lint:
	@dagger call lint
