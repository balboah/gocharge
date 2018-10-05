package gocharge

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const haloAPI = "https://iapi.charge.space/v1"

// HaloStatusCodeError returned on non-200.
type HaloStatusCodeError struct {
	StatusCode int
	Request    *http.Request
}

func (err HaloStatusCodeError) Error() string {
	return http.StatusText(err.StatusCode)
}

type HaloClient struct {
	client      *http.Client
	apiKey      string
	chargerCode string
	serial      string
}

func NewHaloClient(client *http.Client, apiKey, chargerCode, serial string) *HaloClient {
	return &HaloClient{
		client, apiKey, chargerCode, serial,
	}
}

func (h *HaloClient) Status() (*HaloStatusResponse, error) {
	req, err := h.request("GET", "status")
	if err != nil {
		return nil, err
	}
	r, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, HaloStatusCodeError{r.StatusCode, req}
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	resp := HaloStatusResponse{}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

type HaloStatusResponse struct {
	Status              HaloStatus `json:"chargerStatus"`
	TotalConsumptionKWh float64    `json:"totalConsumptionKwh"`
	ChargingCurrent     float64    `json:"chargingCurrent"`
	ChargingVoltage     float64    `json:"chargingVoltage"`
}

func (s HaloStatusResponse) String() string {
	return fmt.Sprintf(
		"%s (%.2fKWh %.1fA %.1fV)",
		s.Status, s.TotalConsumptionKWh, s.ChargingCurrent, s.ChargingVoltage,
	)
}

func (h *HaloClient) ChargerSwitch(s HaloSwitch) error {
	req, err := h.request("PUT", string(s))
	if err != nil {
		return err
	}
	r, err := h.client.Do(req)
	if err != nil {
		return err
	}
	r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return HaloStatusCodeError{r.StatusCode, req}
	}
	return nil
}

func (h *HaloClient) request(method, endpoint string) (*http.Request, error) {
	req, err := http.NewRequest(
		method, strings.Join([]string{haloAPI, "chargers", h.serial, endpoint}, "/"), nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Basic "+h.apiKey)
	req.Header.Set("ChargerCode", h.chargerCode)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}

type HaloStatus string

const (
	HaloStatusOnline   HaloStatus = "Online"
	HaloStatusOffline  HaloStatus = "Offline"
	HaloStatusCharging HaloStatus = "Charging"
)

type HaloSwitch string

const (
	HaloSwitchOnline  HaloSwitch = "on"
	HaloSwitchOffline HaloSwitch = "off"
)
