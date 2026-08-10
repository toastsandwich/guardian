# Makefile for guardian eBPF (XDP) core
#
# Usage:
#   make          - build BPF object
#   make clean    - remove build artifacts
#   make verify   - check object has valid BTF (no root needed)
#   make load     - pin XDP prog to /sys/fs/bpf/guardian (needs root)
#   make unload   - remove pin (needs root)

CLANG      ?= clang
LLVM_STRIP ?= llvm-strip
BPFTOOL    ?= bpftool
PIN_PATH   ?= /sys/fs/bpf/guardian

# Map uname -m → __TARGET_ARCH_* used by bpf_helpers / CO-RE
ARCH := $(shell uname -m | sed \
	-e 's/x86_64/x86/' \
	-e 's/aarch64/arm64/' \
	-e 's/ppc64le/powerpc/' \
	-e 's/mips.*/mips/' \
	-e 's/riscv64/riscv/' \
	-e 's/loongarch64/loongarch/')

SRC := guard.skel.c
OBJ := guardian.bpf.o

# System header search paths (needed when compiling with -target bpf)
CLANG_BPF_SYS_INCLUDES := $(shell $(CLANG) -v -E - </dev/null 2>&1 \
	| sed -n '/<...> search starts here:/,/End of search list./{ s| \(/.*\)|-idirafter \1|p }')

INCLUDES := -I. -I/usr/include

BPF_CFLAGS := \
	-O2 -g -Wall \
	-target bpf \
	-D__TARGET_ARCH_$(ARCH) \
	$(INCLUDES) \
	$(CLANG_BPF_SYS_INCLUDES)

.PHONY: all clean verify load unload

all: $(OBJ)

$(OBJ): $(SRC) vmlinux.h
	$(CLANG) $(BPF_CFLAGS) -c $(SRC) -o $@
	$(LLVM_STRIP) -g $@

# Read-only checks — do not write ELF anywhere (writing to /dev/null fails).
verify: $(OBJ)
	@file $(OBJ) | grep -q 'eBPF' || (echo "not an eBPF object" && exit 1)
	$(BPFTOOL) btf dump file $(OBJ) format raw >/dev/null
	@echo "OK: $(OBJ) is a valid BPF object with BTF"
	@$(BPFTOOL) btf dump file $(OBJ) format c 2>/dev/null | head -5 || true

# Requires root and a mounted bpffs at /sys/fs/bpf
load: $(OBJ)
	$(BPFTOOL) prog load $(OBJ) $(PIN_PATH) type xdp
	@echo "pinned at $(PIN_PATH)"

unload:
	rm -f $(PIN_PATH)

clean:
	rm -f $(OBJ)
