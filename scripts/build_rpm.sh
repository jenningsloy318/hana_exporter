#!/bin/bash -x

export TOP_DIR=$(mktemp -d)
mkdir ${TOP_DIR}/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
SCRIPTPATH=$(dirname "$0")
cp   ${SCRIPTPATH}/../build/hana_exporter ${TOP_DIR}/SOURCES
cp   ${SCRIPTPATH}/hana_exporter.service ${TOP_DIR}/SOURCES
cp   ${SCRIPTPATH}/hana_exporter.yml ${TOP_DIR}/SOURCES
cp   ${SCRIPTPATH}/hana_exporter.spec ${TOP_DIR}/SPECS
rpmbuild --define "_topdir ${TOP_DIR}" -bb ${TOP_DIR}/SPECS/hana_exporter.spec
cp -f ${TOP_DIR}/RPMS/x86_64/hana_exporter*.rpm   ${SCRIPTPATH}/../build

rm -rf ${TOP_DIR}