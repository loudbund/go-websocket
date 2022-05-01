package websocket_v1

import (
	"bytes"
	"errors"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// 结构体1：传输数据上层结构体
type UDataSocket struct {
	CType   int    // 内容类型 1:客户端请求消息 2:服务端表接口消息 4:服务端表内容数据 200:服务端发送结束
	Content []byte // 发送内容
}

// 结构体2：传输数据底层结构体
type unitDataSend struct {
	SendFlag    int    // 消息最前面标记
	CType       int    // 内容类型
	ContentTran []byte // 发送的内容
}

// 结构体3：本模块封装用结构体
type socketMsg struct {
	RevCache []byte // 收到数据缓存
	SendFlag int    // 消息最前面标记
}

// 内部函数1：发送socket消息
func (Me *socketMsg) sendSocketMsg(ws *websocket.Conn, Data UDataSocket) error {
	// 1、拼凑要发送的数据
	KeyBytesBuffer := bytes.NewBuffer([]byte{})
	if true {
		KeyBytesBuffer.Write(utilInt2Bytes(Me.SendFlag))
		KeyBytesBuffer.Write(utilInt2Bytes(Data.CType))
		KeyBytesBuffer.Write(utilInt2Bytes(len(Data.Content)))
		KeyBytesBuffer.Write(Data.Content)
	}

	// 2、发送数据
	err := ws.WriteMessage(2, KeyBytesBuffer.Bytes())
	if err != nil {
		return err
	}
	return nil
}

// 内部函数2：读取socket消息
func (Me *socketMsg) getSocketMsg(ws *websocket.Conn, fSuccess func(data *UDataSocket) bool) error {
	// 循环
	for {
		// 1、读取头数据
		KeyBuffHeader, err := Me.readSocketSizeData(ws, 12)
		if err != nil {
			return err
		}

		// 2、解析投数据
		KeyRevSendFlag := utilBytes2Int(KeyBuffHeader[0:4])
		KeyRevCType := utilBytes2Int(KeyBuffHeader[4:8])
		if true {
			if KeyRevSendFlag != Me.SendFlag {
				log.Error("传输码校验失败")
				return errors.New("传输码校验失败")
			}
		} // 校验码判断

		// 3、读取内容数据
		KeyContentTran := make([]byte, 0)
		if true {
			ContentTranLength := utilBytes2Int(KeyBuffHeader[8:12])
			if ContentTranLength > 1*1024 {
				log.Error("头消息不能超过1K")
				return errors.New("头消息不能超过1K")
			}
			ContentTran, err := Me.readSocketSizeData(ws, ContentTranLength)
			if err != nil {
				return err
			}
			KeyContentTran = ContentTran
		} // 读取内容数据到 KeyContentTran 变量

		// 4、回调收到消息事件
		continueRead := fSuccess(&UDataSocket{
			CType:   KeyRevCType,
			Content: KeyContentTran,
		})
		if !continueRead {
			break
		}
	}
	return nil
}

// 内部函数3：读取指定长度数据
func (Me *socketMsg) readSocketSizeData(ws *websocket.Conn, length int) ([]byte, error) {
	// 1、想读取0个字节，就拼凑一个给他
	if length <= 0 {
		return make([]byte, 0), nil
	}

	// 2、缓存池里有这么多数据，就直接返给他
	if ret, err := Me.getFromCache(length); err == nil {
		return ret, nil
	}

	// 3、再接收点数据吧
	for {
		// 3.1、读一波内容
		if _, ContentTran, err := ws.ReadMessage(); err == nil {
			// 3.1.1、读取的内容写入缓冲池
			Me.RevCache = append(Me.RevCache, ContentTran...)
			// 3.1.2、缓冲池数据够了就返回，否则就循环继续接收
			if ret, err := Me.getFromCache(length); err == nil {
				return ret, nil
			}
		} else {
			return nil, err
		}
	}
}

// 内部函数4：从缓存里读取指定长度数据
func (Me *socketMsg) getFromCache(length int) ([]byte, error) {
	// 缓冲池的数据够了，就返回
	if len(Me.RevCache) >= length {
		ret := Me.RevCache[0:length]
		Me.RevCache = Me.RevCache[length:]
		return ret, nil
	}
	// 不够就返回失败
	return nil, errors.New("缓存数据不足")
}
