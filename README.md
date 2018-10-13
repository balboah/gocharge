# Go Charge

This tool is used for switching a [Halo Charger](https://charge-amps.com/products/charging-stations/halo-wallbox/) on/off depending on when the hourly price on [Tibber](tibber.com) is the cheapest.

## Dependencies

Both services requires that you have requested your personal API access tokens.
This example runs through Docker and is implemendeted in such a way that it requires an external scheduler a.k.a. crontab.
