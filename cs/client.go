package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func getServerMsg(conn net.Conn) {

	defer conn.Close()

	for {
		buff := make([]byte, 1024*100)
		nRead, eRead := conn.Read(buff)
		if eRead != nil {
			if eRead == io.EOF {
				break
			}
			if nRead != 0 {
				fmt.Println("[CLIENT]read server message fail...")
			}
			continue
		}

		fmt.Print(string(buff[:nRead]))
	}

}

func main() {

	conn, error := net.Dial("tcp", ":9800")

	defer conn.Close()

	if error != nil {
		fmt.Println(error)
		os.Exit(0)
	}

	//阻塞模式读取用户键盘输入并发送给服务器

	go getServerMsg(conn)

	for {
		userInputStr := ""
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			if len(scanner.Bytes()) == 0 {
				continue
			}
			userInputStr = scanner.Text()
		}

		if strings.ToLower(userInputStr) == "exit" || strings.ToLower(userInputStr) == "exit\r" {
			fmt.Printf("[CLIENT]用户退出成功...\n")
			break
		}

		_, err := conn.Write([]byte(userInputStr))

		if err != nil {
			fmt.Println("[CLIENT]消息发送失败...")
			continue
		}

		//fmt.Print("please input :")

	}

}
