#!/bin/bash


if ! grep "^prometheus:" /etc/group &>/dev/null; then
    groupadd -r prometheus
fi

if ! id hana_exporter &>/dev/null; then
    useradd -r -M hana_exporter -s /bin/false -d /etc/hana_exporter -g prometheus
fi


systemctl enable hana_exporter || true
systemctl daemon-reload || true