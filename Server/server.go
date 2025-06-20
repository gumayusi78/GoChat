package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	maplock   sync.RWMutex
	Message   chan string
}

// NewServer 创建一个server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// BroadCast 广播消息
func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := user.Username + ": " + msg
	s.Message <- sendMsg
}

// 监听Message广播消息
func (s *Server) ListenMessage() {
	for {
		msg := <-s.Message
		s.maplock.RLock()
		for _, cli := range s.OnlineMap {
			cli.Channel <- msg
		}
		s.maplock.RUnlock()
	}
}

func (s *Server) Handler(conn net.Conn) {

	user := NewUser(conn, s)

	user.Online() //用户上线

	//监听用户是否活跃的Channel
	isActive := make(chan bool)

	//处理用户发送的消息
	go func() {
		buf := make([]byte, 4096) // 创建一个4096字节的缓冲区，用于临时存储从连接读取的数据
		for {
			n, err := conn.Read(buf) // 从连接中读取数据到buf缓冲区
			if n == 0 {
				user.Offline() //用户下线
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("conn.Read err:", err)
				return
			}
			//处理用户发送的消息去除\n
			msg := string(buf[:n-1]) // 将缓冲区数据转为字符串并去除末尾的换行符

			user.DoMsg(msg) // 处理用户发送的消息
			isActive <- true
		}
	}()

	for {
		select {
		case <-isActive:
			//用户活跃,重置定时器
			//不做任何操作，为了激活select，更新定时器
		case <-time.After(300 * time.Second):
			//超时
			user.SendMsg("五分钟未做任何操作，你已被踢下线")
			//关闭连接
			close(user.Channel)
			conn.Close()
			return

		}
	}
}

//启动服务器接口

func (s *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println("listener.Close err:", err)
		}
	}(listener)

	go s.ListenMessage()

	//accept
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept err:", err)
			continue
		}
		//do handler
		go s.Handler(conn)
	}

	//close
}
