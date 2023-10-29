package plugin

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

type TcpConfig struct {
	Ip   string `json:"ip"`
	Port int    `json:"port"`
}

type TcpServer struct {
	Listener        net.Listener
	Config          TcpConfig
	MessageCallback TcpMessageCallback
	Clients         map[string]*TcpServerClient
}

type TcpMessageCallback func(clientid *string, message []byte)

func NewTcpServer(config TcpConfig, message_callback TcpMessageCallback) (tcp_server TcpServer) {

	tcp_server.Config = config
	tcp_server.Clients = map[string]*TcpServerClient{}
	tcp_server.MessageCallback = message_callback

	return
}

func (tcp_server *TcpServer) accept() {

	var (
		err          error
		conn         net.Conn
		client_timer int64
	)

	defer tcp_server.Listener.Close()

	for {
		if conn, err = tcp_server.Listener.Accept(); err != nil {
			continue
		}

		log.Printf("Client connect: %s\n", conn.RemoteAddr().String())

		client_timer++

		new_server_client := TcpServerClient{
			Clientid:   "",
			Conn:       conn,
			ActiveTime: time.Now().Unix(),
			Addr:       conn.RemoteAddr().String(),
		}

		tcp_server.Clients[fmt.Sprintf("%d%d", time.Now().UnixMicro(), client_timer)] = &new_server_client

		if tcp_server.MessageCallback != nil {
			go new_server_client.accept(tcp_server, tcp_server.MessageCallback)
		}
	}
}

func (tcp_server *TcpServer) client_close(server_client *TcpServerClient) {

	log.Printf("Client disconnect: %s\n", server_client.Addr)

	server_client.Conn.Close()

	for index := range tcp_server.Clients {
		if tcp_server.Clients[index] == server_client {
			delete(tcp_server.Clients, index)
		}
	}
}

func (tcp_server *TcpServer) Listen() (err error) {

	if tcp_server.Listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", tcp_server.Config.Ip, tcp_server.Config.Port)); err != nil {
		return
	}

	go tcp_server.accept()

	return
}

func (tcp_server *TcpServer) Exist(clientid string) *TcpServerClient {

	for _, server_client := range tcp_server.Clients {
		if server_client.Clientid == clientid {
			return server_client
		}
	}

	return nil
}

func (tcp_server *TcpServer) Send(clientid string, message []byte) (length int, err error) {

	for _, server_client := range tcp_server.Clients {
		if server_client.Clientid == clientid {
			length, err = server_client.Conn.Write(message)
		}
	}

	return
}

func (tcp_server *TcpServer) Read(clientid string, timeout int64) (data []byte, err error) {

	data = []byte{}

	server_client := (*TcpServerClient)(nil)

	for index := range tcp_server.Clients {
		if tcp_server.Clients[index].Clientid == clientid {
			server_client = tcp_server.Clients[index]
		}
	}

	if server_client == nil {
		err = errors.New("device_id no exist")
		return
	}

	length, buffer := 0, make([]byte, 1024)

READ:
	server_client.Conn.SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Millisecond))

	length, err = server_client.Conn.Read(buffer)

	if err != nil && strings.Contains(err.Error(), "i/o timeout") {
		return
	}

	if err != nil {
		tcp_server.client_close(server_client)
		return
	}

	if length == 0 && len(data) == 0 {
		goto END
	}

	data = append(data, buffer[:length]...)

	if length == len(buffer) {
		goto READ
	}

	server_client.ActiveTime = time.Now().Unix()

END:
	server_client.Conn.SetReadDeadline(time.Time{})
	return
}

type TcpServerClient struct {
	Clientid   string
	ActiveTime int64
	Addr       string
	Conn       net.Conn
}

func (server_client *TcpServerClient) accept(tcp_server *TcpServer, message_callback TcpMessageCallback) {

	defer tcp_server.client_close(server_client)

	buffer, data := make([]byte, 5), []byte{}

	for {
		length, err := server_client.Conn.Read(buffer)

		if err != nil {
			return
		}

		if length == 0 && len(data) == 0 {
			continue
		}

		data = append(data, buffer[:length]...)

		if length == len(buffer) {
			continue
		}

		server_client.ActiveTime = time.Now().Unix()

		message_callback(&server_client.Clientid, data)

		data = []byte{}
	}
}
