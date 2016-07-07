package asocks


import (
    "net"
    "time"
)

func PipeThenClose(src, dst net.Conn, finish chan bool) {
    defer func(){
        src.Close()
        dst.Close()
        finish <- true
    }()

    buf := GetBuffer()
    defer GiveBuffer(buf)

    for {
        src.SetReadDeadline(time.Now().Add(60 * time.Second))
        n, err := src.Read(buf);
        if n > 0 {
            data := buf[0:n]
            EncodeData(data)
            if _, err := dst.Write(data); err != nil {
                break
            }
        }
        if err != nil {
            break
        }
    }
}

func EncodeData(data []byte) {
    for i, _ := range data {
        data[i] ^= 100;
    }
}