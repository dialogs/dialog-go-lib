.DEFAULT_GOAL=all

PACKAGES_WITH_TESTS:=$(shell go list -f="{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}" ./... | grep -v '/vendor/')
TEST_TARGETS:=$(foreach p,${PACKAGES_WITH_TESTS},test-$(p))
TEST_OUT_DIR:=testout
PROJECT:=github.com/dialogs/dialog-go-lib

.PHONY: all
all: static mock proto mod easyjson mod lint testall

.PHONY: mod
mod:
	rm -rf vendor
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod download

.PHONY: static
static:
	docker run -it --rm \
	-v "$(shell pwd):/go/src/${PROJECT}" \
	-v "${GOPATH}/pkg:/go/pkg" \
	-w "/go/src/${PROJECT}" \
	-e "GOFLAGS=" \
	go-tools-embedded:latest \
	sh -c '\
	rm -fv ${PROJECT}/db/migrations/test/static.go && \
	go generate ${PROJECT}/db/migrations/test'

.PHONY: mock
mock:
	$(eval $@_source := kafka)
	$(eval $@_target := ${$@_source}/mocks)

	rm -f $($@_target)/IReader.go
	rm -f $($@_target)/IWriter.go

	docker run -it --rm \
	-v "$(shell pwd):/go/src/${PROJECT}" \
	-v "${GOPATH}/pkg:/go/pkg" \
	-w "/go/src/${PROJECT}" \
	-e "GOFLAGS=" \
	go-tools-mock:latest \
	sh -c 'mockery -name=IReader -dir=${$@_source} -recursive=false -output=$($@_target) && \
	mockery -name=IWriter -dir=${$@_source} -recursive=false -output=$($@_target)'

.PHONY: easyjson
easyjson:
	docker run -it --rm \
	-v "$(shell pwd):/go/src/${PROJECT}" \
	-v "${GOPATH}/pkg:/go/pkg" \
	-w "/go/src/${PROJECT}" \
	-e "GOFLAGS=" \
	go-tools-easyjson:latest \
	sh -c 'rm -rfv kafka/schemaregistry/*_easyjson.go && \
	easyjson -all kafka/schemaregistry/request.go && \
	easyjson -all kafka/schemaregistry/response.go'

.PHONY: proto
proto:
	$(eval $@_source := service/test)
	$(eval $@_target := service/test)

	rm -f ${$@_target}/*.pb.go

	docker run -it --rm \
	-v "$(shell pwd):/go/src/${PROJECT}" \
	-v "${GOPATH}/pkg:/go/pkg" \
	-w "/go/src/${PROJECT}" \
	-e "GOFLAGS=" \
	go-tools-protoc:latest \
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
	-v "$(shell pwd):/go/src/${PROJECT}" \
	-v "${GOPATH}/pkg:/go/pkg" \
	-w "/go/src/${PROJECT}" \
	-e "GOFLAGS=" \
	go-tools-linter:latest \
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
