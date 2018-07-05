# hana_exporter 

this is initial hana exporter to monitor hana database.

gather all hana info from its system table/views, following the description at 

http://sap.optimieren.de/hana/hana/html/sys_statistics_views.html

# Usage 
create a configuration `hana-exporter.yml`
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