package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

//创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,

		server: server,
	}
	//启动监听当前user channel消息的goroutine
	go user.ListenMessage()

	return user
}

//用户上线的业务
func (this *User) Online() {
	//用户上线，将用户加入到onlineMap中
	this.server.mapLock.Lock()
	this.server.onlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	//广播当前用户上线消息
	this.server.BroadCast(this, "已上线")
}

//用户下线的任务
func (this *User) Offline() {

	//用户下线，将用户从onlineMap中删除
	this.server.mapLock.Lock()
	delete(this.server.onlineMap, this.Name)
	this.server.mapLock.Unlock()

	//广播当前用户下线消息
	this.server.BroadCast(this, "下线")
}

//给当前user对应的客户端发送消息
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

//用户消息处理
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		//查询当前在线用户

		this.server.mapLock.Lock()
		for _, user := range this.server.onlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线....\n"
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//增加修改用户名功能，消息格式 rename|张三
		newName := strings.Split(msg, "|")[1]

		//判断naem是否存在
		_, ok := this.server.onlineMap[newName]

		if ok {
			this.SendMsg("当前用户名已被使用")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.onlineMap, this.Name)
			this.server.onlineMap[newName] = this
			this.server.mapLock.Unlock()

			//更新用户名
			this.Name = newName
			this.SendMsg("您已更新用户名为:" + this.Name + "\n")
		}

	} else {

		this.server.BroadCast(this, msg)
	}
}

//监听当前的User Channel的方法，一旦有消息，就直接发送给对端客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}
