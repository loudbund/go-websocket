package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/zhoujianstudio/go-websocket/websocket_v1"
	"html/template"
	"log"
	"net/http"
	"time"
)

// 1、结构体 -------------------------------------------------------------------------

// 2、全局变量 -------------------------------------------------------------------------

var upgrader = websocket.Upgrader{
	EnableCompression: true,
	HandshakeTimeout:  5 * time.Second,
	// CheckOrigin: 处理跨域问题，线上环境慎用
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 3、初始化函数 -------------------------------------------------------------------------
var (
	Server *websocket_v1.Server
)

// 4、开放的函数 -------------------------------------------------------------------------

// 5、内部函数 -------------------------------------------------------------------------

func home(w http.ResponseWriter, r *http.Request) {
	_ = homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
}

// 处理数据,多线程转单线程处理
func onHookEvent(Event websocket_v1.HookEvent) {
	// 事件处理在此处 ///////////////////////////////////////////////////////////////
	switch Event.EventType {
	case "message": // 1、消息事件
	case "offline": // 2、下线事件
	case "online": // 3、上线消息
	}
	// ////////////////////////////////////////////////////////////////////////////
}

// 发送数据给所有客户端
func goTestSendMsg() {
	for {
		_ = Server.SendMsg(nil, websocket_v1.UDataSocket{
			CType:   1000,
			Content: []byte("hello"),
		})
		time.Sleep(time.Second)
	}
}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	go Server.NewUser(c, Server)
}

