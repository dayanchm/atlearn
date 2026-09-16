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
