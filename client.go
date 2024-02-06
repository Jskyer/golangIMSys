package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(ip string, port int) *Client {
	client := &Client{
		ServerIp:   ip,
		ServerPort: port,
		flag:       999,
	}

	// 连接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", client.ServerIp, client.ServerPort))
	if err != nil {
		fmt.Println("net dial error")
		return nil
	}

	client.conn = conn

	return client
}

func (this *Client) Menu() bool {
	var flag int
	fmt.Println("=====菜单模式选择=====")
	fmt.Println("=====1. 公聊模式=====")
	fmt.Println("=====2. 私聊模式=====")
	fmt.Println("=====3. 更新用户名=====")
	fmt.Println("=====0. 退出=====")

	_, err := fmt.Scanln(&flag)
	if err != nil {
		fmt.Println("Error: ", err)
		fmt.Println("输入不合法")
		return false
	}

	if flag >= 0 && flag <= 3 {
		this.flag = flag
		return true
	} else {
		fmt.Println("输入不合法")
		return false
	}
}

func (this *Client) Run() {
	for this.flag != 0 {
		for !this.Menu() {
		}

		switch this.flag {
		case 1:
			fmt.Println("进入公聊")
			this.PublicChat()
		case 2:
			fmt.Println("进入私聊")
			this.PrivateChat()
		case 3:
			fmt.Println("进入更新用户名")
			this.UpdataName()
		}
	}

	fmt.Println("退出菜单")
}

// flag = 1
func (this *Client) PublicChat() {
	var msg string = ""

	for {
		fmt.Println("====请输入消息, 输入exit退出====")
		// 读取msg
		fmt.Scanln(&msg)

		if len(msg) == 0 {
			fmt.Println("消息不可为空")
			continue
		}

		if msg == "exit" {
			break
		}

		// 发送msg给server
		msg += "\n"
		_, err := this.conn.Write([]byte(msg))
		if err != nil {
			fmt.Println("Write error:", err)
			break
		}

		fmt.Println("PublicChat send success")

		msg = ""
	}

	fmt.Println("退出公聊模式")
}

func (this *Client) SelectUsers() {
	msg := "who\n"
	_, err := this.conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("查询用户失败：", err)
		return
	}

}

// flag = 2
func (this *Client) PrivateChat() {
	var remoteName string = ""
	var chatMsg string = ""

	this.SelectUsers()
	for {
		fmt.Println("====请输入私聊用户, 输入exit退出====")
		// 读取user
		fmt.Scanln(&remoteName)

		if len(remoteName) == 0 {
			fmt.Println("私聊用户不可为空")
			continue
		}

		if remoteName == "exit" {
			break
		}

		for {
			fmt.Println("====请输入消息, 输入exit退出====")
			// 读取msg
			fmt.Scanln(&chatMsg)

			if len(chatMsg) == 0 {
				fmt.Println("消息不可为空")
				continue
			}

			if chatMsg == "exit" {
				break
			}

			// 发送msg给server
			msg := "to|" + remoteName + "|" + chatMsg + "\n"
			_, err := this.conn.Write([]byte(msg))
			if err != nil {
				fmt.Println("Write error:", err)
				break
			}

			chatMsg = ""

		}

		remoteName = ""
	}

	fmt.Println("退出私聊模式")
}

// flag = 3
func (this *Client) UpdataName() bool {
	fmt.Println("请输入用户名")
	fmt.Scanln(&this.Name)

	msg := "rename|" + this.Name + "\n"
	_, err := this.conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("Conn Write Error", err)
		return false
	}

	return true
}

func (this *Client) DealWithResponse() {
	io.Copy(os.Stdout, this.conn)
}

var ServerIp string
var ServerPort int

func init() {
	// ./client -ip 127.0.0.1 -port 8888
	flag.StringVar(&ServerIp, "ip", "127.0.0.1", "默认ip地址(127.0.0.1)")
	flag.IntVar(&ServerPort, "port", 8888, "默认端口号(8888)")
}

func main() {
	// 命令行解析
	flag.Parse()

	client := NewClient(ServerIp, ServerPort)
	if client == nil {
		fmt.Println("client connect error")
		return
	}

	fmt.Println("client connect success")

	// 开启一个goroutine，处理服务器响应
	go client.DealWithResponse()

	// 启动客户端业务
	client.Run()
	// select {}
}
