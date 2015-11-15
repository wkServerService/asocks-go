package main

import (
    "fmt"
    "net"
    "io"
)

func handleConn(conn *net.TCPConn, raddr *net.TCPAddr) {
    remote, err := net.DialTCP("tcp", nil, raddr)
    if err != nil {
        fmt.Println("dial remote fail. err:", err)
        conn.Close()
        return
    }
    
    finish := make(chan bool, 2)
    go pipe(conn, remote, finish)
    pipe(remote, conn, finish)
    <- finish
    <- finish
    conn.Close()
    remote.Close()
}

func pipe(src, dst *net.TCPConn, finish chan bool) {
    defer func() {
        //src.CloseRead()
        //dst.CloseWrite()
        finish <- true
    }()

    io.Copy(dst, src)
}

var laddr string = ":17570" 
var raddr string = "actself.me:17570" 

func main() {
    a, _ := net.ResolveTCPAddr("tcp", laddr) 
    b, _ := net.ResolveTCPAddr("tcp", raddr)
    
    ln, err := net.ListenTCP("tcp", a) 
    if err != nil {
        fmt.Println("listen error:", err)
        return;
    }
    defer ln.Close()
    fmt.Println("listening on ", ln.Addr())

    for {
        conn, err := ln.AcceptTCP()
        if err != nil {
            fmt.Println("accept err:", err)
            continue
        }

        go handleConn(conn, b)
    }
}
