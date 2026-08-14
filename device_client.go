package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type DeviceController interface {
	SetIndividualLEDs(ctx context.Context, deviceIP string, values LEDValues) error
}

type PukkDeviceHTTPClient struct {
	httpClient *http.Client
}

func NewPukkDeviceHTTPClient() *PukkDeviceHTTPClient {
	return &PukkDeviceHTTPClient{httpClient: &http.Client{Timeout: 2 * time.Second}}
}

func (c *PukkDeviceHTTPClient) SetIndividualLEDs(ctx context.Context, deviceIP string, values LEDValues) error {
	if !isPushableDeviceIP(deviceIP) {
		return nil
	}
	body := struct {
		LEDValues LEDValues `json:"led_values"`
	}{LEDValues: values}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://"+deviceIP+"/api/setLeds/individual", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("device LED push returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func isPushableDeviceIP(deviceIP string) bool {
	ip := net.ParseIP(deviceIP)
	return ip != nil && !ip.IsLoopback() && !ip.IsUnspecified()
}
