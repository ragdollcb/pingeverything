package main

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"net"
	"os"
	"time"
)

func checkLatency(dial net.Conn) time.Duration {
	// 记录开始时间
	start := time.Now()

	// 发送 PING 命令（使用 Redis 协议格式）
	//_, err := dial.Write([]byte("*1\r\n$4\r\nPING\r\n"))
	_, err := dial.Write([]byte("PING"))
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("发送数据\n")

	// 读取响应（不关心内容）
	buffer := make([]byte, 1024)
	_, err = dial.Read(buffer)
	if err != nil {
		// 如果读取超时或发生其他错误，打印错误信息
		fmt.Printf("读取响应时发生错误: %v\n", err)
	}

	// 计算延迟
	latency := time.Since(start)
	fmt.Printf("连接延迟: %v\n", latency)

	return latency
}

func main() {
	// 创建日志文件
	logFile, err := os.OpenFile("./connection.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	// 设置日志输出到文件
	logger := log.New(logFile, "", log.LstdFlags)

	//配置文件获取
	// 读取配置文件
	viper.SetConfigName("config") // 不带扩展名的文件名
	viper.SetConfigType("ini")    // 配置文件类型
	viper.AddConfigPath(".")      // 当前目录
	err = viper.ReadInConfig()
	if err != nil {
		logger.Fatal(err)
	}

	//基本参数设定
	host := viper.GetString("address.host")
	port := viper.GetString("address.port")
	sleepTime := viper.GetInt("address.sleeptime")
	alarmTime := viper.GetInt("address.threshold")
	address := host + ":" + port

	for {
		// 建立TCP连接
		//fmt.Printf("准备建立连接")
		timeout := time.Duration(5) * time.Second
		dial, err := net.DialTimeout("tcp", address, timeout)
		if err != nil {
			logger.Printf("%v", err) // 继续下一次循环
			fmt.Printf("无法连接：%v", err)
		}
		//fmt.Printf("连接建立完成")
		defer dial.Close()
		//尝试连接并计算值
		latency := checkLatency(dial)

		// 定义延迟阈值

		threshold := time.Duration(alarmTime) * time.Millisecond // 例如 100 毫秒

		// 检查延迟并记录日志
		if latency > threshold {
			logger.Printf("Warning: Connection latency exceeded threshold for %v! Latency: %v\n", address, latency)
		}

		time.Sleep(time.Duration(sleepTime) * time.Millisecond)
	}

}
