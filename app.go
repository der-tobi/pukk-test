package main

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Rooms interface {
	FreeBusyReader
	FindBookings(ctx context.Context, start, end time.Time) ([]Booking, error)
	FindCurrentBooking(ctx context.Context, now time.Time) (*Booking, error)
	CreateAdHocBooking(ctx context.Context, start, end time.Time) (*Booking, error)
	ExtendBooking(ctx context.Context, booking Booking, end time.Time) (*Booking, error)
	CheckInBooking(ctx context.Context, booking Booking) (*Booking, error)
	ReleaseBooking(ctx context.Context, booking Booking) error
}

type AppConfig struct {
	ResourceID                int
	Now                       func() time.Time
	CommitWait                time.Duration
	RoomsRefreshInterval      time.Duration
	DeviceController          DeviceController
	DevicePushTimeout         time.Duration
	ButtonAnimationFrameDelay time.Duration
	Logger                    Logger
}

type App struct {
	rooms Rooms
	cache *BookingCache
	cfg   AppConfig

	mu                sync.Mutex
	active            *Booking
	exactBusy         []TimeRange
	exactBusyKnown    bool
	pending           *pendingSelection
	pendingGeneration int
	deviceIPs         map[string]string
	debug             DebugStatus
}

type DebugStatus struct {
	ResourceID                   int          `json:"resourceId"`
	ServerTime                   time.Time    `json:"serverTime"`
	TotalRequests                int          `json:"totalRequests"`
	PollRequests                 int          `json:"pollRequests"`
	EventRequests                int          `json:"eventRequests"`
	LastRequest                  DebugRequest `json:"lastRequest"`
	LastRefreshOK                time.Time    `json:"lastRefreshOk,omitempty"`
	LastRefreshError             string       `json:"lastRefreshError,omitempty"`
	LastAvailabilityRefreshOK    time.Time    `json:"lastAvailabilityRefreshOk,omitempty"`
	LastAvailabilityRefreshError string       `json:"lastAvailabilityRefreshError,omitempty"`
	LastBookingsRefreshOK        time.Time    `json:"lastBookingsRefreshOk,omitempty"`
	LastBookingsRefreshError     string       `json:"lastBookingsRefreshError,omitempty"`
	AvailabilityBits             string       `json:"availabilityBits,omitempty"`
	AvailabilityStart            time.Time    `json:"availabilityStart,omitempty"`
	AvailabilityEnd              time.Time    `json:"availabilityEnd,omitempty"`
	AvailabilityIntervalMinutes  int          `json:"availabilityIntervalMinutes,omitempty"`
	ActiveBookingID              int          `json:"activeBookingId,omitempty"`
	PendingUntil                 time.Time    `json:"pendingUntil,omitempty"`
	PendingStart                 time.Time    `json:"pendingStart,omitempty"`
	PendingEnd                   time.Time    `json:"pendingEnd,omitempty"`
	PendingBlocks                int          `json:"pendingBlocks,omitempty"`
	LastLEDHex                   []string     `json:"lastLedHex,omitempty"`
	LastBlueLEDs                 int          `json:"lastBlueLeds,omitempty"`
	LastRedLEDs                  int          `json:"lastRedLeds,omitempty"`
	LastDevicePushOK             time.Time    `json:"lastDevicePushOk,omitempty"`
	LastDevicePushError          string       `json:"lastDevicePushError,omitempty"`
	LastDevicePushIP             string       `json:"lastDevicePushIp,omitempty"`
	ExactBusyKnown               bool         `json:"exactBusyKnown"`
	ExactBusyCount               int          `json:"exactBusyCount"`
	ExactBusyRanges              []TimeRange  `json:"exactBusyRanges,omitempty"`
	DeviceIPs                    []string     `json:"deviceIps,omitempty"`
}

type DebugRequest struct {
	Time     time.Time `json:"time,omitempty"`
	Method   string    `json:"method,omitempty"`
	Action   string    `json:"action,omitempty"`
	MAC      string    `json:"mac,omitempty"`
	RemoteIP string    `json:"remoteIp,omitempty"`
	Path     string    `json:"path,omitempty"`
}

type pendingSelection struct {
	active      *Booking
	baseStart   time.Time
	baseEnd     time.Time
	additions   []time.Duration
	undoing     bool
	timer       *time.Timer
	commitAfter time.Duration
	deadline    time.Time
	deviceIP    string
}

