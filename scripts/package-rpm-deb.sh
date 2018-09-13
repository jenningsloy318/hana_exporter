#!/bin/bash

REPOROOT=$(dirname $(readlink -f $0))/../
VER=$(cat ${REPOROOT}/VERSION)
echo "Package hana exporter for version ${VER}"
TMPDIR=$(mktemp -d -p /tmp)
echo "Copy all files to  temp dir ${TMPDIR}"

mkdir -p  ${TMPDIR}/usr/bin 
cp ${REPOROOT}/build/hana_exporter ${TMPDIR}/usr/bin/
mkdir -p  ${TMPDIR}/usr/lib/systemd/system 

cp ${REPOROOT}/scripts/hana_exporter.service ${TMPDIR}/usr/lib/systemd/system/

mkdir -p ${TMPDIR}/etc/hana_exporter

cp ${REPOROOT}/example/hana_exporter.yml ${TMPDIR}/etc/hana_exporter/

echo "build deb package"

fpm -f -s dir --log error  --vendor jenningsloy318  --url https://github.com/jenningsloy318/hana_exporter --maintainer jenningsloy318@gmail.com --config-files  /etc/hana_exporter/hana_exporter.yml   --after-install ${REPOROOT}/scripts/post-install.sh  --before-install ${REPOROOT}/scripts/pre-install.sh  --after-remove ${REPOROOT}/scripts/post-remove.sh  --before-remove ${REPOROOT}/scripts/pre-remove.sh  --description "hana exporter to monitor hana database." --name hana_exporter -a amd64 -t deb --version ${VER} --iteration 1 -C ${TMPDIR} -p ${REPOROOT}/build

echo "build rpm package"
fpm -f -s dir --log error  --vendor jenningsloy318  --url https://github.com/jenningsloy318/hana_exporter  --maintainer jenningsloy318@gmail.com   --config-files /etc/hana_exporter/hana_exporter.yml  --after-install ${REPOROOT}/scripts/post-install.sh  --before-install ${REPOROOT}/scripts/pre-install.sh  --after-remove ${REPOROOT}/scripts/post-remove.sh  --before-remove ${REPOROOT}/scripts/pre-remove.sh  --description "hana exporter to monitor hana database." --name hana_exporter -a amd64 -t rpm --version ${VER} --iteration 1 -C ${TMPDIR} -p ${REPOROOT}/build


echo "clear tmp directory"

#rm -rf ${TMPDIR}
