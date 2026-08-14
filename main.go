package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	logger := standardLogger{}
	password := promptPassword()
	rooms := NewRoomsHTTPClient(RoomsConfig{Password: password})
	app := NewApp(rooms, AppConfig{
		ResourceID:       DefaultResourceID,
		Logger:           logger,
		DeviceController: NewPukkDeviceHTTPClient(),
	})
	if err := app.ValidateReady(); err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	printServerAddresses(5000)
	go refreshLoop(ctx, app, logger)

	server := &http.Server{
		Addr:              ":5000",
		Handler:           NewServer(app, logger),
		ReadHeaderTimeout: 5 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func promptPassword() string {
	fmt.Print("3V Rooms password for tobiapi4: ")
	reader := bufio.NewReader(os.Stdin)
	password, _ := reader.ReadString('\n')
	return strings.TrimSpace(password)
}

func refreshLoop(ctx context.Context, app *App, logger Logger) {
	ticker := time.NewTicker(app.RoomsRefreshInterval())
	defer ticker.Stop()
	logger.Printf("fetching initial 3V Rooms availability...")
	if err := app.RefreshAvailability(ctx); err != nil {
		logger.Printf("initial 3V Rooms availability refresh failed; HTTP server remains reachable and unknown slots render busy: %v", err)
	} else {
		logAvailabilityStatus(app, logger, "initial 3V Rooms availability refresh completed")
	}
	if err := app.RefreshActiveBooking(ctx); err != nil {
		logger.Printf("initial active booking lookup failed; availability bitstring is still available: %v", err)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := app.Refresh(ctx); err != nil {
				logger.Printf("3V Rooms refresh failed; serving last-known data and treating gaps as busy: %v", err)
			} else {
				logAvailabilityStatus(app, logger, "3V Rooms refresh completed")
			}
		}
	}
}

func logAvailabilityStatus(app *App, logger Logger, prefix string) {
	status := app.DebugStatus()
	logger.Printf("%s: bits=%s window=%s..%s interval=%dmin", prefix, status.AvailabilityBits, status.AvailabilityStart.Format(time.RFC3339), status.AvailabilityEnd.Format(time.RFC3339), status.AvailabilityIntervalMinutes)
}

func printServerAddresses(port int) {
	ips := localIPv4s()
	fmt.Println("PuKK server listening on:")
	fmt.Printf("  http://localhost:%d\n", port)
	for _, ip := range ips {
		fmt.Printf("  http://%s:%d\n", ip, port)
	}
	fmt.Println()
	fmt.Println("Diagnostics:")
	fmt.Printf("  http://localhost:%d/healthz\n", port)
	fmt.Printf("  http://localhost:%d/debug/status\n", port)
	fmt.Println()
	fmt.Println("Configure the PuKK CMS Server Address to: http://<reachable-host-ip>:5000")
	if onlyDockerBridgeIPs(ips) {
		fmt.Println("Warning: this process only sees Docker bridge IPs. A PuKK on WiFi probably cannot reach these addresses.")
		fmt.Println("For real-device testing, run pukk-test.exe on the Windows host and use the Windows WiFi IP.")
	}
}

func onlyDockerBridgeIPs(ips []string) bool {
	if len(ips) == 0 {
		return false
	}
	for _, ip := range ips {
		if !strings.HasPrefix(ip, "172.17.") {
			return false
		}
	}
	return true
}

func localIPv4s() []string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var ips []string
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			ip = ip.To4()
			if ip != nil {
				ips = append(ips, ip.String())
			}
		}
	}
	return ips
}