func NewApp(rooms Rooms, cfg AppConfig) *App {
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.CommitWait == 0 {
		cfg.CommitWait = 5 * time.Second
	}
	if cfg.RoomsRefreshInterval == 0 {
		cfg.RoomsRefreshInterval = 30 * time.Second
	}
	if cfg.DevicePushTimeout == 0 {
		cfg.DevicePushTimeout = 2 * time.Second
	}
	if cfg.ButtonAnimationFrameDelay == 0 {
		cfg.ButtonAnimationFrameDelay = 100 * time.Millisecond
	}
	if cfg.Logger == nil {
		cfg.Logger = standardLogger{}
	}
	return &App{
		rooms: rooms,
		cache: NewBookingCache(rooms, CacheConfig{
			ResourceID:      cfg.ResourceID,
			DisplayWindow:   time.Hour,
			RefreshInterval: 5 * time.Minute,
			SafetyMargin:    10 * time.Minute,
			IntervalMinutes: 1,
		}),
		cfg:       cfg,
		deviceIPs: map[string]string{},
	}
}

func (a *App) RoomsRefreshInterval() time.Duration {
	return a.cfg.RoomsRefreshInterval
}

func (a *App) Refresh(ctx context.Context) error {
	availabilityErr := a.RefreshAvailability(ctx)
	activeErr := a.RefreshActiveBooking(ctx)
	if availabilityErr != nil {
		return availabilityErr
	}
	return activeErr
}

func (a *App) RefreshAvailability(ctx context.Context) error {
	now := a.cfg.Now().UTC()
	err := a.cache.Refresh(ctx, now)
	a.mu.Lock()
	window := a.cache.Window()
	if err != nil {
		a.debug.LastAvailabilityRefreshError = err.Error()
		a.debug.LastRefreshError = err.Error()
	} else {
		a.debug.LastAvailabilityRefreshOK = now
		a.debug.LastAvailabilityRefreshError = ""
		a.debug.AvailabilityBits = window.Data
		a.debug.AvailabilityStart = window.Start
		a.debug.AvailabilityEnd = window.End
		a.debug.AvailabilityIntervalMinutes = window.IntervalMinutes
	}
	a.mu.Unlock()
	return err
}

func (a *App) RefreshActiveBooking(ctx context.Context) error {
	now := a.cfg.Now().UTC()
	windowEnd := now.Add(time.Hour)
	bookings, err := a.rooms.FindBookings(ctx, now.Add(-24*time.Hour), windowEnd)
	a.mu.Lock()
	defer a.mu.Unlock()
	if err != nil {
		if a.active == nil || !now.Before(a.active.End) {
			a.active = nil
		}
		a.exactBusy = nil
		a.exactBusyKnown = false
		a.debug.LastRefreshError = err.Error()
		a.debug.LastBookingsRefreshError = err.Error()
		return err
	}

	active, exactBusy := selectActiveAndExactBusy(bookings, a.cfg.ResourceID, now, windowEnd)
	if active != nil {
		a.active = active
	} else if a.active != nil && now.Before(a.active.End) {
		active = cloneBooking(a.active)
		if active.Start.Before(windowEnd) && active.End.After(now) {
			exactBusy = append(exactBusy, TimeRange{Start: active.Start, End: active.End})
		}
	} else {
		a.active = nil
	}
	a.exactBusy = exactBusy
	a.exactBusyKnown = true
	a.debug.LastRefreshOK = now
	a.debug.LastRefreshError = ""
	a.debug.LastBookingsRefreshOK = now
	a.debug.LastBookingsRefreshError = ""
	return nil
}

func selectActiveAndExactBusy(bookings []Booking, resourceID int, now, windowEnd time.Time) (*Booking, []TimeRange) {
	var active *Booking
	exactBusy := make([]TimeRange, 0, len(bookings))
	for _, booking := range bookings {
		if !booking.MatchesResource(resourceID) || !booking.End.After(now) || !booking.Start.Before(windowEnd) {
			continue
		}
		if !now.Before(booking.Start) && now.Before(booking.End) {
			b := booking
			active = &b
		}
		exactBusy = append(exactBusy, TimeRange{Start: booking.Start, End: booking.End})
	}
	return active, exactBusy
}

