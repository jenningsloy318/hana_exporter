# hana_exporter 

this is initial hana exporter to monitor hana database.

gather all hana info from its system table/views, all table/veiw under `SYS` schema with `M_` as prefix are monitoring tables/views. so we just fetch data from `SYS.M_*` tables/views.

 
the exporter itself metrics exposed at `/metrics`, and the hana database metrics exposed at `/hana`
# Usage 
create a configuration `hana.yml`, which contains the credentials of hana instance.
```yaml
credentials:
    default:
        user: "user"
        pass: "password"
    192.168.100.237:30015:
        user: "SYSTEM"
        pass: "Password"
```

Note: This user should have role `SAP_INTERNAL_HANA_SUPPORT` to access schema `SYS`, Without the `SAP_INTERNAL_HANA_SUPPORT` role, this information can be selected only by the `SYSTEM` user.

```sql
grant  SAP_INTERNAL_HANA_SUPPORT  to user
```

then start hana-exporter via 
```sh
hana_exporter --config.file=hana_exporter.yml
```

then we can get the metrics via 
```
curl http://<hana-export host>:9460/hana?target=192.168.100.237:30015

```

## NOTE: The usre configured at lest have `select` permission on schema `SYS`, all the collector will collect the info from tables/views under this schema.

## prometheus job conf
add hana-exporter job conif as following
```yaml
  - job_name: 'hana-exporter'

    # metrics_path defaults to '/metrics'
    metrics_path: /hana


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
or package it as rpm or deb
```
make package-release
```

then you package can be found in `./build`


# TO DO

also need to implement following metrics:
- metrics about user, user count, expired user, be expiring user, locked user  from "SYS"."USERS"
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
