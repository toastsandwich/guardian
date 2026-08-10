#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#define ETH_P_IP 0x0800

// parser_t will contain the start pointer to the packet
struct parser_t {
  void *offset;
};

// parses ethhdr
static __always_inline int parse_ethhdr(struct parser_t *parser, void *data_end,
                                        struct ethhdr **ptr) {
  struct ethhdr *eth      = parser->offset;
  __u32          hdr_size = sizeof(struct ethhdr);

  if ((void *)(eth + 1) > data_end) {
    return -1;
  }

  parser->offset += hdr_size;
  *ptr = eth;

  return (*ptr)->h_proto;
}

// parsers iphdr
static __always_inline int parse_iphdr(struct parser_t *parser, void *data_end,
                                       struct iphdr **ptr) {
  struct iphdr *iph      = parser->offset;
  __u32         hdr_size = sizeof(struct iphdr);

  if ((void *)(iph + 1) > data_end) {
    return -1;
  }

  parser->offset += hdr_size;
  *ptr = iph;

  return (*ptr)->tot_len;
}

// map to contain blacklisted ips
struct {
  __uint(type, BPF_MAP_TYPE_HASH);
  __type(key, __be32);
  __type(value, __u8);
  __uint(max_entries, 256);
} guardian_m SEC(".maps");

SEC("xdp")
int guardian(struct xdp_md *ctx) {
  void *data;
  void *data_end;

  struct parser_t parser;
  struct ethhdr  *eth;
  struct iphdr   *iph;

  int eth_p;
  int action;

  static __u8 allowed;

  action        = XDP_PASS;
  allowed       = 1;
  data          = (void *)(long)(ctx->data);
  data_end      = (void *)(long)(ctx->data_end);
  parser.offset = data;

  eth_p = parse_ethhdr(&parser, data_end, &eth);
  if (eth_p < 0) {
    goto end;
  }
  if (bpf_ntohs(eth_p) == ETH_P_IP) {
    int len;
    len = parse_iphdr(&parser, data_end, &iph);
    if (len < 0) {
      goto end;
    }

    __be32 saddr;
    __u8   ok;
    void  *value;
    saddr = iph->saddr;
    value = bpf_map_lookup_elem(&guardian_m, &saddr);
    if (!value) {
      bpf_map_update_elem(&guardian_m, &saddr, &allowed, BPF_NOEXIST);
      goto end;
    }
    ok = *(__u8 *)(value);
    if (!ok) {
      bpf_printk("dropping packet because %pI4 was blacklisted",
                 bpf_ntohl(saddr));
      action = XDP_DROP;
      goto end;
    }
  }

end:
  return action;
}

char LICENSE[] SEC("license") = "Dual BSD/GPL";