func (a *App) RecordRequest(method, path, action, mac, remoteIP string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.debug.TotalRequests++
	if action == "poll" {
		a.debug.PollRequests++
	} else if action != "" {
		a.debug.EventRequests++
	}
	a.debug.LastRequest = DebugRequest{
		Time:     a.cfg.Now().UTC(),
		Method:   method,
		Action:   action,
		MAC:      mac,
		RemoteIP: remoteIP,
		Path:     path,
	}
}

func (a *App) DebugStatus() DebugStatus {
	a.mu.Lock()
	defer a.mu.Unlock()
	status := a.debug
	status.ResourceID = a.cfg.ResourceID
	status.ServerTime = a.cfg.Now().UTC()
	if a.active != nil {
		status.ActiveBookingID = a.active.BookingID()
	}
	status.ExactBusyRanges = append([]TimeRange(nil), a.exactBusy...)
	status.ExactBusyKnown = a.exactBusyKnown
	status.ExactBusyCount = len(a.exactBusy)
	if a.pending != nil {
		pendingRange := a.pending.Range()
		status.PendingStart = pendingRange.Start
		status.PendingEnd = pendingRange.End
		status.PendingUntil = a.pending.deadline
		status.PendingBlocks = len(a.pending.additions)
	}
	status.DeviceIPs = make([]string, 0, len(a.deviceIPs))
	for _, ip := range a.deviceIPs {
		status.DeviceIPs = append(status.DeviceIPs, ip)
	}
	return status
}

func (a *App) Poll(mac, deviceIP string) LEDCommand {
	now := a.cfg.Now().UTC()
	a.commitExpiredPending(now)
	a.mu.Lock()
	if mac != "" && deviceIP != "" {
		a.deviceIPs[mac] = deviceIP
	}
	active := cloneBooking(a.active)
	exactBusy := append([]TimeRange(nil), a.exactBusy...)
	exactBusyKnown := a.exactBusyKnown
	var provisional []TimeRange
	if a.pending != nil {
		if r := a.pending.Range(); !r.Empty() {
			provisional = []TimeRange{r}
		}
	}
	a.mu.Unlock()

	command := a.renderRing(now, active, exactBusy, exactBusyKnown, provisional)
	a.recordLEDCommand(command)
	a.pushLEDCommand(deviceIP, command)
	return command
}

func (a *App) renderRing(now time.Time, active *Booking, exactBusy []TimeRange, exactBusyKnown bool, provisional []TimeRange) LEDCommand {
	snapshot := a.cache.Snapshot(now)
	return RenderRing(RingInput{
		Now:         now,
		Busy:        snapshot.Busy,
		Known:       snapshot.Known,
		Active:      active,
		ExactBusy:   exactBusy,
		ExactKnown:  exactBusyKnown,
		Provisional: provisional,
	})
}

func (a *App) recordLEDCommand(command LEDCommand) {
	hex := make([]string, 0, len(command.LEDValues.Colors))
	blue := 0
	red := 0
	for _, color := range command.LEDValues.Colors {
		value := color.Hex()
		hex = append(hex, value)
		switch value {
		case ProvisionalBlue:
			blue++
		case BusyRed:
			red++
		}
	}
	a.mu.Lock()
	a.debug.LastLEDHex = hex
	a.debug.LastBlueLEDs = blue
	a.debug.LastRedLEDs = red
	a.mu.Unlock()
}

func (a *App) pushLEDCommand(deviceIP string, command LEDCommand) {
	if a.cfg.DeviceController == nil || deviceIP == "" {
		return
	}
	values := command.LEDValues
	timeout := a.cfg.DevicePushTimeout
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := a.pushLEDValues(ctx, deviceIP, values); err != nil {
			a.cfg.Logger.Printf("PuKK local LED push to %s failed: %v", deviceIP, err)
			return
		}
	}()
}

func (a *App) pushLEDValues(ctx context.Context, deviceIP string, values LEDValues) error {
	if a.cfg.DeviceController == nil || deviceIP == "" {
		return nil
	}
	err := a.cfg.DeviceController.SetIndividualLEDs(ctx, deviceIP, values)
	now := a.cfg.Now().UTC()
	a.mu.Lock()
	a.debug.LastDevicePushIP = deviceIP
	if err != nil {
		a.debug.LastDevicePushError = err.Error()
		a.mu.Unlock()
		return err
	}
	a.debug.LastDevicePushOK = now
	a.debug.LastDevicePushError = ""
	a.mu.Unlock()
	a.recordLEDCommand(LEDCommand{Command: "set_leds_individual", LEDValues: values})
	return nil
}

