go build -o build/proxy-client-64-ws.exe src/asocks_local/local.go
CGO_ENABLED=0 GOOS=windows GOARCH=386  go build -o build/proxy-client-32-ws.exe src/asocks_local/local.go
CGO_ENABLED=0 GOOS=linux GOARCH=armv6  go build -o build/proxy-client-armv6l-linux-ws src/asocks_local/local.go
go build -o build/proxy-server-64-win-ws.exe src/asocks_server/server.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build -o build/proxy-server-64-linux-ws src/asocks_server/server.go