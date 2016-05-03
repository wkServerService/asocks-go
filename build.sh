go build -o build/proxy-client-64.exe src/asocks_local/local.go
CGO_ENABLED=0 GOOS=windows GOARCH=386  go build -o build/proxy-client-32.exe src/asocks_local/local.go
go build -o build/proxy-server-64-win.exe src/asocks_local/local.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build -o build/proxy-server-64-linux src/asocks_local/local.go