package websocket_v1

import (
	"fmt"
	"github.com/gorilla/websocket"
	"reflect"
	"time"
)

// 1、结构体 -------------------------------------------------------------------------

// 本模块封装用结构体
type serverUser struct {
	socketMsg
	ClientId string           // 客户端id(服务器给客户端分配的唯一标志)
	Conn     *websocket.Conn  // 客户端连接
	Name     string           // 客户端名称
	Addr     string           // 客户端地址
	Server   *Server          // Server指针，主要用来将自己加入和移除Server用户列表里
	C        chan UDataSocket // 发消息channel，Server将要消息推送到channel里
}

// 2、全局变量 -------------------------------------------------------------------------

// 3、初始化函数 -------------------------------------------------------------------------

// 4、函数 -------------------------------------------------------------------------

// 创建一个用户的API
func newUser(conn *websocket.Conn, server *Server) *serverUser {
	user := &serverUser{
		ClientId:  utilUuidShort(),
		Conn:      conn,
		Name:      conn.RemoteAddr().String(),
		Addr:      conn.RemoteAddr().String(),
		Server:    server,
		C:         make(chan UDataSocket, 10),
		socketMsg: socketMsg{SendFlag: server.SendFlag},
	}
	return user
}

// //////////////////////////////////////////////////////////////////

// 启动运行
func (Me *serverUser) goListenClientMsg() {
	// 监听用户是否活跃的channel
	isLive := make(chan bool)

	// 接收消息
	go func() {
		if err := Me.getSocketMsg(Me.Conn, func(msg *UDataSocket) bool {
			// 用户的任意消息，代表当前用户是一个活跃的
			isLive <- true

			// 收到问候的消息
			if msg.CType == 7 {
				fmt.Println(Me.ClientId, msg.CType, string(msg.Content))
				_ = Me.sendSocketMsg(Me.Conn, UDataSocket{8, []byte("hello test msg from server")})
			}

			// 3、消息发给主进程
			Me.Server.ChanHookEvent <- &HookEvent{"message", Me, *msg}
			return true
		}); err != nil {
			// fmt.Println("消息接收终止，退出消息接收协程")
			isLive <- false
			Me.Offline()
		}
	}()

	// 等待心跳
	Me.waitHeartBeet(isLive)
}

// 等待心跳
func (Me *serverUser) waitHeartBeet(isLive chan bool) {
	// 阻塞住
	for {
		select {
		case msg := <-Me.C:
			if !reflect.DeepEqual(msg, reflect.Zero(reflect.TypeOf(msg)).Interface()) {
				if err := Me.sendSocketMsg(Me.Conn, msg); err != nil {
					Me.Offline()
					return // 退出socket协程 // fmt.Println("消息发送失败，用户进程阻塞终止，退出用户协程")
				}
			}

		case live, ok := <-isLive:
			if !live || !ok { // fmt.Println("收到掉线通知，用户进程阻塞终止，退出用户协程")
				return
			}

			_ = Me.Conn.SetReadDeadline(time.Now().Add(time.Duration(Me.Server.ClientHeartTimeOut) * time.Second))
			_ = Me.Conn.SetWriteDeadline(time.Now().Add(time.Duration(Me.Server.ClientHeartTimeOut) * time.Second))

			// 当前用户是活跃的，应该重置定时器
			// 不做任何事情，为了激活select，更新下面的定时器

		case <-time.After(time.Second * time.Duration(Me.Server.ClientHeartTimeOut)):
			Me.Offline()

			return // runtime.Goexit() // fmt.Println("心跳过期，用户进程阻塞终止，退出用户协程")
		}
	}
}

// 用户的上线业务
func (Me *serverUser) Online() {

	// 用户上线,将用户加入到onlineMap中
	Me.Server.onlineMapLock.Lock()
	Me.Server.onlineMap[Me.ClientId] = Me
	Me.Server.onlineMapLock.Unlock()

	// 2、事件发给server
	Me.Server.ChanHookEvent <- &HookEvent{"online", Me, UDataSocket{}}
}

// 用户的下线业务
func (Me *serverUser) Offline() {

	// 用户下线，将用户从onlineMap里移除
	Me.Server.onlineMapLock.Lock()
	if _, ok := Me.Server.onlineMap[Me.ClientId]; ok {
		_ = Me.Conn.Close()                      // 释放资源 - 关闭socket链接
		close(Me.C)                              // 释放资源 - 销毁用的资源
		delete(Me.Server.onlineMap, Me.ClientId) // 移除用户
		fmt.Println(Me.Name, "退出成功", "当前在线", len(Me.Server.onlineMap))
	}
	Me.Server.onlineMapLock.Unlock()

	// 2、事件发给server
	Me.Server.ChanHookEvent <- &HookEvent{"offline", Me, UDataSocket{}}
}