func (a *App) HandleEvent(ctx context.Context, action, mac, deviceIP string) error {
	switch action {
	case "short_press", "double_press", "multiple_press":
		a.handleShortPressForDevice(ctx, mac, deviceIP)
		return nil
	case "nfc":
		return a.handleNFC(ctx, mac, deviceIP)
	case "long_press_3s":
		return a.handleCheckout(ctx, mac, deviceIP)
	default:
		return nil
	}
}

func (a *App) handleShortPress() {
	a.handleShortPressForDevice(context.Background(), "", "")
}

func (a *App) handleShortPressForDevice(ctx context.Context, mac, deviceIP string) {
	now := a.cfg.Now().UTC()
	a.ensureExactBusyForInteraction(ctx)
	a.mu.Lock()
	if mac != "" && deviceIP != "" {
		a.deviceIPs[mac] = deviceIP
	}
	if deviceIP == "" && mac != "" {
		deviceIP = a.deviceIPs[mac]
	}

	active := cloneBooking(a.active)
	exactBusy := append([]TimeRange(nil), a.exactBusy...)
	exactBusyKnown := a.exactBusyKnown
	var beforeProvisional []TimeRange
	var beforeEnd time.Time
	if a.pending != nil {
		beforeEnd = a.pending.CurrentEnd()
		if r := a.pending.Range(); !r.Empty() {
			beforeProvisional = []TimeRange{r}
		}
	}

	if a.pending == nil {
		baseStart := now
		baseEnd := now
		var active *Booking
		if a.active != nil && !now.Before(a.active.Start) && now.Before(a.active.End) {
			active = cloneBooking(a.active)
			baseStart = a.active.End
			baseEnd = a.active.End
		}
		a.pending = &pendingSelection{
			active:      active,
			baseStart:   baseStart,
			baseEnd:     baseEnd,
			commitAfter: a.cfg.CommitWait,
		}
		beforeEnd = baseEnd
	}
	if deviceIP != "" {
		a.pending.deviceIP = deviceIP
	}

	a.pending.Step(a.maxSelectableEnd(now, a.pending.baseEnd, exactBusy, exactBusyKnown))
	a.resetPendingTimerLocked(now)
	generation := a.pendingGeneration
	afterEnd := a.pending.CurrentEnd()
	var afterProvisional []TimeRange
	if r := a.pending.Range(); !r.Empty() {
		afterProvisional = []TimeRange{r}
	}
	a.mu.Unlock()

	before := a.renderRing(now, active, exactBusy, exactBusyKnown, beforeProvisional)
	after := a.renderRing(now, active, exactBusy, exactBusyKnown, afterProvisional)
	indexes := changedLEDIndexes(before, after)
	if afterEnd.Before(beforeEnd) {
		reverseInts(indexes)
	}
	a.animateLEDTransitionAsync(deviceIP, generation, before, after, indexes)
}

func (a *App) ensureExactBusyForInteraction(ctx context.Context) {
	a.mu.Lock()
	exactBusyKnown := a.exactBusyKnown
	a.mu.Unlock()
	if exactBusyKnown {
		return
	}
	if err := a.RefreshActiveBooking(ctx); err != nil {
		a.cfg.Logger.Printf("button exact booking refresh failed; continuing without exact booking cap: %v", err)
	}
}

func (a *App) maxSelectableEnd(now, baseEnd time.Time, exactBusy []TimeRange, exactBusyKnown bool) time.Time {
	windowEnd := now.Add(time.Hour)
	if exactBusyKnown {
		maxEnd := windowEnd
		for _, busyRange := range exactBusy {
			if busyRange.Empty() || !busyRange.End.After(baseEnd) {
				continue
			}
			if !busyRange.Start.After(baseEnd) {
				return baseEnd
			}
			if busyRange.Start.Before(maxEnd) {
				maxEnd = busyRange.Start
			}
		}
		return maxEnd
	}
	return windowEnd
}

