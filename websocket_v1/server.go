package websocket_v1

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
)

// 1、结构体 -------------------------------------------------------------------------
type Server struct {
	ClientHeartTimeOut   int                    // 客户端超时时间 默认60秒
	OnHookEvent          func(Msg HookEvent)    // hook回调消息
	ChanHookEvent        chan *HookEvent        // 所有消息，各个子连接传过来的
	chanBroadCastMessage chan UDataSocket       // 消息广播的channel
	onlineMap            map[string]*serverUser // 在线用户的列表
	onlineMapLock        sync.RWMutex           // 同步锁
	SendFlag             int                    // socket验证标记
}

//
type HookEvent struct {
	EventType string // 事件类型 online / offline / message
	User      *serverUser
	Message   UDataSocket
}

// 2、全局变量 -------------------------------------------------------------------------

// 3、初始化函数 -------------------------------------------------------------------------

// 创建一个server的实例
func NewServer(OnHookEvent func(Msg HookEvent)) *Server {
	server := &Server{
		onlineMap:            make(map[string]*serverUser),
		ClientHeartTimeOut:   60 * 3,
		chanBroadCastMessage: make(chan UDataSocket),
		ChanHookEvent:        make(chan *HookEvent),
		OnHookEvent:          OnHookEvent,
	}

	// 启动监听Message的goroutine
	go server.goTranHookMessage()

	return server
}

// 对外函数2：连接服务器
func (Me *Server) Set(opt string, value interface{}) *Server {
	if opt == "SendFlag" {
		Me.SendFlag = value.(int)
	}
	return Me
}

// 消息发送
func (Me *Server) SendMsg(ClientId *string, Msg UDataSocket) error {
	if ClientId == nil {
		// 将msg发送给全部的在线User
		Me.onlineMapLock.Lock()
		for _, cli := range Me.onlineMap {
			cli.C <- Msg
		}
		Me.onlineMapLock.Unlock()

		return nil
	} else {
		Me.onlineMapLock.Lock()
		user, ok := Me.onlineMap[*ClientId]
		Me.onlineMapLock.Unlock()

		if ok {
			if err := user.sendSocketMsg(user.Conn, Msg); err != nil {
				user.Offline()
				return err
			}
			return nil
		} else {
			return errors.New("用户不在线")
		}
	}
}

// ////////////////////////////////////////////////////

// 处理客户端连接
func (Me *Server) NewUser(ws *websocket.Conn, S *Server) {
	// 1、实例化用户：新用户来了
	user := newUser(ws, Me)

	// 2、在线处理
	user.Online()

	// 3、打印信息
	fmt.Println("链接建立成功", user.ClientId, " 当前用户:", len(Me.onlineMap))

	// 4、接收客户端消息
	user.goListenClientMsg()
	// fmt.Println("用户守护进程已退出！")
}

// /////////////////////////////////////////////////

// 转发hook所有消息
func (Me *Server) goTranHookMessage() {
	for {
		select {
		case Event, ok := <-Me.ChanHookEvent:
			if !ok {
				return
			}
			// 推给应用的事件，除了心跳的所有事件
			if Event.Message.CType != 1 {
				Me.OnHookEvent(*Event)
			}
		}
	}
}
