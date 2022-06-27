#GOARCH=amd64 GOOS=freebsd go build &&  scp clash root@10.0.0.1:
GOARCH=amd64 GOOS=linux go build &&  scp clarity sma@192.168.182.2:bin
