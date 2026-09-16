# net.Conn lifecycle

**net.Conn is the fundamental abstraction of a network connection established in GO**


## Basic Licecycle 
```
Dial
 ↓
connection established
 ↓
Read / Write
 ↓
Deadlines / errors / cancellation
 ↓
Close
 ↓
connection resources released
```

### For exmample

```
conn, err := net.Dial("tcp", "example.com:80")
if err != nil {
    return err
}
defer conn.Close()

_, err = conn.Write([]byte(
    "GET / HTTP/1.1\r\nHost: example.com\r\n\r\n",
))
if err != nil {
    return err
}

buf := make([]byte, 4096)

n, err := conn.Read(buf)
if err != nil {
    return err
}

fmt.Println(string(buf[:n]))
```

## Production at lifecycle

```
        Dial
          │
          ▼
     CONNECTING
          │
          ▼
     ESTABLISHED
       │      │
       │      │
     Read    Write
       │      │
       └──┬───┘
          │
     deadline?
          │
     peer closed?
          │
     local close?
          │
          ▼
       CLOSING
          │
          ▼
        CLOSED
```
