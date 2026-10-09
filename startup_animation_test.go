package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type startupRetryDevice struct {
	*recordingDeviceController
	mu     sync.Mutex
	failed bool
}

func (d *startupRetryDevice) SetIndividualLEDs(ctx context.Context, ip string, values LEDValues) error {
	d.mu.Lock()
	fail := !d.failed
	d.failed = true
	d.mu.Unlock()
	if fail {
		return errors.New("PuKK local API still booting")
	}
	return d.recordingDeviceController.SetIndividualLEDs(ctx, ip, values)
}

func TestStartupRetriesUntilFirstFrameAcceptedBeforeCountingFiveSeconds(t *testing.T) {
	frames := &recordingDeviceController{done: make(chan struct{}, 100)}
	device := &startupRetryDevice{recordingDeviceController: frames}
	app := NewApp(&recordingRooms{}, AppConfig{DeviceController: device, StartupAnimation: true, Logger: discardLogger{}})
	app.Poll("abc", "192.0.2.10")
	first := waitForDeviceFrames(t, frames, 1)
	if first[0].Colors[0].Hex() != "#FF0085" {
		t.Fatal("failed startup push was consumed instead of retried")
	}
	app.cancelStartupAnimation("abc", "192.0.2.10")
}

func TestServerNormalizesWindowsMappedIPv4SourceAddress(t *testing.T) {
	if got := remoteIP("[::ffff:192.0.2.10]:1234"); got != "192.0.2.10" {
		t.Fatalf("source IP = %q", got)
	}
}

func TestStartupShowsSaturatedColorsWithBalancedBrightness(t *testing.T) {
	want := []string{"#FF0085", "#9000FF", "#8800FF", "#9800FF", "#C000FF", "#FF00BC", "#FF009A", "#FF0055", "#FF0002", "#FF4A00", "#FF8700", "#FF4E00"}
	for range 3 {
		command := startupLight()
		if len(command.LEDValues.Colors) != 12 {
			t.Fatalf("colors = %d", len(command.LEDValues.Colors))
		}
		for i, color := range command.LEDValues.Colors {
			if color.Hex() != want[i] {
				t.Fatalf("LED %d = %s, want %s", i, color.Hex(), want[i])
			}
			brightness := 100
			if i == 0 || i == 7 {
				brightness = 70
			}
			if color.Brightness != brightness {
				t.Fatalf("LED %d brightness = %d, want %d", i, color.Brightness, brightness)
			}
		}
	}
}

func TestStartupAnimationRunsOnceAndRestoresCurrentBookingAfterFiveSeconds(t *testing.T) {
	device := &recordingDeviceController{done: make(chan struct{}, 100)}
	now := time.Date(2026, 10, 9, 10, 0, 0, 0, time.UTC)
	rooms := &recordingRooms{currentBooking: &Booking{ID: 7, ResourceID: 44, Start: now, End: now.Add(15 * time.Minute), CheckinConfirmed: true}}
	app := NewApp(rooms, AppConfig{ResourceID: 44, Now: func() time.Time { return now }, DeviceController: device, StartupAnimation: true, Logger: discardLogger{}})
	if err := app.RefreshActiveBooking(context.Background()); err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	first := app.Poll("abc", "192.0.2.10")
	if first.LEDValues.Colors[0].Hex() != "#FF0085" {
		t.Fatal("first poll must return reference palette")
	}
	waitForDeviceFrames(t, device, 1)
	time.Sleep(150 * time.Millisecond)
	during := app.Poll("abc", "192.0.2.10")
	if during.LEDValues.Colors[0].Hex() != "#FF0085" {
		t.Fatal("poll overwrote startup palette")
	}
	deadline := time.After(6500 * time.Millisecond)
	for {
		frames := device.snapshotFrames()
		if len(frames) > 0 && frames[len(frames)-1].Colors[0].Hex() == BusyRed {
			if time.Since(started) < 4900*time.Millisecond || time.Since(started) > 6500*time.Millisecond {
				t.Fatalf("startup duration = %s", time.Since(started))
			}
			if len(frames) != 2 {
				t.Fatalf("got %d frames, want only steady startup and normal ring", len(frames))
			}
			assertLEDValuesPattern(t, frames[len(frames)-1], "RRRGGGGGGGGG")
			break
		}
		select {
		case <-device.done:
		case <-deadline:
			t.Fatal("startup did not restore booking display")
		}
	}
	assertRingPattern(t, app.Poll("abc", "192.0.2.10"), "RRRGGGGGGGGG")
	if rooms.created.Empty() == false || rooms.checkedInID != 0 || rooms.releasedID != 0 {
		t.Fatal("marketing sequence must not mutate Rooms")
	}
}

