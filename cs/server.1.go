package main

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	UserChannel chan string
	Name        string
}

//广播信息列表
var msgList = make(chan string)

//在线列表
var onlineList = make(map[string]Client)

//广播协程:
//1、实时获取消息通道msgList
//2、将获取到的消息传给所有连接的client专用的信息通道
func broadcast(conn net.Conn) {

	for {
		msg := <-msgList
		if strings.HasPrefix(msg, "[系统消息]") {
			fmt.Println("[INFO]-->", msg)
		}

		//遍历在线用户列表，并把这条消息推送到用户专属通道中
		for _, val := range onlineList {
			val.UserChannel <- msg
		}

		//fmt.Println("map", onlineList)
	}
}

//消息格式化
func msgFormat(message, addr string, isSys bool) (msg string) {
	var _msgFormat string
	if isSys {
		_msgFormat = "[系统消息]用户[%v]于%v %v \n"
	} else {
		_msgFormat = "[用户消息][%v] %v 说：%v \n"
	}

	if len(addr) == 0 {
		msg = fmt.Sprintf("[系统消息]%v \n", message)
	} else {
		msg = fmt.Sprintf(_msgFormat, addr, time.Now().Format("15:04:05"), message)
	}
	return
}

func main() {

	//开启监听
	listener, err := net.Listen("tcp", ":9800")

	if err != nil {
		fmt.Println("listen error:", err)
		return
	} else {
		fmt.Println("[INFO]-->", "等待客户端连接中...")
	}

	defer listener.Close()

	//启动监控协程，监控-->广播通道
	//go broadcast()

	//主进程等待连接
	for {
		conn, error := listener.Accept()

		addr := conn.RemoteAddr().String()

		if error != nil {
			fmt.Println("listen Accept error:", error)
			continue
		}

		go broadcast(conn)

		//连接成功用户后，初始化用户信息，包含昵称Name和专属通道UserChannel
		cli := Client{Name: addr, UserChannel: make(chan string)}
		onlineList[addr] = cli

		//推送用户登录信息-->广播通道
		msgList <- msgFormat("登录成功...", addr, true)

		//新增接收协程:接收当前连接客户端发送的消息
		go receiveFromClient(conn, addr, cli)

		//新增发送协程:将用户专属通道UserChannel中的消息推送到当前连接的客户端
		go sendToClient(conn, cli)

	}

}

//接收协程:接收所有client发送的消息
func receiveFromClient(conn net.Conn, addr string, cli Client) {

	defer conn.Close()

	buff := make([]byte, 1024*100)

	hasDataChannel := make(chan bool)

	go func() {
		for {

			author := cli.Name

			n, err := conn.Read(buff)

			if err != nil {
				if err == io.EOF {
					msgList <- msgFormat("退出...", author, true)
				} else {
					fmt.Printf("[%v]conn err: %s\n", author, err)
				}
				delete(onlineList, addr)
				break
			}

			message := strings.ToLower(string(buff[:n]))
			//当客户端发送exit时，自动断开连接
			if message == "exit" || message == "exit\r" {
				//fmt.Println(message)
				msgList <- msgFormat("下线...", author, true)
				delete(onlineList, addr)
				break
			}

			if message == "who" {
				conn.Write([]byte(msgFormat("当前在线："+strconv.Itoa(len(onlineList))+"人", "", true)))
				continue
			}

			if strings.HasPrefix(message, "rename") {
				_str := strings.Split(message, " ")
				//fmt.Printf("%+v %+v %v", _str, _str[1], _str[1])
				cli.Name = _str[1]
				fmt.Println(onlineList)
				continue
			}

			msgList <- msgFormat(message, author, false) //fmt.Sprintf("[%v]%v %v", addr, time.Now().Format("15:04:05"), message)

			hasDataChannel <- true
		}
	}()

TIMEOUT:
	for {
		select {
		case <-hasDataChannel:
			fmt.Printf("hasDataChannel %+v \n", <-hasDataChannel)
			//利用hasDataChannel通道的变化判断用户是否超时，当hasDataChannel没有值变化时就会执行下面的代码
		case <-time.After(100 * time.Second):
			msgList <- msgFormat("超时退出...", addr, true)
			conn.Close()
			break TIMEOUT
		}
	}

}

//发送协程:将用户专属通道UserChannel中的消息推送到当前连接的client
func sendToClient(conn net.Conn, cli Client) {
	//defer conn.Close()
	for {
		userMsg := <-cli.UserChannel
		conn.Write([]byte(fmt.Sprintf("%s \n", userMsg)))
	}
}