func (a *App) resetPendingTimerLocked(now time.Time) {
	if a.pending == nil || a.pending.commitAfter <= 0 {
		return
	}
	a.pendingGeneration++
	generation := a.pendingGeneration
	a.pending.deadline = now.Add(a.pending.commitAfter)
	if a.pending.timer != nil {
		a.pending.timer.Stop()
	}
	a.pending.timer = time.AfterFunc(a.pending.commitAfter, func() {
		if err := a.CommitPendingGeneration(context.Background(), generation); err != nil {
			a.cfg.Logger.Printf("commit pending selection failed: %v", err)
		}
	})
}

func (p *pendingSelection) Step(maxEnd time.Time) {
	current := p.CurrentEnd()
	if p.undoing || !current.Before(maxEnd) {
		p.undoing = true
		if len(p.additions) > 0 {
			p.additions = p.additions[:len(p.additions)-1]
		}
		if len(p.additions) == 0 {
			p.undoing = false
		}
		return
	}

	delta := 15 * time.Minute
	if remaining := maxEnd.Sub(current); remaining < delta {
		delta = remaining
	}
	if delta > 0 {
		p.additions = append(p.additions, delta)
	}
}

func (p *pendingSelection) CurrentEnd() time.Time {
	end := p.baseEnd
	for _, add := range p.additions {
		end = end.Add(add)
	}
	return end
}

func (p *pendingSelection) Range() TimeRange {
	return TimeRange{Start: p.baseStart, End: p.CurrentEnd()}
}

func (a *App) CommitPending(ctx context.Context) error {
	return a.CommitPendingGeneration(ctx, 0)
}

func (a *App) CommitPendingGeneration(ctx context.Context, generation int) error {
	pending, animationGeneration := a.takePendingGeneration(generation)
	if pending == nil || len(pending.additions) == 0 {
		return nil
	}
	a.applyOptimisticPending(pending)
	a.animatePendingCommit(pending, animationGeneration)
	return a.commitSelection(ctx, pending)
}

func (a *App) commitExpiredPending(now time.Time) {
	pending, animationGeneration := a.takeExpiredPending(now)
	if pending == nil || len(pending.additions) == 0 {
		return
	}
	a.applyOptimisticPending(pending)
	go func() {
		a.animatePendingCommit(pending, animationGeneration)
		if err := a.commitSelection(context.Background(), pending); err != nil {
			a.cfg.Logger.Printf("expired pending selection commit failed: %v", err)
		}
	}()
}

func (a *App) takeExpiredPending(now time.Time) (*pendingSelection, int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.pending == nil || a.pending.deadline.IsZero() || now.Before(a.pending.deadline) {
		return nil, 0
	}
	pending := a.pending
	a.pending = nil
	if pending.timer != nil {
		pending.timer.Stop()
	}
	a.pendingGeneration++
	return pending, a.pendingGeneration
}

func (a *App) takePendingGeneration(generation int) (*pendingSelection, int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if generation != 0 && generation != a.pendingGeneration {
		return nil, 0
	}
	pending := a.pending
	a.pending = nil
	if pending != nil && pending.timer != nil {
		pending.timer.Stop()
	}
	if pending != nil {
		a.pendingGeneration++
		return pending, a.pendingGeneration
	}
	return nil, 0
}

func (a *App) applyOptimisticPending(pending *pendingSelection) {
	optimistic := pending.optimisticBooking()
	a.mu.Lock()
	a.active = optimistic
	a.mu.Unlock()
}

func (a *App) commitSelection(ctx context.Context, pending *pendingSelection) error {
	end := pending.CurrentEnd()
	var err error
	var committed *Booking
	if pending.active != nil {
		committed, err = a.rooms.ExtendBooking(ctx, *pending.active, end)
	} else {
		committed, err = a.rooms.CreateAdHocBooking(ctx, pending.baseStart, end)
	}
	if err != nil {
		a.revertPending(pending)
		return err
	}
	if committed == nil {
		committed = pending.optimisticBooking()
	}
	if committed.Start.IsZero() {
		committed.Start = pending.bookingStart()
	}
	if committed.End.IsZero() {
		committed.End = end
	}
	a.mu.Lock()
	a.active = committed
	a.mu.Unlock()
	if err := a.Refresh(ctx); err != nil {
		a.cfg.Logger.Printf("post-commit Rooms refresh failed; keeping local booking state for display: %v", err)
	}
	return nil
}

