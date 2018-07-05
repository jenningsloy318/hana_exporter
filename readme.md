# hana_exporter 

this is initial hana exporter to monitor hana database.

gather all hana info from its system table/views, following the description at 

http://sap.optimieren.de/hana/hana/html/sys_statistics_views.html

# Usage 
create a configuration `hana-exporter.yml`, which contains the credentials of hana instance.
```yaml
credentials:
    default:
        user: "user"
        pass: "password"
    192.168.100.237:30015:
        user: "SYSTEM"
        pass: "Password"
```
then start hana-exporter via 
```sh
hana_exporter --config.file=hana-exporter.yml
```

then we can get the metrics via 
```
curl http://<hana-export host>:9460/metrics?target=192.168.100.237:30015

```

## NOTE: The usre configured at lest have `select` permission on schema `SYS`, all the collector will collect the info from tables/views under this schema.

## prometheus job conf
add hana-exporter job conif as following
```yaml
  - job_name: 'hana-exporter'

    # metrics_path defaults to '/metrics'
    metrics_path: /metrics


    # scheme defaults to 'http'.

    static_configs:
    - targets:
       - 192.168.100.237:30015   
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: localhost:9460  ### the address of the hana-exporter address
````