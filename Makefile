EXPECT_VERSION = 15
ifneq ($(shell find /bin/ -name "llvm-config*" | grep $(EXPECT_VERSION)), )
	CONFIG=$(shell find /bin/ -maxdepth 1 -name "llvm-config*" | grep $(EXPECT_VERSION) | cut -d \/ -f 3)
else
	CONFIG = llvm-config
endif

VERSION = $(shell $(CONFIG) --version)
VERSION_MAJOR = $(firstword $(subst ., ,$(VERSION)))

CONFIG_FILE = llvm_config.go

.PHONY: check
check:
	@if [ $(VERSION_MAJOR) != $(EXPECT_VERSION) ]; then echo "[Error] need llvm$(EXPECT_VERSION)"; exit 1; fi

.PHONY: clean
clean:
	@rm -f $(CONFIG_FILE)

.PHONY: config
config: check clean
	@echo "// Automatically generated by \`$(CONFIG)\`, do not edit." >> $(CONFIG_FILE)
	@echo "" >> $(CONFIG_FILE)
	@echo "package main" >> $(CONFIG_FILE)
	@echo "" >> $(CONFIG_FILE)
	@echo "// #cgo CFLAGS: $(shell $(CONFIG) --cflags)" >> $(CONFIG_FILE)
	@echo "// #cgo CXXFLAGS: $(shell $(CONFIG) --cxxflags)" >> $(CONFIG_FILE)
	@echo "// #cgo LDFLAGS: -L$(shell $(CONFIG) --libdir) $(shell $(CONFIG) --libs)" >> $(CONFIG_FILE)
	@echo "import \"C\"" >> $(CONFIG_FILE)