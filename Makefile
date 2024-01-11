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
COMPILER_SRC_DIR = $(WORK_PATH)/compiler
EXAMPLE_DIR = $(COMPILER_SRC_DIR)/examples
TEST_DIR = $(COMPILER_SRC_DIR)/tests
TEST_FILE = $(EXAMPLE_DIR)/main.$(EXT_NAME)
BIN_PATH = $(GOPATH)/bin/$(BIN_FILE)

.PHONY: find_test_files
find_test_files:
    @files=`find $(TEST_DIR) -type f -name "*.sim"`; \
    for file in $$files ; do \
        echo `realpath $$file` ; \
    done

.PHONY: lex
lex: $(COMPILER_SRC_DIR)/lex.go $(TEST_FILE)
	@go run -tags lex $(COMPILER_SRC_DIR) $(TEST_FILE)

.PHONY: parse
parse: $(COMPILER_SRC_DIR)/parse.go $(TEST_FILE)
	@go run -tags parse $(COMPILER_SRC_DIR) $(TEST_FILE)

.PHONY: analyse
analyse: $(COMPILER_SRC_DIR)/analyse.go $(TEST_FILE)
	@go run -tags analyse $(COMPILER_SRC_DIR) $(TEST_FILE)

.PHONY: codegenir
codegenir: $(COMPILER_SRC_DIR)/codegen_ir.go $(TEST_FILE)
	@go run -tags codegenir $(COMPILER_SRC_DIR) $(TEST_FILE)

.PHONY: codegenllvm
codegenllvm: $(COMPILER_SRC_DIR)/codegen_llvm.go $(TEST_FILE)
	@go run -tags codegenllvm $(COMPILER_SRC_DIR) $(TEST_FILE)

.PHONY: codegenasm
codegenasm: $(COMPILER_SRC_DIR)/codegen_llvm.go $(TEST_FILE)
	@go run -tags codegenasm $(COMPILER_SRC_DIR) $(TEST_FILE)

.PHONY: run
run: $(COMPILER_SRC_DIR)/main.go $(TEST_FILE)
	@go run $(COMPILER_SRC_DIR) $(TEST_FILE)

.PHONY: build
build: clean $(COMPILER_SRC_DIR)/main.go
	go build -o $(BIN_FILE) $(COMPILER_SRC_DIR)
	ln -s $(WORK_PATH)/$(BIN_FILE) $(BIN_PATH)

.PHONY: clean
clean:
	rm -f $(WORK_PATH)/$(BIN_FILE)
	rm -f $(BIN_PATH)

.PHONY: test
test:
	@make build > /dev/null
	@files=`find $(TEST_DIR) -type f -name "*.$(EXT_NAME)"`; \
	for file in $$files; do \
	  	name=`basename $$file .$(EXT_NAME)`; \
		(./$(BIN_FILE) $$file > /dev/null 2>&1 || (echo -e "\e[33m 测试失败 $$name \e[0m"; exit 1)) && echo -e "\e[32m 测试成功 $$name \e[0m"; \
    done; \
    make clean > /dev/null