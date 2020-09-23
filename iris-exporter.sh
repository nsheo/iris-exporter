#!/bin/bash
cat << EOF > ${PWD}/iris-exporter.service
[Unit]
Description=Prometheus iris exporter
ConditionPathExists=${PWD}/iris-exporter
After=network.target

[Service]
Type=simple
LimitNOFILE=1024
Restart=on-failure
RestartSec=10
WorkingDirectory=${PWD}
ExecStart=${PWD}/iris-exporter -sedfile=${PWD}/sedcommand.file
PermissionsStartOnly=true
ExecStartPre=/bin/mkdir -p /var/log/iris-exporter
ExecStartPre=/bin/chmod 755 /var/log/iris-exporter
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=iris-exporter

[Install]
WantedBy=multi-user.target
EOF