package main

import (
    "fmt"
    "net"
    "strconv"
    "encoding/binary"
    "runtime"
    "flag"
    "github.com/wkServerService/asocks-go/src/asocks"
    "io"
    "net/http"
    "golang.org/x/net/websocket"
    "log"
)



func main() {
    var localAddr string
    flag.StringVar(&localAddr, "l", ":8080", "监听端口")
    flag.Parse()

    numCPU := runtime.NumCPU()
    runtime.GOMAXPROCS(numCPU)

    http.Handle("/ws", websocket.Handler(handleConnection))
    http.HandleFunc("/", handleHttp)
    log.Println("listened on:", localAddr)
    err := http.ListenAndServe(localAddr, nil)
    if err != nil {
        panic("ListenAndServe: " + err.Error())
    }
}
func handleHttp(w http.ResponseWriter, r *http.Request)  {
    w.Write([]byte("hello world!"))
}

func handleConnection(conn *websocket.Conn) {
    err := getRequest(conn)

    if err != nil {
        fmt.Println("err:", err)
        conn.Close()
    }
}

func getRequest(conn *websocket.Conn) (err error){
    var n int
    buf := make([]byte, 257)

    if n, err = io.ReadAtLeast(conn, buf, 2); err != nil {
        return
    }

    asocks.EncodeData(buf)

    addressType := buf[0]
    reqLen := 0;

    var host string;
    switch addressType {
        case 1:
            // ipv4
            reqLen = 1 + 4 + 2
            host = net.IP(buf[1:5]).String()
        case 3:
            // domain
            reqLen = 1 + 1 + int(buf[1]) + 2
            dstAddr := buf[2 : 2 + int(buf[1])]
            host = string(dstAddr)
        case 4:
            // ipv6
            reqLen = 1 + 16 + 2
            host = net.IP(buf[1:17]).String()
        default:
            // unnormal, close conn
            err = fmt.Errorf("error ATYP:%d\n", buf[0])
            return
    }
    
    if n < reqLen {
        if _, err = io.ReadFull(conn, buf[n : reqLen]); err != nil {
            return
        }
        asocks.EncodeData(buf[n:reqLen]) 
    }

    port := binary.BigEndian.Uint16(buf[reqLen - 2 : reqLen])
    host = net.JoinHostPort(host, strconv.Itoa(int(port)))

    fmt.Println("dst:", host)

    var remote *net.TCPConn
    remoteAddr, _ := net.ResolveTCPAddr("tcp", host)
    if remote, err = net.DialTCP("tcp", nil, remoteAddr); err != nil {
        return
    }
    
    // 如果有额外的数据，转发给remote。正常情况下是没有额外数据的，但如果客户端通过端口转发连接服务端，就会有
    if n > reqLen {
        if _, err = remote.Write(buf[reqLen : n]); err != nil {
            return
        }
    }
   
    finish := make(chan bool, 2) 
    go asocks.PipeThenClose(conn, remote, finish)
    asocks.PipeThenClose(remote, conn, finish)
    <- finish
    <- finish
    conn.Close()
    remote.Close()

    return nil
}
