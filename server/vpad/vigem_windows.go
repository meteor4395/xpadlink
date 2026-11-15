//go:build windows && amd64

package vpad

import (
	"fmt"
	"runtime"
	"sync"
	"syscall"
	"unsafe"

	"xpadlink/proto"
)

const (
	vigemErrorNone = 0x20000000

	xusbDPadUp        = 0x0001
	xusbDPadDown      = 0x0002
	xusbDPadLeft      = 0x0004
	xusbDPadRight     = 0x0008
	xusbStart         = 0x0010
	xusbBack          = 0x0020
	xusbLeftThumb     = 0x0040
	xusbRightThumb    = 0x0080
	xusbLeftShoulder  = 0x0100
	xusbRightShoulder = 0x0200
	xusbGuide         = 0x0400
	xusbA             = 0x1000
	xusbB             = 0x2000
	xusbX             = 0x4000
	xusbY             = 0x8000
)

type xusbReport struct {
	buttons      uint16
	leftTrigger  uint8
	rightTrigger uint8
	leftX        int16
	leftY        int16
	rightX       int16
	rightY       int16
}

type vigemAPI struct {
	dll                    *syscall.DLL
	alloc                  *syscall.Proc
	free                   *syscall.Proc
	connect                *syscall.Proc
	disconnect             *syscall.Proc
	targetAlloc            *syscall.Proc
	targetFree             *syscall.Proc
	targetAdd              *syscall.Proc
	targetRemove           *syscall.Proc
	registerNotification   *syscall.Proc
	unregisterNotification *syscall.Proc
	update                 *syscall.Proc
}

type vigemPad struct {
	api    *vigemAPI
	client uintptr
	target uintptr
	notify uintptr

	rumbleMu sync.RWMutex
	rumbleFn func(strong, weak, ms uint16)
}

// NewUinput creates an Xbox 360 controller through ViGEmBus.
// ViGEmClient.dll and the ViGEmBus driver must be installed on the PC.
func NewUinput(_ string) (Pad, error) {
	api, err := loadVigem()
	if err != nil {
		return nil, err
	}

	client, _, _ := api.alloc.Call()
	if client == 0 {
		api.dll.Release()
		return nil, fmt.Errorf("ViGEm could not allocate a client")
	}
	if code, _, _ := api.connect.Call(client); code != vigemErrorNone {
		api.free.Call(client)
		api.dll.Release()
		return nil, fmt.Errorf("connect to ViGEmBus failed (0x%08x); install or repair the ViGEmBus driver", uint32(code))
	}

	target, _, _ := api.targetAlloc.Call()
	if target == 0 {
		api.disconnect.Call(client)
		api.free.Call(client)
		api.dll.Release()
		return nil, fmt.Errorf("ViGEm could not allocate an Xbox 360 target")
	}
	if code, _, _ := api.targetAdd.Call(client, target); code != vigemErrorNone {
		api.targetFree.Call(target)
		api.disconnect.Call(client)
		api.free.Call(client)
		api.dll.Release()
		return nil, fmt.Errorf("add Xbox 360 controller to ViGEmBus failed (0x%08x)", uint32(code))
	}

	p := &vigemPad{api: api, client: client, target: target}
	p.notify = syscall.NewCallback(func(_, _, large, small, _, _ uintptr) uintptr {
		p.rumbleMu.RLock()
		fn := p.rumbleFn
		p.rumbleMu.RUnlock()
		if fn != nil {
			fn(uint16(large)*257, uint16(small)*257, 0)
		}
		return 0
	})
	if code, _, _ := api.registerNotification.Call(client, target, p.notify, 0); code != vigemErrorNone {
		api.targetRemove.Call(client, target)
		api.targetFree.Call(target)
		api.disconnect.Call(client)
		api.free.Call(client)
		api.dll.Release()
		return nil, fmt.Errorf("register ViGEm rumble callback failed (0x%08x)", uint32(code))
	}
	return p, nil
}

