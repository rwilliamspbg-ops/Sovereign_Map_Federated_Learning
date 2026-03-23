TOOLROOT ?= /go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.7.linux-amd64

ifneq ("$(wildcard $(TOOLROOT)/bin/go)","")
GO := GOROOT=$(TOOLROOT) GOTOOLCHAIN=local $(TOOLROOT)/bin/go
else
GO := go
endif

.PHONY: test tidy fmt

test:
	$(GO) test ./...

tidy:
	$(GO) mod tidy

fmt:
	$(GO) fmt ./...
