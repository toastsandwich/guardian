# Guardian

XDP-based eBPF firewall that drops traffic from blacklisted source IPs.

## How it works

1. An **XDP program** (`core/`) attaches to a network interface.
2. For each IPv4 packet it looks up the source address in a BPF hash map (`guardian_m`).
3. Unknown sources are auto-recorded as allowed; entries marked blocked are dropped (`XDP_DROP`).
4. A **Go userspace daemon** (`userspace/`) loads the program, pins the XDP link, and exposes a small CLI over a Unix socket (`/tmp/guardian.sock`).

## Layout

```
core/           BPF C source + Makefile (clang → guardian.bpf.o)
userspace/      Go CLI + daemon (cilium/ebpf, cobra)
  cmd/          attach | start | ping | stop
  internal/
    daemon/     Unix socket server/client
    guardian/   eBPF load, attach, pin helpers (bpf2go)
```

## Build

**Kernel program** (needs `clang`, `llvm-strip`, `bpftool`, kernel headers / `vmlinux.h`):

```bash
cd core && make
```

**Userspace binary**:

```bash
cd userspace && make
# → ./guardian
```

Regenerate Go bindings after changing the BPF source:

```bash
cd userspace/internal/guardian
go generate
```

## Usage

Root is required to attach XDP and manage BPF maps.

```bash
# Start the daemon (listens on /tmp/guardian.sock)
sudo ./guardian start

# Attach XDP to an interface
sudo ./guardian attach --to eth0

# Health check
./guardian ping

# Stop the daemon
./guardian stop
```

## Requirements

- Linux kernel with XDP and BTF/CO-RE support
- Go 1.22+ (module targets 1.26)
- `clang` / LLVM for BPF compilation
- `cilium/ebpf` for loading and attaching programs

## Status

Early prototype: attach/detach and basic allow/drop map logic are in place. Map population for blacklisting and production lifecycle (systemd unit, unpin on stop, etc.) are still evolving.
