#!/bin/bash
cat << EOF > ${PWD}/iris-exporter.service
[Unit]
Description=Prometheus iris exporter
ConditionPathExists=${HOME}/iris-exporter/iris-exporter
After=network.target

[Service]
Type=simple
LimitNOFILE=1024
Restart=on-failure
RestartSec=10
AmbientCapabilities=CAP_SYS_CHROOT
User=iris
Group=iris
RootDirectoryStartOnly=true
WorkingDirectory=${HOME}
ExecStart=/bin/bash -c 'source ${HOME}/.bash_profile && ${PWD}/iris-exporter -masterIP=192.168.10.76 -listen=:9102 -irisBinPath=${HOME}/IRIS/bin/'
PermissionsStartOnly=true
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=iris-exporter

[Install]
WantedBy=multi-user.target
EOF