.DEFAULT_GOAL=all

PACKAGES_WITH_TESTS:=$(shell go list -f="{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}" ./... | grep -v '/vendor/')
TEST_TARGETS:=$(foreach p,${PACKAGES_WITH_TESTS},test-$(p))
TEST_OUT_DIR:=testout

.PHONY: all
all: static mock proto easyjson mod lint testall

.PHONY: mod
mod:
	rm -rf vendor
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod download
	GO111MODULE=on go mod vendor

.PHONY: static
static:
	$(eval $@_target := github.com/dialogs/dialog-go-lib/db/migrations/test)
	rm -f $($@_target)/static.go

	docker run -it --rm \
	-v "$(shell pwd):/go/src/github.com/dialogs/dialog-go-lib" \
	-w "/go/src/github.com/dialogs/dialog-go-lib" \
	go-tools-embedded:1.0.0 \
	go generate $($@_target)

.PHONY: mock
mock:
	$(eval $@_source := kafka)
	$(eval $@_target := ${$@_source}/mocks)

	rm -f $($@_target)/IReader.go
	rm -f $($@_target)/IWriter.go

	docker run -it --rm \
	-v "$(shell pwd):/go/src/github.com/dialogs/dialog-go-lib" \
	-w "/go/src/github.com/dialogs/dialog-go-lib" \
	go-tools-mock:1.0.0 \
	mockery -name=IReader -dir=${$@_source} -recursive=false -output=$($@_target) && \
	mockery -name=IWriter -dir=${$@_source} -recursive=false -output=$($@_target)

.PHONY: easyjson
easyjson:
	docker run -it --rm \
	-v "$(shell pwd):/go/src/github.com/dialogs/dialog-go-lib" \
	-w "/go/src/github.com/dialogs/dialog-go-lib/" \
	go-tools-easyjson:1.0.0 \
	rm -rfv kafka/registry/*_easyjson.go && \
	easyjson -all kafka/registry/request.go && \
	easyjson -all kafka/registry/response.go

.PHONY: proto
proto:
	$(eval $@_source := service/test)
	$(eval $@_target := service/test)

	rm -f ${$@_target}/*.pb.go

	docker run -it --rm \
	-v "$(shell pwd):/go/src/github.com/dialogs/dialog-go-lib" \
	-w "/go/src/github.com/dialogs/dialog-go-lib" \
	go-tools-protoc:1.0.0 \
	protoc \
	-I=${$@_source} \
	-I=vendor \
	--gogofaster_out=plugins=grpc,\
	Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types,\
	:${$@_target} \
	${$@_source}/*.proto

.PHONY: lint
lint:
	docker run -it --rm \
	-v "$(shell pwd):/go/src/github.com/dialogs/dialog-go-lib" \
	-w "/go/src/github.com/dialogs/dialog-go-lib" \
	go-tools-linter:1.0.0 \
	golangci-lint run ./... --exclude "is deprecated"

.PHONY: testall
testall:
	rm -rf ${TEST_OUT_DIR}
	mkdir -p -m 755 $(TEST_OUT_DIR)
	$(MAKE) -j 5 $(TEST_TARGETS)
	@echo "=== tests: ok ==="

.PHONY: $(TEST_TARGETS)
$(TEST_TARGETS):
	$(eval $@_package := $(subst test-,,$@))
	$(eval $@_filename := $(subst /,_,$($@_package)))

	@echo "== test directory $($@_package) =="
	@GO111MODULE=on go test $($@_package) -v -race \
    -coverprofile $(TEST_OUT_DIR)/$($@_filename)_cover.out \
    >> $(TEST_OUT_DIR)/$($@_filename).out \
   || ( echo 'fail $($@_package)' && cat $(TEST_OUT_DIR)/$($@_filename).out; exit 1);