// 6、主函数 -------------------------------------------------------------------------
func main() {
	addr := "0.0.0.0:3020"
	Server = websocket_v1.NewServer(func(Event websocket_v1.HookEvent) {
		onHookEvent(Event)
	})

	// 演示用: 循环发消息
	go goTestSendMsg()

	// 开始监听：
	http.HandleFunc("/websocket/", echo)
	http.HandleFunc("/", home)
	fmt.Println("websocket开始监听:" + addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html lang="en">

<head>
    <meta charset="UTF-8">
    <title>Hello</title>
    <script language="javascript">
        YxpWebSocket = function () {
            let me = this
            let wsuri = "";     // 连接websocket的url "ws://192.168.61.59:3021/websocket";
            let sock = null;    // socket实例句柄
            let connected = false // 连接状态

            // 开始连接
            this.Init = function (url) {
                wsuri = url
                // 开始连接
                sock = new WebSocket(wsuri);
                // 设置二进制传输
                sock.binaryType = "arraybuffer";
                // 打开事件响应
                sock.onopen = function () {
                    console.log("websocket 连接成功")
                    connected = true

                    // 发送问候消息
                    me.DoSend({ CType: 7, Content: "hello test msg from client" })

                    // 回调
                    me.onOpen()
                }
                // 消息事件响应
                sock.onmessage = function (e) {
                    Data = receiveByte2Data(e.data)
                    me.onMessage(Data)
                }
                // 断开时间响应
                sock.onclose = function (e) {
                    connected = false
                    console.log("5秒后重连")
                    setTimeout(() => {
                        me.Init(wsuri)
                    }, 5000)
                    me.onClose(e)
                }
                // 出错响应
                sock.onerror = function (e) {
                    console.log("出错", e)
                    me.onError(e)
                }

            }

            // 使用方赋值
            this.onOpen = function () { }

            // 使用方赋值
            this.onClose = function () { }

            // 使用方赋值
            this.onMessage = function () { }

            // 使用方赋值
            this.onError = function () { }

            // 消息发送
            this.DoSend = function (Data) {
                console.log("send Msg:",Data)
                let Pass = 398359203
                let CType = Data.CType
                let Content = Data.Content
                let bytes = stringToBytes(Content)
                let buffer = new ArrayBuffer(bytes.length + 12);
                let view = new DataView(buffer);
                view.setUint32(0, Pass);
                view.setUint32(4, CType);
                view.setUint32(8, bytes.length);
                for (let i = 0; i < bytes.length; i++) {
                    view.setUint8(i + 12, bytes[i]);
                }
                let err = sock.send(view);
            };

            // 字符串转字节数组
            function stringToBytes(str) {
                var ch, st, re = [];
                for (var i = 0; i < str.length; i++) {
                    ch = str.charCodeAt(i);  // get char
                    st = [];                 // set up "stack"

                    do {
                        st.push(ch & 0xFF);  // push byte to stack
                        ch = ch >> 8;          // shift value down by 1 byte
                    }

                    while (ch);
                    // add stack contents to result
                    // done because chars have "wrong" endianness
                    re = re.concat(st.reverse());
                }
                // return an array of bytes
                return re;
            }

            // 接收到的字节数据转结构体数据
            function receiveByte2Data(buffer) {
                let receive = [];
                let length = 0;

                receive = receive.concat(Array.from(new Uint8Array(buffer)));
                if (receive.length < 4) {
                    return;
                }
                let View = new DataView(new Uint8Array(receive).buffer)
                let Pass = View.getUint32(0);
                let CType = View.getUint32(4);
                let Len = View.getUint32(8);

                // console.log(Pass, CType, Len)
                // console.log(utf8ByteToUnicodeStr(receive.slice(12)))
                return {
                    CType: CType,
                    Content: utf8ByteToUnicodeStr(receive.slice(12))
                }
            };

            // utf8字节转中文字符串
            function utf8ByteToUnicodeStr(utf8Bytes) {
                var unicodeStr = "";
                for (var pos = 0; pos < utf8Bytes.length;) {
                    var flag = utf8Bytes[pos];
                    var unicode = 0;
                    if ((flag >>> 7) === 0) {
                        unicodeStr += String.fromCharCode(utf8Bytes[pos]);
                        pos += 1;

                    } else if ((flag & 0xFC) === 0xFC) {
                        unicode = (utf8Bytes[pos] & 0x3) << 30;
                        unicode |= (utf8Bytes[pos + 1] & 0x3F) << 24;
                        unicode |= (utf8Bytes[pos + 2] & 0x3F) << 18;
                        unicode |= (utf8Bytes[pos + 3] & 0x3F) << 12;
                        unicode |= (utf8Bytes[pos + 4] & 0x3F) << 6;
                        unicode |= (utf8Bytes[pos + 5] & 0x3F);
                        unicodeStr += String.fromCharCode(unicode);
                        pos += 6;

                    } else if ((flag & 0xF8) === 0xF8) {
                        unicode = (utf8Bytes[pos] & 0x7) << 24;
                        unicode |= (utf8Bytes[pos + 1] & 0x3F) << 18;
                        unicode |= (utf8Bytes[pos + 2] & 0x3F) << 12;
                        unicode |= (utf8Bytes[pos + 3] & 0x3F) << 6;
                        unicode |= (utf8Bytes[pos + 4] & 0x3F);
                        unicodeStr += String.fromCharCode(unicode);
                        pos += 5;

                    } else if ((flag & 0xF0) === 0xF0) {
                        unicode = (utf8Bytes[pos] & 0xF) << 18;
                        unicode |= (utf8Bytes[pos + 1] & 0x3F) << 12;
                        unicode |= (utf8Bytes[pos + 2] & 0x3F) << 6;
                        unicode |= (utf8Bytes[pos + 3] & 0x3F);
                        unicodeStr += String.fromCharCode(unicode);
                        pos += 4;

                    } else if ((flag & 0xE0) === 0xE0) {
                        unicode = (utf8Bytes[pos] & 0x1F) << 12;;
                        unicode |= (utf8Bytes[pos + 1] & 0x3F) << 6;
                        unicode |= (utf8Bytes[pos + 2] & 0x3F);
                        unicodeStr += String.fromCharCode(unicode);
                        pos += 3;

                    } else if ((flag & 0xC0) === 0xC0) { //110
                        unicode = (utf8Bytes[pos] & 0x3F) << 6;
                        unicode |= (utf8Bytes[pos + 1] & 0x3F);
                        unicodeStr += String.fromCharCode(unicode);
                        pos += 2;

                    } else {
                        unicodeStr += String.fromCharCode(utf8Bytes[pos]);
                        pos += 1;
                    }
                }
                return unicodeStr;
            }

            // 心跳发送
            setInterval(() => {
                if (connected) {
                    this.DoSend({ CType: 1, Content: "" })
                }
            }, 5000)

            return this
        }



    </script>
</head>

<body>
    <h1 style="float:left">WebSocket Echo Test<span id="teacherid"></span></h1>
    <div style="float: right;">
        <form>
            <p>
                Message: <input id="message" type="text" value="Hello, world!">
            </p>
        </form>
        <button onclick="send();" style="width:200px;height:50px;">Send Message</button>
        <button onclick="Online();" style="width:200px;height:50px;">上线</button>
        <button onclick="Hiding();" style="width:200px;height:50px;">隐身</button>
    </div>
    <div style="clear:both;"></div>
    <br><br>
    <textarea id="receivedId" style="width: 100%;height:400px;"></textarea>
</body>

</html>


<script type="text/javascript">
    var teacherId = 500
    function getQueryString(name) {
        var reg = new RegExp("(^|&)" + name + "=([^&]*)(&|$)", "i");
        var r = window.location.search.substr(1).match(reg);
        if (r != null)
            return unescape(r[2]);
        return null;
    }
    if (getQueryString("teacherid")){
        teacherId=getQueryString("teacherid")
    }
    document.getElementById("teacherid").innerHTML = ":" + teacherId
    var wsocket = YxpWebSocket()

    window.onload = function () {

        // 初始化启动
        wsocket.Init("ws://" + window.location.host + "/websocket/")
        // zjWebSocket.Init("ws://192.168.159.130:3021/websocket")
        // 连接成功
        wsocket.onOpen = function () {
            wsocket.DoSend({
                CType: 2001,
                // Content: JSON.stringify({ teacherId: "12345", status: "hiding" })
                Content: JSON.stringify({ teacherId: "" + teacherId, status: "online" })
            })
            console.log("connected ok ");
        }
        // 收到消息
        wsocket.onMessage = function (Data) {
            console.log("message received: " + Data.Content);
            document.getElementById("receivedId").innerHTML = "message received: " + Data.Content + "\n" + document.getElementById("receivedId").innerHTML
        }
        // 连接断开事件
        wsocket.onClose = function (e) {
            console.log("connection closed (" + e.code + ")");
        }


    };

    function send() {
        var msg = document.getElementById('message').value;
        wsocket.DoSend({
            CType: 1011,
            Content: msg,
        })
    };
    function Online(){
        wsocket.DoSend({
            CType: 2001,
            // Content: JSON.stringify({ teacherId: "12345", status: "hiding" })
            Content: JSON.stringify({ teacherId: "" + teacherId, status: "online" })
        })
    }
    function Hiding(){
        wsocket.DoSend({
            CType: 2001,
            // Content: JSON.stringify({ teacherId: "12345", status: "hiding" })
            Content: JSON.stringify({ teacherId: "" + teacherId, status: "hiding" })
        })
    }

</script>
`))
