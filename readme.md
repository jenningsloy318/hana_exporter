# hana_exporter 

this is initial hana exporter to monitor hana database with itself opentracing implementation, all prometheus metrcis and tracing functions are implemented by [opencensus](https://opencensus.io/).

gather all hana info from its system table/views, following the description at 

http://sap.optimieren.de/hana/hana/html/sys_statistics_views.html

all metrics exposed at `/metrics`.
# Usage 
create a configuration `hana.yml`, which contains the credentials of hana instance.
```yaml
credentials:
    default:
        user: "user"
        pass: "password"
Jaeger:
    agentEndpointURI: "192.168.1.1:6831"
    collectorEndpointURI: "http://192.168.1.1:14268"
```

then start hana-exporter via 

```sh
hana_exporter --config.file=hana_exporter.yml
```

then we can get the metrics via 
```
curl http://<hana-export host>:9460/metrics?target=192.168.100.237:30015

```

also we can find the tracing details in jaeger http://192.168.1.1:16686(if all components of jaeger is installed on 192.168.1.1)


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

# Build

build the binary is pretty simple

```sh
git clone git@github.com:jenningsloy318/hana_exporter.git
cd hana_exporter
make build
```
# Parameter Explanation

 - --collect.sys_m_service_statistics, the metric hana_sys_m_service_statistics_status value and status mapping as following table:

    value | status |  
    ---------|---------- 
    0 | NO
    1 | YES
    2 | UNKNOWN
    3 |STARTING
    4 |STOPPING
 - --collect.sys_m_service_replication, the metric hana_sys_m_service_statistics_status value and status mapping as following table:
 
    value | status |  
    ---------|---------- 
    0 | ERROR
    1 | ACTIVE
    2 | UNKNOWN
    3 | INITIALIZING
    4 | SYNCING