package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var ipv4Regex = regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)

func main() {
	//定义命令行参数
	inputFile := flag.String("input", "", "输入文件路径（必填）")
	outputFile := flag.String("output", "output.txt", "输出文件路径")
	flag.Parse()

	//验证输入参数
	if *inputFile == "" {
		fmt.Println("错误：必须指定输入文件路径")
		flag.Usage()
		os.Exit(1)
	}

	//读取文件
	content, err := os.ReadFile(*inputFile)
	if err != nil {
		fmt.Println("文件读取失败: %v\n", err)
		os.Exit(2)
	}

	//预处理阶段（兼容多分隔符）
	rawText := strings.ReplaceAll(string(content), "\n", " ")
	rawText = strings.ReplaceAll(rawText, ",", " ")
	rawText = strings.ReplaceAll(rawText, "\t", " ")
	rawText = strings.ReplaceAll(rawText, "\r", " ")

	//精准过滤流程
	var validIPs []string
	for _, candidate := range strings.Fields(rawText) {
		if ipv4Regex.MatchString(candidate) {
			validIPs = append(validIPs, candidate)
		}
	}

	if len(validIPs) == 0 {
		fmt.Println("错误：未检测到有效的 IPv4地址")
		os.Exit(3)
	}

	output := "'" + strings.Join(validIPs, " ") + "'"
	if err := os.WriteFile(*outputFile, []byte(output), 0644); err != nil {
		fmt.Println("文件写入失败", err)
		os.Exit(4)
	}

	fmt.Printf("过滤完成，保留 %d 个 IPv4地址，保存至：：%s\n", len(validIPs), *outputFile)

}
