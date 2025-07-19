# 小智硬件连接配置

## 服务器信息
- **WebSocket地址**: ws://你的服务器IP:8000/xiaozhi/v1/
- **HTTP API地址**: http://你的服务器IP:8080/api/
- **设备ID**: 你的小智设备ID（如MAC地址）
- **客户端ID**: 你的小智客户端ID

## 配置步骤

### 1. 修改小智固件配置

在小智固件的配置文件中修改：

```cpp
// 服务器配置
#define SERVER_HOST "你的服务器IP"
#define SERVER_PORT 8000
#define SERVER_WS_PATH "/xiaozhi/v1/"

// 设备配置
#define DEVICE_ID "你的设备ID"  // 如: AA:BB:CC:DD:EE:FF
#define CLIENT_ID "你的客户端ID" // 如: 1111111
```

### 2. 连接测试

小智硬件连接后，你应该在服务器日志中看到：

```
缓存mac-session: AA:BB:CC:DD:EE:FF <session_id>
缓存client-id-session: 1111111 <session_id>
```

### 3. 推送消息测试

```bash
curl -X POST http://你的服务器IP:8080/api/push \
  -H "Content-Type: application/json" \
  -d '{"id":"1111111","text":"你好小智"}'
```

## 常见问题

### 1. 连接失败
- 检查服务器IP是否正确
- 确认端口8000和8080是否开放
- 检查网络连接

### 2. 音频问题
- 确认小智硬件音频模块正常
- 检查麦克风和扬声器连接

### 3. 推送失败
- 确认设备ID和客户端ID正确
- 检查设备是否在线 