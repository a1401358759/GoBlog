GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BINARY=goblog

all: check

run:
	$(GOCMD) run main.go

check:
	$(GOCMD) fmt ./
	$(GOCMD) vet ./

clean:
	$(GOCLEAN)
	@if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

deps:
	$(GOGET) -u github.com/swaggo/swag/cmd/swag
	$(GOGET) -u github.com/cosmtrek/air

help:
	@echo "make - 初始化数据 启动项目"
	@echo "make run - 直接启动项目"
	@echo "make check - 校验语法错误及格式整理"
	@echo "make clean - 移除二进制文件"
	@echo "make deps - 依赖的包"
