package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	Conn net.Conn

	Server *Server
}

func (this *User) Online() {
	// 添加用户信息到map
	this.Server.MapLock.Lock()
	this.Server.OnlineMap[this.Name] = this
	this.Server.MapLock.Unlock()

	this.Server.Broadcast(this, "已上线")

	fmt.Println(this.Name, "online")
}

// 用户下线
func (this *User) Offline() {
	this.Server.MapLock.Lock()
	delete(this.Server.OnlineMap, this.Name)
	this.Server.MapLock.Unlock()

	this.Server.Broadcast(this, this.Name+" off line")

	fmt.Println(this.Name, "offline")
}

func (this *User) SendMsg(msg string) {
	this.Conn.Write([]byte(msg))
}

// 根据用户指令处理信息
func (this *User) DoProcessMsg(msg string) {
	if msg == "who" {
		// 查询在线用户列表
		this.Server.MapLock.Lock()
		for _, user := range this.Server.OnlineMap {
			infoMsg := "user: " + user.Name + "\n"
			this.SendMsg(infoMsg)
		}
		this.Server.MapLock.Unlock()

	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 重命名
		newName := strings.Split(msg, "|")[1]

		_, ok := this.Server.OnlineMap[newName]

		if ok {
			// 已存在newName
			this.SendMsg("已存在用户名，不可更改: " + newName + "\n")

		} else {
			this.Server.MapLock.Lock()
			delete(this.Server.OnlineMap, this.Name)
			this.Server.OnlineMap[newName] = this
			this.Server.MapLock.Unlock()

			this.Name = newName
			this.SendMsg("您已更新用户名\n")
		}

	} else if len(msg) > 4 && msg[:3] == "to|" {
		// 消息格式：to|zhangsan|okkk
		// 获取用户名
		strFormat := strings.Split(msg, "|")

		if len(strFormat) < 3 {
			this.SendMsg("发送给指定用户的格式错误, 请使用格式(\"to|zhangsan|okkk\")\n")
			return
		}

		remoteUserName := strFormat[1]
		if remoteUserName == "" {
			this.SendMsg("接收方用户名不可为空\n")
			return
		}

		// 获取user对象
		remoteUser, ok := this.Server.OnlineMap[remoteUserName]
		if !ok {
			this.SendMsg("接收方用户名不存在\n")
			return
		}

		// 获取消息
		toMsg := strFormat[2]
		if toMsg == "" {
			this.SendMsg("发送消息不能为空\n")
			return
		}

		// 实现私聊,this: 发送方
		remoteUser.SendMsg("user: " + this.Name + ", 和你说: " + toMsg + "\n")

	} else {
		// 广播消息
		this.Server.Broadcast(this, msg)
	}
}

// 监听channel，有消息就发送给对应的client
func (this *User) SendClient() {
	for {
		msg := <-this.C

		this.Conn.Write([]byte(msg + "\n"))
	}
}

func NewUser(conn net.Conn, server *Server) *User {
	addr := conn.RemoteAddr().String()
	name := addr

	user := &User{
		Name:   name,
		Addr:   addr,
		C:      make(chan string),
		Conn:   conn,
		Server: server,
	}

	go user.SendClient()

	return user
}