func TestStartupAnimationCancelsForDeviceInteraction(t *testing.T) {
	device := &recordingDeviceController{done: make(chan struct{}, 100)}
	app := NewApp(&recordingRooms{}, AppConfig{DeviceController: device, StartupAnimation: true, CommitWait: time.Hour, Logger: discardLogger{}})
	app.Poll("abc", "192.0.2.10")
	waitForDeviceFrames(t, device, 1)
	if err := app.HandleEvent(context.Background(), "short_press", "abc", "192.0.2.10"); err != nil {
		t.Fatal(err)
	}
	defer app.CommitPending(context.Background())
	assertRingPattern(t, app.Poll("abc", "192.0.2.10"), "BBBGGGGGGGGG")
	if app.Poll("abc", "192.0.2.10").LEDValues.Colors[0].Hex() == "#FF0085" {
		t.Fatal("canceled animation restarted")
	}
}

func TestStartupAnimationWaitsForReachableDevice(t *testing.T) {
	app := NewApp(&recordingRooms{}, AppConfig{DeviceController: &recordingDeviceController{}, StartupAnimation: true, Logger: discardLogger{}})
	if app.Poll("abc", "127.0.0.1").LEDValues.Colors[0].Hex() == "#FF0085" {
		t.Fatal("browser poll must not trigger startup")
	}
}

func TestStartupAnimationSuppressesCompetingLEDWrites(t *testing.T) {
	device := &recordingDeviceController{done: make(chan struct{}, 100)}
	app := NewApp(&recordingRooms{}, AppConfig{DeviceController: device, StartupAnimation: true, Logger: discardLogger{}})
	app.Poll("abc", "192.0.2.10")
	waitForDeviceFrames(t, device, 1)
	defer app.cancelStartupAnimation("abc", "192.0.2.10")
	colors := make([]LEDColor, 12)
	for i := range colors {
		colors[i] = ColorFromHex(FreeGreen, 100)
	}
	app.sendRingToDevice("192.0.2.10", LEDCommand{Command: "set_leds_individual", LEDValues: LEDValues{Colors: colors}})
	frames := device.snapshotFrames()
	if frames[len(frames)-1].Colors[0].Hex() == FreeGreen {
		t.Fatal("competing normal LED push overwrote marketing palette")
	}
}

type unavailableStartupDevice struct{}

func (*unavailableStartupDevice) SetIndividualLEDs(context.Context, string, LEDValues) error {
	return errors.New("local LED API unavailable")
}

func TestFailedFirstStartupDoesNotSurpriseRestartOnLaterPoll(t *testing.T) {
	app := NewApp(&recordingRooms{}, AppConfig{DeviceController: &unavailableStartupDevice{}, StartupAnimation: true, Logger: discardLogger{}})
	app.Poll("abc", "192.0.2.10")
	deadline := time.After(6500 * time.Millisecond)
	for {
		app.mu.Lock()
		state := app.startupAnimations[startupDeviceKey("abc", "192.0.2.10")]
		finished := state != nil && !state.running
		app.mu.Unlock()
		if finished {
			break
		}
		select {
		case <-deadline:
			t.Fatal("failed initial startup should finish once")
		case <-time.After(20 * time.Millisecond):
		}
	}
	command := app.Poll("abc", "192.0.2.10")
	if command.LEDValues.Colors[0].Hex() == "#FF0085" {
		t.Fatal("later poll unexpectedly restarted marketing sequence")
	}
}

