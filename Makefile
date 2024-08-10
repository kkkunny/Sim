UNAME_S := $(shell uname -s)

OS = WIN
ifeq ($(UNAME_S),Linux)
    OS = LINUX
endif
ifeq ($(UNAME_S),Darwin)
    OS = MAC
endif

ifeq ($(OS),WIN)
    BIN_FILE = sim.exe
endif
ifeq ($(OS),LINUX)
    BIN_FILE = sim
endif

EXT_NAME = sim
WORK_PATH = $(shell pwd)
EXAMPLE_DIR = $(WORK_PATH)/examples
TEST_DIR = $(WORK_PATH)/tests
TEST_FILE = $(EXAMPLE_DIR)/main.$(EXT_NAME)
BIN_PATH = $(GOPATH)/bin/$(BIN_FILE)

.PHONY: find_test_files
find_test_files:
    @files=`find $(TEST_DIR) -type f -name "*.sim"`; \
    for file in $$files ; do \
        echo `realpath $$file` ; \
    done

.PHONY: lex
lex: $(WORK_PATH)/lex.go $(TEST_FILE)
	@go run -tags lex $(WORK_PATH) $(TEST_FILE)

.PHONY: parse
parse: $(WORK_PATH)/parse.go $(TEST_FILE)
	@go run -tags parse $(WORK_PATH) $(TEST_FILE)

.PHONY: analyse
analyse: $(WORK_PATH)/analyse.go $(TEST_FILE)
	@go run -tags analyse $(WORK_PATH) $(TEST_FILE)

.PHONY: codegenir
codegenir: $(WORK_PATH)/codegen_ir.go $(TEST_FILE)
	@go run -tags codegenir $(WORK_PATH) $(TEST_FILE)

.PHONY: codegenasm
codegenasm: $(WORK_PATH)/codegen_asm.go $(TEST_FILE)
	@go run -tags codegenasm $(WORK_PATH) $(TEST_FILE)

.PHONY: run
run: $(WORK_PATH)/main.go $(TEST_FILE)
	@go run $(WORK_PATH) $(TEST_FILE)

.PHONY: build
build: clean $(WORK_PATH)/main.go
	go build -o $(BIN_FILE) $(WORK_PATH)
	ln -s $(WORK_PATH)/$(BIN_FILE) $(BIN_PATH)

.PHONY: clean
clean:
	rm -f $(WORK_PATH)/$(BIN_FILE)
	rm -f $(BIN_PATH)

.PHONY: test
test:
	@make build > /dev/null
	@okfiles=`find $(TEST_DIR)/success -type f -name "*.$(EXT_NAME)"`; \
		for file in $$okfiles; do \
			name=`basename $$file .$(EXT_NAME)`; \
			(./$(BIN_FILE) $$file > /dev/null 2>&1 || (echo -e "\e[33m 测试失败 $$name \e[0m"; exit 1)) && echo -e "\e[32m 测试成功 $$name \e[0m"; \
		done;
	@errfiles=`find $(TEST_DIR)/failed -type f -name "*.$(EXT_NAME)"`; \
		for file in $$errfiles; do \
			name=`basename $$file .$(EXT_NAME)`; \
			(./$(BIN_FILE) $$file > /dev/null 2>&1 || (echo -e "\e[32m 测试成功 $$name \e[0m"; exit 1)) && echo -e "\e[33m 测试失败 $$name \e[0m"; \
		done; \
		make clean > /dev/null