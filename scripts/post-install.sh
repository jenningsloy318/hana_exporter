#!/bin/bash


if  ! getent group prometheus >/dev/null  ; then
    groupadd -r prometheus
fi

if  ! getent passwd hana_exporter  >/dev/null  ; then
    useradd -r -M -s /bin/false -d /etc/hana_exporter -g prometheus hana_exporter
fi

chown -R hana_exporter /etc/hana_exporter
chgrp -R prometheus /etc/hana_exporter
systemctl daemon-reload || true
systemctl enable hana_exporter || true
systemctl start hana_exporter