func (a *App) revertPending(pending *pendingSelection) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if pending.active != nil {
		a.active = cloneBooking(pending.active)
		return
	}
	if a.active != nil && a.active.Start.Equal(pending.baseStart) && a.active.End.Equal(pending.CurrentEnd()) {
		a.active = nil
	}
}

func (p *pendingSelection) optimisticBooking() *Booking {
	checkinConfirmed := true
	id := 0
	if p.active != nil {
		checkinConfirmed = p.active.CheckinConfirmed
		id = p.active.BookingID()
	}
	return &Booking{
		ID:               id,
		Start:            p.bookingStart(),
		End:              p.CurrentEnd(),
		CheckinConfirmed: checkinConfirmed,
	}
}

func (p *pendingSelection) activeBookingID() int {
	if p.active == nil {
		return 0
	}
	return p.active.BookingID()
}

func (p *pendingSelection) bookingStart() time.Time {
	if p.active != nil && !p.active.Start.IsZero() {
		return p.active.Start
	}
	return p.baseStart
}

func (a *App) animatePendingCommit(pending *pendingSelection, generation int) {
	if pending.deviceIP == "" {
		return
	}
	now := a.cfg.Now().UTC()
	beforeActive := cloneBooking(pending.active)
	before := a.renderRing(now, beforeActive, nil, false, []TimeRange{pending.Range()})
	after := a.renderRing(now, pending.optimisticBooking(), nil, false, nil)
	indexes := changedLEDIndexes(before, after)
	a.animateLEDTransition(pending.deviceIP, generation, before, after, indexes)
}

func (a *App) animateLEDTransitionAsync(deviceIP string, generation int, before, after LEDCommand, indexes []int) {
	if deviceIP == "" || len(indexes) == 0 {
		return
	}
	go a.animateLEDTransition(deviceIP, generation, before, after, indexes)
}

func (a *App) animateLEDTransition(deviceIP string, generation int, before, after LEDCommand, indexes []int) {
	if a.cfg.DeviceController == nil || deviceIP == "" || len(indexes) == 0 {
		return
	}
	colors := append([]LEDColor(nil), before.LEDValues.Colors...)
	for i, idx := range indexes {
		if generation != 0 && !a.isAnimationGenerationCurrent(generation) {
			return
		}
		if idx < 0 || idx >= len(colors) || idx >= len(after.LEDValues.Colors) {
			continue
		}
		colors[idx] = after.LEDValues.Colors[idx]
		values := LEDValues{
			Colors:     append([]LEDColor(nil), colors...),
			DurationMS: after.LEDValues.DurationMS,
		}
		ctx, cancel := context.WithTimeout(context.Background(), a.cfg.DevicePushTimeout)
		err := a.pushLEDValues(ctx, deviceIP, values)
		cancel()
		if err != nil {
			a.cfg.Logger.Printf("PuKK local LED animation push to %s failed: %v", deviceIP, err)
			return
		}
		if i < len(indexes)-1 && a.cfg.ButtonAnimationFrameDelay > 0 {
			time.Sleep(a.cfg.ButtonAnimationFrameDelay)
		}
	}
}

func (a *App) isAnimationGenerationCurrent(generation int) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return generation != 0 && generation == a.pendingGeneration
}

