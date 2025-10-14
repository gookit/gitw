# link https://github.com/humbug/box/blob/master/Makefile
#SHELL = /bin/sh
.DEFAULT_GOAL := help
# 每行命令之前必须有一个tab键。如果想用其他键，可以用内置变量.RECIPEPREFIX 声明
# mac 下这条声明 没起作用 !!
#.RECIPEPREFIX = >
.PHONY: all usage help clean

# 需要注意的是，每行命令在一个单独的shell中执行。这些Shell之间没有继承关系。
# - 解决办法是将两行命令写在一行，中间用分号分隔。
# - 或者在换行符前加反斜杠转义 \

# 接收命令行传入参数 make COMMAND tag=v2.0.4
# TAG=$(tag)

BIN_NAME=chlog
MAIN_SRC_FILE=main.go
ROOT_PACKAGE := main
VERSION=$(shell git for-each-ref refs/tags/ --count=1 --sort=-version:refname --format='%(refname:short)' 1 |  sed 's/^v//')

# Full build flags used when building binaries. Not used for test compilation/execution.
BUILD_FLAGS := -ldflags \
  " -X $(ROOT_PACKAGE).Version=$(VERSION)"

##there some make command for the project
##

help:
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//' | sed -e 's/: / /'

##Available Commands:

ins2bin: ## Install to GOPATH/bin
	cd cmd/chlog && go build $(BUILD_FLAGS) -o $(GOPATH)/bin/chlog $(MAIN_SRC_FILE)
	chmod +x $(GOPATH)/bin/chlog

build-all:linux arm win darwin ## Build for Linux,ARM,OSX,Windows

linux: ## Build for Linux
	cd cmd/chlog && GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o chlog-linux-amd64 $(MAIN_SRC_FILE)
	mv cmd/chlog/chlog-linux-amd64 build/chlog-linux-amd64
	chmod +x build/chlog-linux-amd64

arm: ## Build for ARM
	cd cmd/chlog && GOOS=linux GOARCH=arm go build $(BUILD_FLAGS) -o chlog-linux-arm $(MAIN_SRC_FILE)
	mv cmd/chlog/chlog-linux-arm build/chlog-linux-arm
	chmod +x build/chlog-linux-arm

darwin: ## Build for OSX
	cd cmd/chlog && GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o chlog-darwin-amd64 $(MAIN_SRC_FILE)
	mv cmd/chlog/chlog-darwin-amd64 build/chlog-darwin-amd64
	chmod +x build/chlog-darwin-amd64

win: ## Build for Windows
	cd cmd/chlog && GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o chlog-windows-amd64.exe $(MAIN_SRC_FILE)
	mv cmd/chlog/chlog-windows-amd64.exe build/chlog-windows-amd64.exe

  clean:     ## Clean all created artifacts
clean:
	git clean --exclude=.idea/ -fdx

  cs-fix:        ## Fix code style for all files
cs-fix:
	gofmt -w ./

  cs-diff:        ## Display code style error files
cs-diff:
	gofmt -l ./
