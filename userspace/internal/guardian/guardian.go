package guardian

import (
	"encoding/binary"
	"net"
	"sync/atomic"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
)

const pinPath = "/sys/fs/bpf/guardian"

type Guardian struct {
	IfaceName string

	obj         *guardianObjects
	guardianMap *ebpf.Map
	lnk         link.Link

	isClosed atomic.Bool
}

func New(ifaceName string) *Guardian {
	return &Guardian{
		IfaceName: ifaceName,
	}
}

type ListResult struct {
	IP      string
	Allowed bool
}

func (g *Guardian) list(ch chan ListResult) {
	var ip uint32
	var allowed bool
	iter := g.guardianMap.Iterate()
	for iter.Next(&ip, &allowed) {
		ipbuf := make([]byte, 4)
		binary.LittleEndian.PutUint32(ipbuf, ip)

		ch <- ListResult{
			IP:      net.IP(ipbuf).String(),
			Allowed: allowed,
		}
	}
	close(ch)
}

func (g *Guardian) List() chan ListResult {
	ch := make(chan ListResult)
	go g.list(ch)
	return ch
}

func (g *Guardian) Attach() error {
	err := rlimit.RemoveMemlock()
	if err != nil {
		return err
	}

	obj := guardianObjects{}
	err = loadGuardianObjects(&obj, nil)
	if err != nil {
		return err
	}
	g.obj = &obj
	g.guardianMap = obj.guardianMaps.GuardianM
	iface, err := net.InterfaceByName(g.IfaceName)
	if err != nil {
		return err
	}

	lnk, err := link.AttachXDP(link.XDPOptions{
		Interface: iface.Index,
		Program:   obj.Guardian,
	})
	if err != nil {
		return err
	}
	g.lnk = lnk
	return g.lnk.Pin(pinPath)
}

func (g *Guardian) Dettach() error {
	if g == nil || g.lnk == nil {
		return nil
	}
	return g.lnk.Detach()
}

func (g *Guardian) Close() error {
	if g == nil || g.isClosed.Load() {
		return nil
	}
	g.isClosed.Store(true)

	if g.lnk != nil {
		if err := g.lnk.Unpin(); err != nil {
			return err
		}
		if err := g.Dettach(); err != nil {
			return err
		}
		if err := g.lnk.Close(); err != nil {
			return err
		}
	}
	if g.guardianMap != nil {
		if err := g.guardianMap.Close(); err != nil {
			return err
		}
	}
	if g.obj != nil {
		if err := g.obj.Close(); err != nil {
			return err
		}
	}
	return nil
}
