package guardian

import (
	"net"
	"sync/atomic"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
)

const pinPath = "/sys/fs/bpf/guardian"

type be32 = uint32

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
