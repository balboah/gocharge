package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/balboah/gocharge"
)

func main() {
	tibberToken := flag.String("tibberToken", env("TIBBER_TOKEN", ""), "access token for the Tibber API")
	haloToken := flag.String("haloToken", env("HALO_TOKEN", ""), "API access key")
	haloCharger := flag.String("haloCharger", env("HALO_CHARGER", ""), "charger code (wifi password)")
	haloSerial := flag.String("haloSerial", env("HALO_SERIAL", ""), "charger serial number")
	hours := flag.Int("hours", 4, "number of hours charging per day")
	beforeHour := flag.Int("beforeHour", 8, "guaranteed charging before this hour")
	flag.Parse()

	tibber := gocharge.NewTibberClient(*tibberToken)
	halo := gocharge.NewHaloClient(
		http.DefaultClient, *haloToken, *haloCharger, *haloSerial,
	)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ok, err := gocharge.ShouldCharge(ctx, tibber, time.Now(), *hours, *beforeHour)
	if err != nil {
		log.Fatal(err)
	}
	haloSwitch := gocharge.HaloSwitchOnline
	if ok {
		log.Println("charging")
	} else {
		log.Println("not charging")
		haloSwitch = gocharge.HaloSwitchOffline
	}
	if err := halo.ChargerSwitch(ctx, haloSwitch); err != nil {
		log.Fatal(err)
	}
}

func env(key, fallback string) string {
	v, ok := os.LookupEnv(key)
	if ok {
		return v
	}
	return fallback
}
