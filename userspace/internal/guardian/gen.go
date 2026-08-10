package guardian

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang -cflags "-O2 -g -Wall" guardian ./../../../core/guard.skel.c
