# Guardian

Guardian is an XDP-based eBPF program that attaches to a network interface. It is still under development.

The current prototype records source IPs and treats unknown addresses according to the mode:

- **monk** — allow by default, drop IPs added with `deny`
- **sentry** — deny by default, pass IPs added with `allow`

A userspace daemon loads the XDP program and is controlled through a CLI over a Unix socket (`/tmp/guardian.sock`).

## Build

**Kernel program** (needs `clang`, `llvm-strip`, `bpftool`, kernel headers / `vmlinux.h`):

```bash
cd core && make
```

**Userspace binary**:

```bash
cd userspace && make
# → ./cli
```

Root is required to attach XDP and manage BPF maps.

## Commands

Start the daemon first. It listens on `/tmp/guardian.sock` and must keep running while you use the other commands.

### `start`

Start the guardian daemon.

```bash
sudo ./cli start
```

### `attach`

Attach guardian to a network interface and start filtering.

```bash
sudo ./cli attach --to eth0 --mode monk
sudo ./cli attach -T eth0 -M sentry
```

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--to` | `-T` | yes | interface name (for example `eth0`) |
| `--mode` | `-M` | yes | `monk` or `sentry` |

### `detach`

Detach guardian from the interface. The daemon keeps running.

```bash
sudo ./cli detach
```

### `allow`

Allow an IPv4 address to send traffic.

```bash
sudo ./cli allow 192.168.1.10
```

### `deny`

Deny an IPv4 address from sending traffic.

```bash
sudo ./cli deny 192.168.1.10
```

### `list`

List stored IPs and whether each is allowed or denied. Alias: `ls`.

```bash
sudo ./cli list
sudo ./cli ls
```

### `mode`

Get or set the filtering mode.

```bash
sudo ./cli mode get
sudo ./cli mode set monk
sudo ./cli mode set sentry
```

### `stop`

Detach from the interface (if attached) and stop the daemon.

```bash
sudo ./cli stop
```

## Example

```bash
sudo ./cli start

sudo ./cli attach --to eth0 --mode monk
sudo ./cli deny 203.0.113.10
sudo ./cli list
sudo ./cli mode set sentry
sudo ./cli allow 192.168.1.20
sudo ./cli mode get
sudo ./cli stop
```

## Layout

```
core/           BPF C source + Makefile (clang → guardian.bpf.o)
userspace/      Go CLI + daemon (cilium/ebpf, cobra)
  cmd/          start, attach, detach, allow, deny, list, mode, stop
  internal/
    daemon/     Unix socket server/client
    guardian/   eBPF load, attach, pin helpers (bpf2go)
```

Regenerate Go bindings after changing the BPF source:

```bash
cd userspace/internal/guardian
go generate
```

## Requirements

- Linux kernel with XDP and BTF/CO-RE support
- Go 1.22+ (module targets 1.26)
- `clang` / LLVM for BPF compilation
- `cilium/ebpf` for loading and attaching programs
