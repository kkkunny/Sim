BIN_FILE = sim
EXT_NAME = sim
WORK_PATH = $(shell pwd)
TEST_DIR = $(WORK_PATH)/tests
TEST_FILE = $(TEST_DIR)/hello_world.$(EXT_NAME)
BIN_PATH = $(GOPATH)/bin/$(BIN_FILE)

.PHONY: lex
lex: lex.go $(TEST_FILE)
	@go build -tags test,lex -o .test .
	-@./.test $(TEST_FILE) || true
	@rm .test

.PHONY: parse
parse: parse.go $(TEST_FILE)
	@go build -tags test,parse -o .test .
	-@./.test $(TEST_FILE) || true
	@rm .test

.PHONY: analyse
analyse: analyse.go $(TEST_FILE)
	@go build -tags test,analyse -o .test .
	-@./.test $(TEST_FILE) || true
	@rm .test

.PHONY: codegen
codegen: codegen.go $(TEST_FILE)
	@go build -tags test,codegen,llvm14 -o .test .
	-@./.test $(TEST_FILE) || true
	@rm .test

.PHONY: build
build: clean main.go
	go build -tags llvm14 -o $(BIN_FILE) main.go
	ln -s $(WORK_PATH)/$(BIN_FILE) $(BIN_PATH)

.PHONY: clean
clean:
	rm -f $(WORK_PATH)/$(BIN_FILE)
	rm -f $(BIN_PATH)

.PHONY: test
test:
	@make build > /dev/null
	@for file in $(foreach dir, $(TEST_DIR), $(wildcard $(TEST_DIR)/*.$(EXT_NAME))); do \
	    ($(BIN_FILE) run $$file > /dev/null 2>&1 || (echo -e "\e[33m 测试失败 $$file \e[0m"; exit 1)) && echo -e "\e[32m 测试成功 $$file \e[0m"; \
    done; \
    make clean > /dev/null

.PHONY: docker
docker:
	docker build -t $(BIN_FILE):latest .