package main

import (
	"fmt"
	"image/png"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/kbinani/screenshot"
	"github.com/lxn/walk"
	"golang.org/x/crypto/ssh"
	"strconv"
)

func extractIPv4AndFormat(filepath string, blacklistName string) (cmd string, extraIPs []string, err error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", nil, err
	}
	text := string(data)

	ipv4Regex := `\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`
	re := regexp.MustCompile(ipv4Regex)
	ips := re.FindAllString(text, -1)

	ipMap := make(map[string]bool)
	var uniqueIPs []string
	for _, ip := range ips {
		if !ipMap[ip] {
			ipMap[ip] = true
			uniqueIPs = append(uniqueIPs, ip)
		}
	}

	if len(uniqueIPs) == 0 {
		return "", nil, fmt.Errorf("未发现有效 IPv4 地址")
	}

	maxIPs := 256
	mainIPs := uniqueIPs
	if len(uniqueIPs) > maxIPs {
		mainIPs = uniqueIPs[:maxIPs]
		extraIPs = uniqueIPs[maxIPs:]
	}

	ipStr := strings.Join(mainIPs, " ")
	cmd = fmt.Sprintf("define host add name %s ipaddr '%s'", blacklistName, ipStr)

	return cmd, extraIPs, nil
}

func executeSSHCommand(user, password, host string, port int, cmd string) (string, error) {
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	address := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return "", fmt.Errorf("SSH连接失败: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建SSH会话失败: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	return string(output), err
}

func saveScreenshot(filename string) error {
	n := screenshot.NumActiveDisplays()
	if n <= 0 {
		return fmt.Errorf("无法获取显示器")
	}
	bounds := screenshot.GetDisplayBounds(0)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, img)
}

func saveExtraIPs(extra []string, path string) error {
	content := strings.Join(extra, "\n")
	return ioutil.WriteFile(path, []byte(content), 0644)
}

func logExecution(logFile string, cmd string, result string, err error) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	entry := fmt.Sprintf("[%s]\n命令: %s\n结果: %s\n错误: %v\n\n", timestamp, cmd, result, err)
	f, _ := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(entry)
}

func main() {
	var mw *walk.MainWindow
	var output *walk.TextEdit

	var userEdit, passEdit, hostEdit, portEdit, nameEdit *walk.LineEdit

	MainWindow{
		AssignTo: &mw,
		Title:    "风险IP自动封禁系统",
		MinSize:  Size{600, 500},
		Layout:   VBox{},
		Children: []Widget{
			Label{Text: "防火墙地址:"},
			LineEdit{AssignTo: &hostEdit},
			Label{Text: "SSH端口:"},
			LineEdit{AssignTo: &portEdit},
			Label{Text: "用户名:"},
			LineEdit{AssignTo: &userEdit},
			Label{Text: "密码:"},
			LineEdit{AssignTo: &passEdit, PasswordMode: true},
			Label{Text: "封禁列表名称:"},
			LineEdit{AssignTo: &nameEdit},

			PushButton{
				Text: "执行封禁命令",
				OnClicked: func() {
					port, err := strconv.Atoi(portEdit.Text())
					if err != nil {
						output.SetText("❌ 端口号格式错误")
						return
					}
					cmd, extraIPs, err := extractIPv4AndFormat("C:/temp/chatlog.txt", nameEdit.Text())
					if err != nil {
						output.SetText("❌ 处理失败: " + err.Error())
						return
					}

					output.SetText("⏳ 正在连接 SSH 并执行:\n" + cmd)

					go func() {
						result, err := executeSSHCommand(userEdit.Text(), passEdit.Text(), hostEdit.Text(), port, cmd)

						logExecution("firewall_log.txt", cmd, result, err)

						// 保存多余 IP
						if len(extraIPs) > 0 {
							saveExtraIPs(extraIPs, "exceed_ips.txt")
						}

						// 截图
						saveScreenshot("screenshot_" + time.Now().Format("20060102_150405") + ".png")

						if err != nil {
							output.SetText("❌ SSH 执行失败:\n" + err.Error() + "\n返回: " + result)
						} else {
							output.SetText("✅ 执行成功:\n" + result)
						}
					}()
				},
			},
			TextEdit{
				AssignTo: &output,
				ReadOnly: true,
				VScroll:  true,
			},
		},
	}.Run()
}
