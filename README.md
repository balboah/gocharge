# Go Charge

This tool is used for switching a [Halo Charger](https://charge-amps.com/products/charging-stations/halo-wallbox/) on/off depending on when the hourly price on [Tibber](tibber.com) is the cheapest.

## Dependencies

Both services requires that you have requested your personal API access tokens.
This example runs through Docker and is implemendeted in such a way that it requires an external scheduler a.k.a. crontab.

## Example

### One time run with Docker
```bash
docker run --rm -ti balboah/gocharge -hours 4 -tibberToken *changeme* -haloToken *changeme* -haloCharger *changeme* -haloSerial *changeme*
```

### Kubernetes cronjob

```bash
# Set API keys and select charger
kubectl create secret generic gocharge --from-literal=tibber-token='changeme' --from-literal=halo-token='changeme' --from-literal=halo-charger='changeme' --from-literal=halo-serial='changeme'

# Run on cronjob schedule
kubectl create --save-config -f kube/cronjob.yaml
```