func loadVigem() (*vigemAPI, error) {
	dll, err := syscall.LoadDLL("ViGEmClient.dll")
	if err != nil {
		return nil, fmt.Errorf("load ViGEmClient.dll: %w; install ViGEmBus and place the matching-architecture ViGEmClient.dll beside xpadlinkd.exe", err)
	}
	api := &vigemAPI{dll: dll}
	procs := []struct {
		name string
		dst  **syscall.Proc
	}{
		{"vigem_alloc", &api.alloc},
		{"vigem_free", &api.free},
		{"vigem_connect", &api.connect},
		{"vigem_disconnect", &api.disconnect},
		{"vigem_target_x360_alloc", &api.targetAlloc},
		{"vigem_target_free", &api.targetFree},
		{"vigem_target_add", &api.targetAdd},
		{"vigem_target_remove", &api.targetRemove},
		{"vigem_target_x360_register_notification", &api.registerNotification},
		{"vigem_target_x360_unregister_notification", &api.unregisterNotification},
		{"vigem_target_x360_update", &api.update},
	}
	for _, entry := range procs {
		proc, err := dll.FindProc(entry.name)
		if err != nil {
			dll.Release()
			return nil, fmt.Errorf("ViGEmClient.dll is missing %s: %w", entry.name, err)
		}
		*entry.dst = proc
	}
	return api, nil
}

func (p *vigemPad) OnRumble(fn func(strong, weak, ms uint16)) {
	p.rumbleMu.Lock()
	p.rumbleFn = fn
	p.rumbleMu.Unlock()
}

func (p *vigemPad) Update(s proto.State) error {
	report := reportFromState(s)
	// Windows amd64 passes the 12-byte XUSB_REPORT argument indirectly.
	code, _, _ := p.api.update.Call(p.client, p.target, uintptr(unsafe.Pointer(&report)))
	runtime.KeepAlive(&report)
	if code != vigemErrorNone {
		return fmt.Errorf("update ViGEm Xbox 360 controller failed (0x%08x)", uint32(code))
	}
	return nil
}

func reportFromState(s proto.State) xusbReport {
	var buttons uint16
	for _, mapping := range [...]struct {
		input  uint32
		output uint16
	}{
		{proto.BtnA, xusbA},
		{proto.BtnB, xusbB},
		{proto.BtnX, xusbX},
		{proto.BtnY, xusbY},
		{proto.BtnLB, xusbLeftShoulder},
		{proto.BtnRB, xusbRightShoulder},
		{proto.BtnBack, xusbBack},
		{proto.BtnStart, xusbStart},
		{proto.BtnGuide, xusbGuide},
		{proto.BtnL3, xusbLeftThumb},
		{proto.BtnR3, xusbRightThumb},
		{proto.BtnDUp, xusbDPadUp},
		{proto.BtnDDown, xusbDPadDown},
		{proto.BtnDLeft, xusbDPadLeft},
		{proto.BtnDRight, xusbDPadRight},
	} {
		if s.Buttons&mapping.input != 0 {
			buttons |= mapping.output
		}
	}
	return xusbReport{
		buttons:      buttons,
		leftTrigger:  s.LT,
		rightTrigger: s.RT,
		leftX:        s.LX,
		leftY:        invertAxis(s.LY),
		rightX:       s.RX,
		rightY:       invertAxis(s.RY),
	}
}

func invertAxis(value int16) int16 {
	if value == -32768 {
		return 32767
	}
	return -value
}

func (p *vigemPad) Close() error {
	p.rumbleMu.Lock()
	p.rumbleFn = nil
	p.rumbleMu.Unlock()

	p.api.unregisterNotification.Call(p.client, p.target)
	code, _, _ := p.api.targetRemove.Call(p.client, p.target)
	p.api.targetFree.Call(p.target)
	p.api.disconnect.Call(p.client)
	p.api.free.Call(p.client)
	p.api.dll.Release()
	if code != vigemErrorNone {
		return fmt.Errorf("remove ViGEm Xbox 360 controller failed (0x%08x)", uint32(code))
	}
	return nil
}
