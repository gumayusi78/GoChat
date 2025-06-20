package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	// 添加重试逻辑
	var conn net.Conn
	var err error
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		// 连接服务器
		conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
		if err == nil {
			break
		}
		fmt.Printf("连接尝试 %d 失败: %v\n", i+1, err)
		time.Sleep(time.Second * 2) // 等待2秒后重试
	}

	if err != nil {
		fmt.Println("最终连接失败:", err)
		return nil
	}

	client.conn = conn
	return client
}

// 处理服务器返回的消息
func (c *Client) DealResponse() {
	io.Copy(os.Stdout, c.conn)
	//for{
	//	buf := make([]byte, 1024)
	//	n, err := c.conn.Read(buf)
	//	if err != nil {
	//		fmt.Println("服务器已断开")
	//		break
	//	}
	//	fmt.Println(string(buf[:n]))
	//}
}

func (c *Client) Run() {
	for c.flag != 0 {
		for c.ShowMenu() != true {
		}
		// 根据flag的值，执行不同的业务
		switch c.flag {
		case 1:
			c.PublicChat()
		case 2:
			c.PrivateChat()
		case 3:
			c.UpdateName()
		case 4:
			c.flag = 0
		}
	}
}

// 公聊模式
func (c *Client) PublicChat() {
	fmt.Println("请输入聊天内容,exit退出")
	var content string
	fmt.Scanln(&content)
	for content != "exit" {
		if len(content) > 0 {
			// 发送消息
			sendMsg := content + "\n" //协议中需要换行
			_, err := c.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("发送失败:", err)
				return
			}
		}

		content = ""
		fmt.Println("请输入聊天内容,exit退出")
		fmt.Scanln(&content)

	}
}

// 查询在线用户
func (c *Client) QueryUsers() {
	// 发送消息
	sendMsg := "query\n"
	_, err := c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("发送失败:", err)
		return
	}
}

// 私聊模式
func (c *Client) PrivateChat() {
	var name string
	var content string
	c.QueryUsers()
	fmt.Println("请输入对方用户名,exit退出")
	fmt.Scanln(&name)
	for name != "exit" {
		fmt.Println("请输入消息内容,exit退出")
		fmt.Scanln(&content)
		for content != "exit" {
			if len(content) > 0 {
				// 发送消息
				sendMsg := "to|" + name + "|" + content + "\n\n"
				_, err := c.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("发送失败:", err)
					return
				}
			}
			content = ""
			fmt.Println("请输入消息内容,exit退出")
			fmt.Scanln(&content)
		}
		// 内层循环结束后，重新询问用户名
		c.QueryUsers()
		fmt.Println("请输入对方用户名,exit退出")
		fmt.Scanln(&name)
	}
}

// 更新用户名
func (c *Client) UpdateName() {
	fmt.Println("请输入新的用户名:")
	fmt.Scanln(&c.Name)
	// 发送消息
	sendMsg := "rename|" + c.Name + "\n"
	_, err := c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("发送失败:", err)
		return
	}
}

// 菜单显示
func (c *Client) ShowMenu() bool {
	var flag int
	fmt.Println("1. 公聊模式")
	fmt.Println("2. 私聊模式")
	fmt.Println("3. 更新用户名")
	fmt.Println("4. 退出系统")
	fmt.Println("请选择(1-4):")
	fmt.Scanln(&flag)
	if flag >= 1 && flag <= 4 {
		c.flag = flag
		return true
	} else {
		fmt.Println("输入有误，请重新输入")
		return false
	}
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "服务器ip地址")
	flag.IntVar(&serverPort, "port", 8090, "服务器端口")
}

func main() {
	flag.Parse()
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("客户端创建失败")
		return
	}
	//单独起一个goroutine去处理服务器的响应
	go client.DealResponse()
	fmt.Println(">>>>>连接服务器成功...")

	// 启动客户端的业务
	client.Run()

}
