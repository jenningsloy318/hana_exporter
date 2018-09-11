
#!/bin/bash

if -f /etc/hana_exporter/hana_exporter.yml ; then
    backup_name="hana_exporter.yml.$(date +%s).backup"
    echo "A backup of your current configuration can be found at: /etc/hana_exporter/${backup_name}"
    cp -a "/etc/hana_exporter/hana_exporter.yml" "/etc/hana_exporter/${backup_name}"
fi