func TestCanceledStartupDoesNotOverwriteDeviceInteraction(t *testing.T) {
	device := &recordingDeviceController{done: make(chan struct{}, 100)}
	app := NewApp(&recordingRooms{}, AppConfig{DeviceController: device, StartupAnimation: true, Logger: discardLogger{}})
	ctx, cancel := context.WithCancel(context.Background())
	animation := &startupAnimation{deviceIP: "192.0.2.10", cancel: cancel}
	app.startupAnimations = map[string]*startupAnimation{startupDeviceKey("abc", "192.0.2.10"): animation}
	// A button/NFC event now owns the device. A canceled initializer must not
	// enqueue another ambient snapshot behind the event's feedback.
	cancel()
	app.runStartupAnimation(ctx, animation, "abc", "192.0.2.10")
	select {
	case <-device.done:
		t.Fatal("canceled startup pushed an ambient frame over device interaction")
	case <-time.After(150 * time.Millisecond):
	}
}

func TestStartupRestoresFreeRoomAndKeepsNormalButtonSelection(t *testing.T) {
	device := &recordingDeviceController{done: make(chan struct{}, 100)}
	app := NewApp(&recordingRooms{}, AppConfig{DeviceController: device, StartupAnimation: true, CommitWait: time.Hour, ButtonAnimationFrameDelay: time.Millisecond, Logger: discardLogger{}})
	if err := app.RefreshActiveBooking(context.Background()); err != nil {
		t.Fatal(err)
	}
	app.Poll("abc", "192.0.2.10")
	frames := waitForDeviceFrames(t, device, 1)
	deadline := time.After(6500 * time.Millisecond)
	for len(frames) < 2 {
		select {
		case <-device.done:
			frames = device.snapshotFrames()
		case <-deadline:
			t.Fatal("startup never restored free room")
		}
	}
	assertLEDValuesPattern(t, frames[1], "GGGGGGGGGGGG")
	if err := app.HandleEvent(context.Background(), "short_press", "abc", "192.0.2.10"); err != nil {
		t.Fatal(err)
	}
	waitForDeviceFrames(t, device, 5)
	frames = device.snapshotFrames()
	assertLEDValuesPattern(t, frames[len(frames)-1], "BBBGGGGGGGGG")
	defer app.CommitPending(context.Background())
}

func TestStartupRestoreDoesNotCommitExpiredSelection(t *testing.T) {
	device := &recordingDeviceController{done: make(chan struct{}, 100)}
	rooms := &recordingRooms{}
	now := time.Date(2026, 10, 9, 10, 0, 0, 0, time.UTC)
	app := NewApp(rooms, AppConfig{Now: func() time.Time { return now }, DeviceController: device, Logger: discardLogger{}})
	app.exactBusyKnown = true
	selection := &pendingSelection{baseStart: now, baseEnd: now, additions: []time.Duration{15 * time.Minute}, deadline: now.Add(-time.Second)}
	app.pending = selection
	app.restoreStartupRing(context.Background(), "192.0.2.10")
	if app.pending != selection || !rooms.created.Empty() {
		t.Fatal("startup display restoration committed a pending booking")
	}
	assertLEDValuesPattern(t, device.snapshotFrames()[0], "BBBGGGGGGGGG")
}

func TestStartupRestoreHonorsCancellationWhileDeviceWriteQueued(t *testing.T) {
	device := &recordingDeviceController{done: make(chan struct{}, 100)}
	app := NewApp(&recordingRooms{}, AppConfig{DeviceController: device, Logger: discardLogger{}})
	ctx, cancel := context.WithCancel(context.Background())
	app.devicePushMu.Lock()
	done := make(chan struct{})
	go func() {
		app.restoreStartupRing(ctx, "192.0.2.10")
		close(done)
	}()
	cancel()
	app.devicePushMu.Unlock()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("canceled restoration never completed")
	}
	if len(device.snapshotFrames()) != 0 {
		t.Fatal("canceled restoration pushed stale display after device interaction")
	}
}

func TestFinishedStartupStillCancelsItsQueuedRestoration(t *testing.T) {
	app := NewApp(&recordingRooms{}, AppConfig{Logger: discardLogger{}})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	app.startupAnimations = map[string]*startupAnimation{
		startupDeviceKey("abc", "192.0.2.10"): {running: false, deviceIP: "192.0.2.10", cancel: cancel},
	}
	// The visible sequence may have ended while its restore write is queued.
	// Device interaction must still revoke that write's context.
	app.cancelStartupAnimation("abc", "192.0.2.10")
	if ctx.Err() == nil {
		t.Fatal("finished startup retained permission to push queued restoration")
	}
}
