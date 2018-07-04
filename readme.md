# hana_exporter 

this is initial hana exporter to monitor hana database.

gather all hana info from its system table/views, following the description at 

http://sap.optimieren.de/hana/hana/html/sys_statistics_views.html

# Usage 
There are 3 method to pass HANA Database to hana_exporter 

1. set env `HANA_DATA_SOURCE_NAME`  
  ```sh
  export HANA_DATA_SOURCE_NAME=hdb://user:password@hana-host:hana-port
  ```

2. set for env 
  ```sh
 export HANA_HOST=host
 export HANA_PORT=port
 export HANA_USER=user
 export HANA_PASSWORD=pass
 ```

3. create a conf file and pass to hana_exporter via flag `-config.my-cnf`
```conf
[client]
user = user
password = password
host = host
port = port 
```

