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

ifeq ($(OS),WIN)
    STATIC_LIB_FILE = sim.lib
endif
ifeq ($(OS),LINUX)
    STATIC_LIB_FILE = sim.a
endif

EXT_NAME = sim
WORK_PATH = $(shell pwd)
RUNTIME_PATH = $(WORK_PATH)/runtime/build
OUTPUT_PATH = $(WORK_PATH)/output
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
lex: $(WORK_PATH)/ $(TEST_FILE)
	@go run -tags lex $(WORK_PATH) $(TEST_FILE)

.PHONY: parse
parse: $(WORK_PATH)/ $(TEST_FILE)
	@go run -tags parse $(WORK_PATH) $(TEST_FILE)

.PHONY: analyse
analyse: $(WORK_PATH)/ $(TEST_FILE)
	@go run -tags analyse $(WORK_PATH) $(TEST_FILE)

.PHONY: codegenir
codegenir: $(WORK_PATH)/ $(TEST_FILE)
	@go run -tags codegenir $(WORK_PATH) $(TEST_FILE)

.PHONY: codegenasm
codegenasm: $(WORK_PATH)/ $(TEST_FILE)
	@go run -tags codegenasm $(WORK_PATH) $(TEST_FILE)

.PHONY: run
run: $(WORK_PATH)/ $(TEST_FILE)
	@go run $(WORK_PATH) $(TEST_FILE)

.PHONY: build
build: clean $(WORK_PATH)/
	go build -buildmode=c-archive -o $(OUTPUT_PATH)/$(STATIC_LIB_FILE) $(RUNTIME_PATH)
	go build -o $(OUTPUT_PATH)/$(BIN_FILE) $(WORK_PATH)/cmd
	ln -s $(OUTPUT_PATH)/$(BIN_FILE) $(BIN_PATH)

.PHONY: clean
clean:
	rm -rf $(OUTPUT_PATH)
	rm -f $(BIN_PATH)

.PHONY: test
test:
	@make build > /dev/null
	@okfiles=`find $(TEST_DIR)/success -type f -name "*.$(EXT_NAME)"`; \
		for file in $$okfiles; do \
			name=`basename $$file .$(EXT_NAME)`; \
			($(OUTPUT_PATH)/$(BIN_FILE) run $$file > /dev/null 2>&1 || (echo -e "\e[33m 测试失败 $$name \e[0m"; exit 1)) && echo -e "\e[32m 测试成功 $$name \e[0m"; \
		done;
	@errfiles=`find $(TEST_DIR)/failed -type f -name "*.$(EXT_NAME)"`; \
		for file in $$errfiles; do \
			name=`basename $$file .$(EXT_NAME)`; \
			($(OUTPUT_PATH)/$(BIN_FILE) run $$file > /dev/null 2>&1 || (echo -e "\e[32m 测试成功 $$name \e[0m"; exit 1)) && echo -e "\e[33m 测试失败 $$name \e[0m"; \
		done; \
		make clean > /dev/null