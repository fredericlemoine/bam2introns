GO_EXECUTABLE := go
VERSION := $(shell git describe --abbrev=10 --dirty --always --tags)
DIST_DIRS := find * -type d -exec

all: build install

build:
	${GO_EXECUTABLE} build -o bam2introns -ldflags "-X github.com/fredericlemoine/bam2introns/cmd.Version=${VERSION}" github.com/fredericlemoine/bam2introns

install:
	${GO_EXECUTABLE} install -ldflags "-X github.com/fredericlemoine/bam2introns/cmd.Version=${VERSION}" github.com/fredericlemoine/bam2introns

test:
	${GO_EXECUTABLE} test github.com/fredericlemoine/bam2introns/tests/

deploy:
	mkdir -p deploy/${VERSION}
	env GOOS=windows GOARCH=amd64 ${GO_EXECUTABLE} build -o deploy/${VERSION}/bam2introns_amd64.exe -ldflags "-X github.com/fredericlemoine/bam2introns/cmd.Version=${VERSION}" github.com/fredericlemoine/bam2introns
	env GOOS=windows GOARCH=386 ${GO_EXECUTABLE} build -o deploy/${VERSION}/bam2introns_386.exe -ldflags "-X github.com/fredericlemoine/bam2introns/cmd.Version=${VERSION}" github.com/fredericlemoine/bam2introns
	env GOOS=darwin GOARCH=amd64 ${GO_EXECUTABLE} build -o deploy/${VERSION}/bam2introns_amd64_darwin -ldflags "-X github.com/fredericlemoine/bam2introns/cmd.Version=${VERSION}" github.com/fredericlemoine/bam2introns
	env GOOS=darwin GOARCH=386 ${GO_EXECUTABLE} build -o deploy/${VERSION}/bam2introns_386_darwin -ldflags "-X github.com/fredericlemoine/bam2introns/cmd.Version=${VERSION}" github.com/fredericlemoine/bam2introns
	env GOOS=linux GOARCH=amd64 ${GO_EXECUTABLE} build -o deploy/${VERSION}/bam2introns_amd64_linux -ldflags "-X github.com/fredericlemoine/bam2introns/cmd.Version=${VERSION}" github.com/fredericlemoine/bam2introns
	env GOOS=linux GOARCH=386 ${GO_EXECUTABLE} build -o deploy/${VERSION}/bam2introns_386_linux -ldflags "-X github.com/fredericlemoine/bam2introns/cmd.Version=${VERSION}" github.com/fredericlemoine/bam2introns
	tar -czvf deploy/${VERSION}.tar.gz --directory="deploy" ${VERSION}
