// +build integration

package gocharge

import (
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
	status, err := c.Status()
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
	err := c.ChargerSwitch(HaloSwitchOffline)
	if err != nil {
		t.Fatal(err)
	}
}
func TestHaloSwitchOnline(t *testing.T) {
	err := c.ChargerSwitch(HaloSwitchOnline)
	if err != nil {
		t.Fatal(err)
	}
}
