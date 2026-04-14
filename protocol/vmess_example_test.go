package protocol

import (
	"encoding/json"
	"fmt"
	"log"
)

// ExampleDecodeVmessUrl 演示如何解析 vmess URL
func ExampleDecodeVmessUrl() {
	// vmess URL 示例
	vmessURL := "vmess://eyJ2IjoiMiIsInBzIjoiSEsuXHU5OTk5XHU2ZTJmLkMgfCBcdTlhZDhcdTkwMWYuMnhcdTUwMGRcdTczODciLCJhZGQiOiJ4Y3RyYW5zZmVyMXN0LmNva2VjbG91ZC50b3AiLCJwb3J0IjoiMzg2MzAiLCJpZCI6IjNmYzM2NWYyLTk0MGMtNGUwNy05Y2Y3LWExMGM1N2JlODU3YSIsImFpZCI6IjAiLCJuZXQiOiJ0Y3AiLCJ0eXBlIjoibm9uZSIsImhvc3QiOiIiLCJwYXRoIjoiIiwidGxzIjoiIn0="

	// 解析 vmess URL
	node, err := DecodeVmessUrl(vmessURL)
	if err != nil {
		log.Fatalf("解析失败: %v", err)
	}

	fmt.Printf("节点名称: %s\n", node.Remarks)
	fmt.Printf("服务器地址: %s\n", node.Address)
	fmt.Printf("端口: %s\n", node.Port)
	fmt.Printf("UUID: %s\n", node.ID)
	fmt.Printf("网络类型: %s\n", node.Network)

	// Output:
	// 节点名称: HK.香港.C | 高速.2x倍率
	// 服务器地址: xctransfer1st.cokecloud.top
	// 端口: 38630
	// UUID: 3fc365f2-940c-4e07-9cf7-a10c57be857a
	// 网络类型: tcp
}

// ExampleConvertVmessToSingBox 演示如何将 vmess 节点转换为 sing-box 配置
func ExampleConvertVmessToSingBox() {
	// 创建一个 vmess 节点
	node := &VmessNode{
		Version: "2",
		Remarks: "测试节点",
		Address: "example.com",
		Port:    "443",
		ID:      "3fc365f2-940c-4e07-9cf7-a10c57be857a",
		AlterID: "0",
		Network: "ws",
		Type:    "none",
		Host:    "example.com",
		Path:    "/ws",
		TLS:     "tls",
	}

	// 转换为 sing-box 配置
	singboxOut, err := ConvertVmessToSingBox(node)
	if err != nil {
		log.Fatalf("转换失败: %v", err)
	}

	// 序列化为 JSON（美化输出）
	jsonData, err := json.MarshalIndent(singboxOut, "", "  ")
	if err != nil {
		log.Fatalf("序列化失败: %v", err)
	}

	fmt.Println("Sing-box 配置:")
	fmt.Println(string(jsonData))

	// Output:
	// Sing-box 配置:
	// {
	//   "server": "example.com",
	//   "server_port": 443,
	//   "tag": "测试节点",
	//   "tls": {
	//     "enabled": true,
	//     "server_name": "example.com"
	//   },
	//   "transport": {
	//     "headers": {
	//       "Host": [
	//         "example.com"
	//       ]
	//     },
	//     "path": "/ws",
	//     "type": "ws"
	//   },
	//   "type": "vmess",
	//   "security": "auto",
	//   "uuid": "3fc365f2-940c-4e07-9cf7-a10c57be857a",
	//   "multiplex": {}
	// }
}

// ExampleDecodeVmessUrlToSingBox 演示一步完成 vmess URL 到 sing-box 配置的转换
func ExampleDecodeVmessUrlToSingBox() {
	// vmess URL 示例
	vmessURL := "vmess://eyJ2IjoiMiIsInBzIjoiSEsuXHU5OTk5XHU2ZTJmLkMgfCBcdTlhZDhcdTkwMWYuMnhcdTUwMGRcdTczODciLCJhZGQiOiJ4Y3RyYW5zZmVyMXN0LmNva2VjbG91ZC50b3AiLCJwb3J0IjoiMzg2MzAiLCJpZCI6IjNmYzM2NWYyLTk0MGMtNGUwNy05Y2Y3LWExMGM1N2JlODU3YSIsImFpZCI6IjAiLCJuZXQiOiJ0Y3AiLCJ0eXBlIjoibm9uZSIsImhvc3QiOiIiLCJwYXRoIjoiIiwidGxzIjoiIn0="

	// 一步转换
	singboxOut, err := DecodeVmessUrlToSingBox(vmessURL)
	if err != nil {
		log.Fatalf("转换失败: %v", err)
	}

	fmt.Printf("类型: %s\n", singboxOut.Type)
	fmt.Printf("标签: %s\n", singboxOut.Tag)
	fmt.Printf("服务器: %s:%d\n", singboxOut.Server, singboxOut.ServerPort)
	fmt.Printf("UUID: %s\n", singboxOut.UUID)
	fmt.Printf("AlterID: %d\n", singboxOut.AlterID)
	fmt.Printf("安全性: %s\n", singboxOut.Security)

	// Output:
	// 类型: vmess
	// 标签: HK香港C高速2x倍率
	// 服务器: xctransfer1st.cokecloud.top:38630
	// UUID: 3fc365f2-940c-4e07-9cf7-a10c57be857a
	// AlterID: 0
	// 安全性: auto
}
