//go:build windows && amd64

package vpad

import (
	"testing"
	"unsafe"

	"xpadlink/proto"
)

func TestXUSBReportLayout(t *testing.T) {
	if got := unsafe.Sizeof(xusbReport{}); got != 12 {
		t.Fatalf("XUSB report is %d bytes, want 12", got)
	}
}

func TestReportFromState(t *testing.T) {
	state := proto.State{
		Buttons: proto.BtnA | proto.BtnB | proto.BtnX | proto.BtnY |
			proto.BtnLB | proto.BtnRB | proto.BtnBack | proto.BtnStart |
			proto.BtnGuide | proto.BtnL3 | proto.BtnR3 |
			proto.BtnDUp | proto.BtnDDown | proto.BtnDLeft | proto.BtnDRight,
		LX: 123, LY: -456, RX: -789, RY: 321, LT: 42, RT: 255,
	}
	report := reportFromState(state)
	if report.buttons != 0xf7ff {
		t.Errorf("buttons = %#04x, want %#04x", report.buttons, 0xf7ff)
	}
	if report.leftTrigger != 42 || report.rightTrigger != 255 {
		t.Errorf("triggers = %d, %d; want 42, 255", report.leftTrigger, report.rightTrigger)
	}
	if report.leftX != 123 || report.leftY != 456 || report.rightX != -789 || report.rightY != -321 {
		t.Errorf("sticks = (%d, %d, %d, %d); want (123, 456, -789, -321)",
			report.leftX, report.leftY, report.rightX, report.rightY)
	}
	if got := invertAxis(-32768); got != 32767 {
		t.Errorf("invertAxis(-32768) = %d, want 32767", got)
	}
}
