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

type ShowResult struct {
	IP      string
	Allowed bool
}

func (g *Guardian) show(ch chan ShowResult) {
	var ip uint32
	var allowed bool
	iter := g.guardianMap.Iterate()
	for iter.Next(&ip, &allowed) {
		ipbuf := make([]byte, 4)
		binary.BigEndian.PutUint32(ipbuf, ip)

		ch <- ShowResult{
			IP:      net.IP(ipbuf).String(),
			Allowed: allowed,
		}
	}
	close(ch)
}

func (g *Guardian) Show() chan ShowResult {
	ch := make(chan ShowResult)
	go g.show(ch)
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
	return g.lnk.Detach()
}

func (g *Guardian) Close() error {
	if !g.isClosed.Load() {
		g.isClosed.Store(true)
		err := g.lnk.Unpin()
		if err != nil {
			return err
		}
		err = g.Dettach()
		if err != nil {
			return err
		}
		err = g.lnk.Close()
		if err != nil {
			return err
		}
		err = g.guardianMap.Close()
		if err != nil {
			return err
		}
		err = g.obj.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
