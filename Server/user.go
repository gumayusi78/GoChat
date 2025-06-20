package main

import (
	"net"
	"strings"
)

type User struct {
	Username string
	Address  string
	Channel  chan string
	conn     net.Conn
	server   *Server
}

// 创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	user := &User{
		Username: conn.RemoteAddr().String(),
		Address:  conn.RemoteAddr().String(),
		Channel:  make(chan string),
		conn:     conn,
		server:   server,
	}
	go user.ListenMessage()
	return user
}

func (u *User) Online() {
	// 广播当前用户上线
	// server.BroadCast(u, "已上线")
	//用户上线
	u.server.maplock.Lock()
	u.server.OnlineMap[u.Username] = u
	u.server.maplock.Unlock()

	//广播当前用户上线
	u.server.BroadCast(u, "已上线")
}

func (u *User) Offline() {
	// 广播当前用户下线
	// server.BroadCast(u, "已下线")

	//用户下线
	u.server.maplock.Lock()
	delete(u.server.OnlineMap, u.Username)
	u.server.maplock.Unlock()

	//广播当前用户下线
	u.server.BroadCast(u, "已下线")
}

func (u *User) ReName(newName string) {
	// 用户修改用户名
	u.server.maplock.Lock()
	delete(u.server.OnlineMap, u.Username)
	u.Username = newName
	u.server.OnlineMap[newName] = u
	u.server.maplock.Unlock()

	// 只向当前用户发送修改成功的消息
	u.SendMsg("您已修改用户名为：" + newName)
}

func (u *User) SendMsg(msg string) {
	u.conn.Write([]byte(msg + "\n"))
}

// 用户消息处理业务
func (u *User) DoMsg(msg string) {

	msg = strings.TrimSpace(msg)
	if msg == "query" {
		//查询在线用户
		u.server.maplock.RLock()
		for _, user := range u.server.OnlineMap {
			onlineMsg := user.Username + ": " + "在线..."
			u.SendMsg(onlineMsg)
		}
		u.server.maplock.RUnlock()
		return
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//修改用户名
		// 使用Split方法以"|"为分隔符分割字符串，并取第二个元素作为新用户名
		newName := strings.Split(msg, "|")[1]
		//防止名字重复
		_, ok := u.server.OnlineMap[newName]
		if ok {
			u.SendMsg("当前用户名已存在")
			return
		} else {
			u.ReName(newName)
			u.SendMsg("修改用户名成功")
		}
		return
	} else if len(msg) > 4 && msg[:3] == "to|" {
		//消息格式为to|用户名|消息内容
		//1.获取对方用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			u.SendMsg("消息格式不正确,请使用to|用户名|消息内容")
			return
		}
		//2.根据用户名获取对方用户
		remoteUser, ok := u.server.OnlineMap[remoteName]
		if !ok {
			u.SendMsg("该用户不存在")
			return
		}
		//3.将消息转发给对方用户
		content := strings.Split(msg, "|")[2]
		if content == "" {
			u.SendMsg("空内容，发送失败")
			return
		}
		remoteUser.SendMsg(u.Username + "对您说：" + content)
	} else {
		// 将消息广播给其他在线用户
		u.server.BroadCast(u, msg)
	}

}

// 监听用户channel的方法
func (u *User) ListenMessage() {
	for {
		msg := <-u.Channel
		u.conn.Write([]byte(msg + "\n"))
	}
}
