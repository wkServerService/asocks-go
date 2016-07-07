package main

import (
    "fmt"
    "net"
    "runtime"
    "flag"
    "os"
    "github.com/wkServerService/asocks-go/src/asocks"
    "io"
    "golang.org/x/net/websocket"
)

var localAddr string
var ws_origin string
var ws_url string

func main() {
    flag.StringVar(&localAddr, "l", "127.0.0.1:17000", "本地监听IP:端口")
    //flag.StringVar(&ws_origin, "o", "https://getdatas.herokuapp.com/", "websocket origin")
    //flag.StringVar(&ws_url, "u", "ws://getdatas.herokuapp.com/ws", "websocket url")
    flag.StringVar(&ws_origin, "o", "https://127.0.0.1:8080/", "websocket origin")
    flag.StringVar(&ws_url, "u", "ws://127.0.0.1:8080/ws", "websocket url")
    flag.Parse()

    numCPU := runtime.NumCPU()
    runtime.GOMAXPROCS(numCPU)

    bindAddr, _ := net.ResolveTCPAddr("tcp", localAddr)
    ln, err := net.ListenTCP("tcp", bindAddr)
    if  err != nil {
        fmt.Println("listen error:", err)
        return
    }
    defer ln.Close()

    fmt.Println("listening ", ln.Addr())
    //fmt.Println("ws_origin:", ws_origin)
    //fmt.Println("ws_url:", ws_url)

    for {
        conn, err := ln.AcceptTCP()
        if err != nil {
            fmt.Println("accept error:", err)
            continue
        }

        go handleConnection(conn)
    }
}

func handleConnection(conn *net.TCPConn) {
    closed := false

    defer func(){
        if !closed {
            conn.Close()
        }
    }()

    var n int
    var err error
    buf := make([]byte, 257)

    if n, err = io.ReadAtLeast(conn, buf, 2); err != nil {
        fmt.Println(err)
        return
    }

    if buf[0] != 5 {
        fmt.Printf("握手时读取到的socks版本是%d。\n", buf[0])
        return
    }

    nmethods := int(buf[1]) 
    if n < nmethods + 2 {
        if _, err = io.ReadFull(conn, buf[n : nmethods + 2]); err != nil {
            fmt.Println(err)
            return
        }
    }

    if _, err = conn.Write([]byte{5, 0}); err != nil {
        fmt.Println("握手时，写失败。err:", err)
        return
    }
 
    err = getRequest(conn)
    if err != nil {
        fmt.Println("get request err:", err.Error())
        return
    }

    closed = true
    return
}

func getRequest(conn *net.TCPConn) (err error) {
    var n int
    buf := make([]byte, 260)

    if n, err = io.ReadAtLeast(conn, buf, 5); err != nil {
        return
    }

    if buf[0] != 5 && buf[1] != 1 && buf[2] != 0 {
        err = fmt.Errorf("getRequest not socks5 protocol.")
        return
    }
    var reqLen = 0;
    var host string;
    switch buf[3] {
        case 1:
            // ipv4
            reqLen = 4 + 4 + 2
            host = net.IP(buf[4:8]).String()
        case 3:
            // domain
            domainLen := int(buf[4])
            reqLen = 4 + 1 + domainLen + 2
            dstAddr := buf[5 : 5 + int(buf[4])]
            host = string(dstAddr)
        case 4:
            // ipv6
            reqLen = 4 + 16 + 2
            host = net.IP(buf[4:20]).String()
        default:
            // unnormal, close conn
            err = fmt.Errorf("request type不正确：%d", buf[3])
            return
    }

    fmt.Println("Host:", host)

    if n < reqLen {
        if _, err = io.ReadFull(conn, buf[n : reqLen]); err != nil {
            return
        }
    }

    rawRequest := make([]byte, reqLen - 3)
    copy(rawRequest, buf[3:n])
   
    var remote *websocket.Conn
    if remote, err = websocket.Dial(ws_url, "", ws_origin); err != nil {
        return
    }
   
    asocks.EncodeData(rawRequest) 
    if n, err = remote.Write(rawRequest); err != nil {
        fmt.Println("getRequest 发送request到远程服务器失败。err:", err)
        remote.Close()
        return
    }

    // send reply
    var replyBuf []byte = []byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0}
    if _, err = conn.Write(replyBuf); err != nil {
        fmt.Println("getRequest 发送reply给客户端失败。err:", err) 
        remote.Close()
        return 
    }
   
    finishChannel := make(chan bool, 2) 
    go asocks.PipeThenClose(conn, remote, finishChannel)
    asocks.PipeThenClose(remote, conn, finishChannel)
    <- finishChannel
    <- finishChannel
    conn.Close()
    remote.Close()

    return nil
}

func printUsage() {
    fmt.Printf("Usage:%s -s server_addr:server_port -l local_addr:local_port\n", os.Args[0])
}
