go build -o server-test.exe asocks_server/server.go && ./server-test.exe
go build -o local-test.exe asocks_local/local.go && ./local-test.exe -s 127.0.0.1:8388 -l 127.0.0.1:16000