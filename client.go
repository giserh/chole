package main

import (
	"net"
)

type Client struct {
	server string
	name   string
	in     string
	out    string
	proxys []*Proxy
	manage *ManageClient
}

func (client *Client) connect(addr string) net.Conn {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		Error("CLIENT", err)
		return nil
	}
	return conn
}

func (client *Client) newConnect(conn net.Conn) chan bool {
	fromConn := client.connect(client.server + ":" + PROXY_SERVER_PORT)
	toConn := client.connect(client.in)
	if fromConn == nil || toConn == nil {
		TryClose(fromConn)
		TryClose(toConn)
		return nil
	}
	proxy := Proxy{
		from: fromConn,
		to:   toConn,
		init: func(fromConn net.Conn) {
			remoteAddr := client.manage.remoteAddr
			SendPacket(fromConn, "REQUEST_PROXY", remoteAddr)
		},
		valid: func(data []byte) bool {
			// domain := ParseDomain(data)
			// log.Println(domain)
			return true
		},
	}
	client.proxys = append(client.proxys, &proxy)
	return proxy.Start(false)
}

func (client *Client) Close() {
	TryClose(client.manage.conn)
	for _, proxy := range client.proxys {
		proxy.Close()
	}
	client.proxys = client.proxys[:0]
}

func (client *Client) Start() chan bool {
	status := make(chan bool)
	manage := ManageClient{
		server: client.server + ":" + MANAGER_SERVER_PORT,
		onConnect: func(conn net.Conn) {
			SendPacket(conn, "REQUEST_PORT", client.out)
		},
		onEvent: func(conn net.Conn, event string, data string) {
			switch event {
			case "REQUEST_COMING":
				<-client.newConnect(conn)
				break
			case "REQUEST_PORT_ACCEPT":
				client.manage.remoteAddr = data
				status <- true
				break
			case "REQUEST_PORT_REJECT":
				Error("CLIENT", "requested port "+client.out+" has been used")
				conn.Close()
				status <- false
				break
			}
		},
	}
	client.proxys = make([]*Proxy, 0)
	client.manage = &manage
	<-manage.Start()
	return status
}
