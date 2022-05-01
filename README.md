# go-websocket
websocket服务器端封装
## 安装
go get github.com/loudbund/go-websocket

## 引入
```golang
import "github.com/loudbund/go-websocket/websocket_v1"
```

## 使用方法
### 服务端
>A.创建连接并设置传输码，（传输码需要和客户端一致）
>
    addr := "0.0.0.0:3020"
	Server = websocket_v1.NewServer(func(Event websocket_v1.HookEvent) {
		socketOnHookEvent(Event)
	}).Set("SendFlag", 398359203)
 
>B.编写响应服务器消息事件的函数
>
    func socketOnHookEvent(Event websocket_v1.HookEvent) {
 	    switch Event.EventType {
 	    case "message": // 1、消息事件
 		fmt.Println(Event.Message.CType, string(Event.Message.Content))
 	    case "offline": // 2、下线事件
 		onlineNum--
 	    case "online": // 3、上线消息
 		onlineNum++
 	    }
    }
>C.给客户端发消息
>
    _ = Server.SendMsg(nil, websocket_v1.UDataSocket{
        CType:   1000,
        Content: []byte("hello, [" + time.Now().Format("2006-01-02 15:04:05") + "]"),
    })
### 2、网页客户端
>参见example_server.go里面的js部分
```javascript
// 初始化启动
wsocket.Init("ws://" + window.location.host + "/websocket/")
// zjWebSocket.Init("ws://192.168.159.130:3021/websocket")
// 连接成功
wsocket.onOpen = function () {
    wsocket.DoSend({
        CType: 2001,
        // Content: JSON.stringify({ teacherId: "12345", status: "hiding" })
        Content: JSON.stringify({ teacherId: "" + teacherId, status: "online" })
    });
    console.log("connected ok ");
};
// 收到消息
wsocket.onMessage = function (Data) {
    console.log("message received: " + Data.Content);
    document.getElementById("receivedId").innerHTML = "message received: " + Data.Content + "\n" + document.getElementById("receivedId").innerHTML
};
// 连接断开事件
wsocket.onClose = function (e) {
    console.log("connection closed (" + e.code + ")");
}
```