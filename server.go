package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	OnlineMap map[string]*User
	MapLock   sync.RWMutex
	Message   chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

// 连接建立，Message接收用户上线消息
func (this *Server) Handler(conn net.Conn) {
	fmt.Println("连接建立成功")

	user := NewUser(conn, this)

	// 添加用户信息到map
	user.Online()

	// 用于实现超时强踢
	isLive := make(chan bool)

	// 接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)

		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				// this.Broadcast(user, user.Name+" off line")
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("conn read error: ", err)
				return
			}

			// 提取消息
			msg := string(buf[0 : n-1])

			// 处理消息
			// this.Broadcast(user, msg)
			user.DoProcessMsg(msg)

			isLive <- true
		}
	}()

	// 阻塞当前handler,添加定时器功能
	for {
		select {
		case <-isLive:
			// 用户活跃，刷新定时器

		case <-time.After(300 * time.Second):
			// 已超时
			user.SendMsg("你被踢了")

			close(user.C)

			conn.Close()

			this.MapLock.Lock()
			delete(this.OnlineMap, user.Name)
			this.MapLock.Unlock()

			return
		}
	}

}

// Message广播用户上线消息
func (this *Server) Broadcast(user *User, msg string) {
	msg = "user: [" + user.Name + "], msg: " + msg
	this.Message <- msg

}

// 监听Message，广播给User
func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message

		this.MapLock.Lock()
		for _, user := range this.OnlineMap {
			user.C <- msg
		}
		this.MapLock.Unlock()
	}
}

func (this *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))

	if err != nil {
		fmt.Println("Listen error")
		return
	}

	defer listener.Close()

	// Message广播用户上线
	go this.ListenMessage()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Accept error")
			continue
		}

		go this.Handler(conn)
	}

}
