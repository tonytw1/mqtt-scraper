## MQTT scraper



Listens for metrics on a MQTT topic. Caches them for export as a Prometheus scraper target.

We have metrics which are published on an MQTT topic which we'd like to record in Prometheus.

The metrics arrived in this format:
```
metric_name:value
```




