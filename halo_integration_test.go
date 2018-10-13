// +build integration

package gocharge

import (
	"context"
	"flag"
	"net/http"
	"net/http/httputil"
	"testing"
)

var c *HaloClient

func init() {
	key := flag.String("haloToken", "", "API access key")
	charger := flag.String("haloCharger", "", "charger code (wifi password)")
	serial := flag.String("haloSerial", "", "charger serial number")
	flag.Parse()
	c = NewHaloClient(http.DefaultClient, *key, *charger, *serial)
}

func TestHaloStatus(t *testing.T) {
	status, err := c.Status(context.Background())
	if err != nil {
		switch v := err.(type) {
		case HaloStatusCodeError:
			b, _ := httputil.DumpRequest(v.Request, true)
			t.Log(v.StatusCode, string(b))
		}
		t.Fatal(err)
	}
	t.Log(status)
}

func TestHaloSwitchOffline(t *testing.T) {
	err := c.ChargerSwitch(context.Background(), HaloSwitchOffline)
	if err != nil {
		t.Fatal(err)
	}
}
func TestHaloSwitchOnline(t *testing.T) {
	err := c.ChargerSwitch(context.Background(), HaloSwitchOnline)
	if err != nil {
		t.Fatal(err)
	}
}