func changedLEDIndexes(before, after LEDCommand) []int {
	limit := len(before.LEDValues.Colors)
	if len(after.LEDValues.Colors) < limit {
		limit = len(after.LEDValues.Colors)
	}
	indexes := make([]int, 0, limit)
	for i := 0; i < limit; i++ {
		if before.LEDValues.Colors[i] != after.LEDValues.Colors[i] {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

func reverseInts(values []int) {
	for i, j := 0, len(values)-1; i < j; i, j = i+1, j-1 {
		values[i], values[j] = values[j], values[i]
	}
}

func (a *App) handleNFC(ctx context.Context, mac, deviceIP string) error {
	now := a.cfg.Now().UTC()
	a.mu.Lock()
	if mac != "" && deviceIP != "" {
		a.deviceIPs[mac] = deviceIP
	}
	if deviceIP == "" && mac != "" {
		deviceIP = a.deviceIPs[mac]
	}
	active := cloneBooking(a.active)
	exactBusy := append([]TimeRange(nil), a.exactBusy...)
	exactBusyKnown := a.exactBusyKnown
	a.mu.Unlock()
	if active == nil {
		booking, err := a.rooms.FindCurrentBooking(ctx, now)
		if err != nil {
			return err
		}
		active = booking
	}
	if active == nil || active.CheckinConfirmed {
		return nil
	}
	before := a.renderRing(now, active, exactBusy, exactBusyKnown, nil)
	booking, err := a.rooms.CheckInBooking(ctx, *active)
	if err != nil {
		return err
	}
	if booking == nil {
		checkedIn := *active
		checkedIn.CheckinConfirmed = true
		booking = &checkedIn
	}
	if booking.Start.IsZero() {
		booking.Start = active.Start
	}
	if booking.End.IsZero() {
		booking.End = active.End
	}
	a.mu.Lock()
	a.active = booking
	a.mu.Unlock()
	after := a.renderRing(now, booking, exactBusy, exactBusyKnown, nil)
	indexes := changedLEDIndexes(before, after)
	a.animateLEDTransition(deviceIP, 0, before, after, indexes)
	return nil
}

func (a *App) handleCheckout(ctx context.Context, mac, deviceIP string) error {
	now := a.cfg.Now().UTC()
	a.mu.Lock()
	if mac != "" && deviceIP != "" {
		a.deviceIPs[mac] = deviceIP
	}
	if deviceIP == "" && mac != "" {
		deviceIP = a.deviceIPs[mac]
	}
	active := cloneBooking(a.active)
	exactBusy := append([]TimeRange(nil), a.exactBusy...)
	exactBusyKnown := a.exactBusyKnown
	a.mu.Unlock()
	if active == nil {
		booking, err := a.rooms.FindCurrentBooking(ctx, now)
		if err != nil {
			return err
		}
		active = booking
	}
	if active == nil {
		return nil
	}

	before := a.renderRing(now, active, exactBusy, exactBusyKnown, nil)
	indexes := activeLEDIndexes(now, *active)
	reverseInts(indexes)
	blue := commandWithColor(before, indexes, ProvisionalBlue)
	a.animateLEDTransition(deviceIP, 0, before, blue, indexes)

	if err := a.rooms.ReleaseBooking(ctx, *active); err != nil {
		a.pushLEDCommand(deviceIP, before)
		return err
	}

	a.clearLocalBooking(active)
	green := commandWithColor(blue, indexes, FreeGreen)
	a.animateLEDTransition(deviceIP, 0, blue, green, indexes)
	return nil
}

func activeLEDIndexes(now time.Time, active Booking) []int {
	indexes := make([]int, 0, 12)
	activeRange := []TimeRange{{Start: active.Start, End: active.End}}
	for i := 0; i < 12; i++ {
		slotStart := now.Add(time.Duration(i) * 5 * time.Minute)
		slotEnd := slotStart.Add(5 * time.Minute)
		if inDisplayRanges(now, slotStart, slotEnd, activeRange) {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

func commandWithColor(command LEDCommand, indexes []int, hex string) LEDCommand {
	next := LEDCommand{
		Command: command.Command,
		LEDValues: LEDValues{
			Colors:     append([]LEDColor(nil), command.LEDValues.Colors...),
			DurationMS: command.LEDValues.DurationMS,
		},
	}
	color := ColorFromHex(hex, 100)
	for _, idx := range indexes {
		if idx >= 0 && idx < len(next.LEDValues.Colors) {
			next.LEDValues.Colors[idx] = color
		}
	}
	return next
}

func (a *App) clearLocalBooking(booking *Booking) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.active != nil && sameBookingIdentityOrRange(*a.active, *booking) {
		a.active = nil
	}
	filtered := a.exactBusy[:0]
	for _, busyRange := range a.exactBusy {
		if busyRange.Start.Equal(booking.Start) && busyRange.End.Equal(booking.End) {
			continue
		}
		filtered = append(filtered, busyRange)
	}
	a.exactBusy = filtered
}

func sameBookingIdentityOrRange(a, b Booking) bool {
	if a.BookingID() != 0 && b.BookingID() != 0 {
		return a.BookingID() == b.BookingID()
	}
	return a.Start.Equal(b.Start) && a.End.Equal(b.End)
}

func cloneBooking(b *Booking) *Booking {
	if b == nil {
		return nil
	}
	c := *b
	return &c
}

func (a *App) ValidateReady() error {
	if a.rooms == nil {
		return errors.New("rooms client is required")
	}
	return nil
}
