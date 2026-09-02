package guardian

import (
	"encoding/binary"
	"fmt"
	"iter"
	"net"
	"strings"
	"sync/atomic"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
)

const (
	pinPath = "/sys/fs/bpf/guardian"
	monk    = "monk"
	sentry  = "sentry"
)

var (
	allow byte = 1
	deny  byte = 0

	modeKey byte = 1
)

type Mode byte

func (m Mode) toByte() byte {
	return byte(m)
}

func (m Mode) String() string {
	switch m {
	case 1:
		return monk
	case 0:
		return sentry
	default:
		return "unkown"
	}
}

func DetermineMode(mode string) (Mode, error) {
	switch strings.ToLower(mode) {
	case monk:
		return 1, nil
	case sentry:
		return 0, nil
	default:
		return 255, fmt.Errorf("unkown mode")
	}
}

type Guardian struct {
	IfaceName string

	obj         *guardianObjects
	guardianMap *ebpf.Map
	modeMap     *ebpf.Map
	lnk         link.Link

	modeKey byte

	isClosed atomic.Bool
}

func New(ifaceName string) *Guardian {
	return &Guardian{
		IfaceName: ifaceName,
	}
}

func (g *Guardian) Allow(ip string) error {
	b := net.ParseIP(ip)
	be := binary.LittleEndian.Uint32(b.To4())
	err := g.guardianMap.Update(&be, &allow, ebpf.UpdateAny)
	if err != nil {
		return err
	}
	return nil
}

func (g *Guardian) Deny(ip string) error {
	b := net.ParseIP(ip)
	be := binary.LittleEndian.Uint32(b.To4())
	err := g.guardianMap.Update(&be, &deny, ebpf.UpdateAny)
	if err != nil {
		return err
	}
	return nil
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

func (g *Guardian) List() iter.Seq[ListResult] {
	ch := make(chan ListResult)
	go g.list(ch)

	return func(yield func(ListResult) bool) {
		for lr := range ch {
			if !yield(lr) {
				break
			}
		}
	}
}

func (g *Guardian) GetMode() (string, error) {
	var m byte
	err := g.modeMap.Lookup(&g.modeKey, &m)
	if err != nil {
		return "", err
	}
	return Mode(m).String(), nil
}

func (g *Guardian) SetMode(mode Mode) error {
	m := mode.toByte()
	err := g.modeMap.Update(&g.modeKey, &m, ebpf.UpdateExist)
	if err != nil {
		return err
	}
	return nil
}

func (g *Guardian) Attach(mode Mode) error {
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
	g.modeMap = obj.guardianMaps.ModeMap
	iface, err := net.InterfaceByName(g.IfaceName)
	if err != nil {
		return err
	}

	modeKey := byte(0)
	if err := obj.guardianVariables.MODE.Get(&modeKey); err != nil {
		return err
	}

	g.modeKey = modeKey
	modeVal := mode.toByte()
	err = g.modeMap.Put(&modeKey, &modeVal)
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
	if g.modeMap != nil {
		if err := g.modeMap.Close(); err != nil {
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
