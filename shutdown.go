package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"strings"
	"time"
)

func controlPorts(action string) {
	// 登录信息
	user := "admin"
	password := "你的密码" // ← 改成你的密码
	host := "10.0.6.250:22"

	// 要执行的命令
	var commands []string
	if action == "down" {
		commands = []string{
			"system-view",
			"interface GigabitEthernet 1/0/23",
			"shutdown",
			"interface GigabitEthernet 1/0/24",
			"shutdown",
			"return",
			"quit",
		}
	} else if action == "up" {
		commands = []string{
			"system-view",
			"interface GigabitEthernet 1/0/23",
			"undo shutdown",
			"interface GigabitEthernet 1/0/24",
			"undo shutdown",
			"return",
			"quit",
		}
	} else {
		log.Fatalf("未知操作类型: %s", action)
	}

	// 创建 SSH 客户端配置
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 跳过 host key 校验
		Timeout:         5 * time.Second,
	}

	// 建立连接
	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		log.Fatalf("SSH连接失败: %v", err)
	}
	defer client.Close()

	// 启动交互式 shell
	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("创建Session失败: %v", err)
	}
	defer session.Close()

	// 建立交互式终端
	stdin, err := session.StdinPipe()
	if err != nil {
		log.Fatalf("获取输入管道失败: %v", err)
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		log.Fatalf("获取输出管道失败: %v", err)
	}

	// 启动 shell
	if err := session.Shell(); err != nil {
		log.Fatalf("启动Shell失败: %v", err)
	}

	// 发送命令
	for _, cmd := range commands {
		fmt.Fprintf(stdin, "%s\n", cmd)
		time.Sleep(500 * time.Millisecond) // 等待命令执行
	}

	// 结束交互
	fmt.Fprint(stdin, "exit\n")
	fmt.Fprint(stdin, "exit\n")

	// 读取部分输出
	buf := make([]byte, 4096)
	n, _ := stdout.Read(buf)
	fmt.Println("输出结果：")
	fmt.Println(strings.TrimSpace(string(buf[:n])))

	fmt.Println("操作完成！")
}
