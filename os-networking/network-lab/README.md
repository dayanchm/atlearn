# net.Conn lifecycle

````
CLIENT                              SERVER

net.Dial()  ─────────────────────►  Accept()
    │                                  │
    ▼                                  ▼
 net.Conn                           net.Conn
    │                                  │
 Write() ─────────────────────────► Read()
    │                                  │
    │                              n + data
    │
 Close() ─────────────────────────► Read()
                                       │
                                     io.EOF
                                       │
                                     Close()
````


**net.Conn What was I learn ?**
 
Explain 

```
conn.Read()             -> Get Data
conn.Write()            -> Send Data
conn.Close()            -> Closed connection
conn.LocalAddr()        -> My address `ip`
conn.RemoteAddr()       -> The other address `ip`
conn.SetReadDeadline()  -> Read timeout
conn.SetWriteDeadline() -> Write timeout
```

net.Conn = established connection and connection lifecycle control in Go UI



`net.Conn = Manages the established connected `
`net.Dial = It manages how the connection is established. `