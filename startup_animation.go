package main

import (
	"context"
	"time"
)

const startupAnimationDuration = 5 * time.Second

// Reference palette: top first, clockwise, at full saturation/value.
// Right-side violets use less red after live feedback: the original mix
// looked too pale/pink on the device. Blue stays at maximum intensity.
var startupPalette = [...]string{
	"#FF0085", "#9000FF", "#8800FF", "#9800FF", "#C000FF", "#FF00BC",
	"#FF009A", "#FF0055", "#FF0002", "#FF4A00", "#FF8700", "#FF4E00",
}

type startupAnimation struct {
	running  bool
	deviceIP string
	command  LEDCommand
	cancel   context.CancelFunc
}

func startupLight() LEDCommand {
	colors := make([]LEDColor, len(startupPalette))
	for i, hex := range startupPalette {
		brightness := 100
		if i == 0 || i == 7 { // Top and lower-left pink: balance overpowering LEDs.
			brightness = 70
		}
		colors[i] = ColorFromHex(hex, brightness)
	}
	return LEDCommand{Command: "set_leds_individual", LEDValues: LEDValues{Colors: colors}}
}

func startupDeviceKey(mac, deviceIP string) string {
	if mac != "" {
		return "mac:" + mac
	}
	return "ip:" + deviceIP
}

// The device API has no boot event. Run once at the first reachable poll per
// device in this server process, never on ordinary refreshes or browser visits.
func (a *App) startupCommand(mac, deviceIP string) (LEDCommand, bool) {
	if !a.cfg.StartupAnimation || a.cfg.DeviceController == nil || !isPushableDeviceIP(deviceIP) {
		return LEDCommand{}, false
	}
	key := startupDeviceKey(mac, deviceIP)
	a.mu.Lock()
	if mac != "" {
		a.deviceIPs[mac] = deviceIP
	}
	if a.startupAnimations == nil {
		a.startupAnimations = make(map[string]*startupAnimation)
	}
	if animation, seen := a.startupAnimations[key]; seen {
		command, running := animation.command, animation.running
		a.mu.Unlock()
		return command, running
	}
	ctx, cancel := context.WithCancel(context.Background())
	animation := &startupAnimation{running: true, deviceIP: deviceIP, command: startupLight(), cancel: cancel}
	a.startupAnimations[key] = animation
	command := animation.command
	a.mu.Unlock()
	go a.runStartupAnimation(ctx, animation, mac, deviceIP)
	return command, true
}

func (a *App) runStartupAnimation(ctx context.Context, animation *startupAnimation, mac, deviceIP string) {
	accepted := false
	defer animation.cancel()
	defer func() {
		a.mu.Lock()
		animation.running = false
		a.mu.Unlock()
		if !accepted && ctx.Err() == nil {
			a.cfg.Logger.Printf("PuKK startup LED sequence skipped: local API unavailable at first poll; no later restart (mac=%s ip=%s)", mac, deviceIP)
		}
		if accepted {
			if ctx.Err() != nil {
				a.cfg.Logger.Printf("PuKK startup LED sequence interrupted by device interaction: mac=%s ip=%s", mac, deviceIP)
			} else {
				a.cfg.Logger.Printf("PuKK startup LED sequence ended: mac=%s ip=%s", mac, deviceIP)
			}
		}
		// Interaction owns the ring after cancellation. Do not enqueue an
		// ambient snapshot that can overwrite its selection/NFC feedback.
		if ctx.Err() == nil {
			a.restoreStartupRing(ctx, deviceIP)
		}
	}()
	// The device can poll before its local REST API is fully ready. Begin the
	// visible five seconds only after it has accepted the first palette frame.
	readyCtx, readyCancel := context.WithTimeout(ctx, 5*time.Second)
	defer readyCancel()
	for {
		pushCtx, cancel := context.WithTimeout(readyCtx, a.cfg.DevicePushTimeout)
		err := a.pushDeviceLEDValues(pushCtx, deviceIP, startupLight().LEDValues, true)
		cancel()
		if err == nil {
			accepted = true
			break
		}
		if ctx.Err() != nil {
			return
		}
		a.cfg.Logger.Printf("PuKK startup LED sequence to %s waiting for local API: %v", deviceIP, err)
		if readyCtx.Err() != nil {
			return
		}
		timer := time.NewTimer(100 * time.Millisecond)
		select {
		case <-readyCtx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
	a.cfg.Logger.Printf("PuKK startup LED sequence started: mac=%s ip=%s duration=5s steady saturated colors brightness=100%% (top/lower-left pink=70%%)", mac, deviceIP)
	timer := time.NewTimer(startupAnimationDuration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

// Restore only the display, without committing pending bookings or queuing an
// unbound background write. Cancellation remains effective until the push.
func (a *App) restoreStartupRing(ctx context.Context, deviceIP string) {
	now := a.cfg.Now().UTC()
	a.mu.Lock()
	active := a.currentBookingLocked(now)
	exactBusy := append([]TimeRange(nil), a.exactBusy...)
	exactKnown := a.exactBusyKnown
	var provisional []TimeRange
	if a.pending != nil {
		if r := a.pending.Range(); !r.Empty() {
			provisional = []TimeRange{r}
		}
	}
	a.mu.Unlock()
	command := a.renderRing(now, active, exactBusy, exactKnown, provisional)
	pushCtx, cancel := context.WithTimeout(ctx, a.cfg.DevicePushTimeout)
	defer cancel()
	if err := a.pushLEDValues(pushCtx, deviceIP, command.LEDValues); err != nil && ctx.Err() == nil {
		a.cfg.Logger.Printf("PuKK startup ring restore to %s failed: %v", deviceIP, err)
	}
}

func (a *App) cancelStartupAnimation(mac, deviceIP string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if deviceIP == "" && mac != "" {
		deviceIP = a.deviceIPs[mac]
	}
	// A finished visible sequence can still have its restoration queued.
	if animation := a.startupAnimations[startupDeviceKey(mac, deviceIP)]; animation != nil {
		animation.running = false
		animation.cancel()
	}
}
