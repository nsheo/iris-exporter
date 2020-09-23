# iris-exporter   

simple Prometheus iris exporter   

need iris master node to run

running node, must have iris mps / ntop binary 

1. Create Build File   
```
cd ${GOPATH}/src
git clone https://gitlab.com/nsheo/iris-exporter.git
cd iris-exporter
go get
go build
```

2. Make Service for CentOS& or Systemd   
```
chmod 751 ./iris-exporter.sh
./iris-exporter.sh
cp iris-exporter.service /etc/systemd/system/
systemctl enable iris-exporter.service
systemctl start iris-exporter.service
```

3. Parameter for iris-exporter   
```
./iris-exporter -masterIP=<localip of run this exporter(0.0.0.0)> -listen=<metrics http port(default:9102) -irisBinPath=<iris uitl binary path(default:$M6_HOME/bin)> -sedfile=<sed filter file to filter subnode mps (default:sedcommand.file)> 
```

