# 小智硬件烧录配置

## 烧录前准备

### 1. 获取烧录工具
- **ESP32**: 使用 `esptool.py`
- **ESP8266**: 使用 `esptool.py` 或 `NodeMCU-PyFlasher`

### 2. 准备固件文件
- 小智固件 `.bin` 文件
- 分区表文件（如果需要）

## 烧录配置方法

### 方法一：通过串口配置（推荐）

#### 1. 烧录固件
```bash
# ESP32 烧录命令
esptool.py --chip esp32 --port /dev/ttyUSB0 --baud 921600 \
  --before default_reset --after hard_reset write_flash \
  0x1000 bootloader.bin \
  0x8000 partition-table.bin \
  0x10000 xiaozhi-firmware.bin
```

#### 2. 串口配置服务器地址
烧录完成后，通过串口工具连接设备：

```bash
# 连接串口
screen /dev/ttyUSB0 115200
# 或使用 minicom
minicom -D /dev/ttyUSB0 -b 115200
```

#### 3. 发送配置命令
在串口终端中输入：

```
# 设置服务器地址
set_server 你的服务器IP
set_port 8000
set_device_id 你的设备ID
set_client_id 你的客户端ID

# 保存配置
save_config
reset
```

### 方法二：修改固件源码重新编译

#### 1. 获取固件源码
```bash
git clone https://github.com/xiaozhi-project/xiaozhi-firmware.git
cd xiaozhi-firmware
```

#### 2. 修改配置文件
编辑 `config.h` 或 `main.cpp`：

```cpp
// 服务器配置
#define SERVER_HOST "你的服务器IP"
#define SERVER_PORT 8000
#define SERVER_WS_PATH "/xiaozhi/v1/"

// 设备配置
#define DEVICE_ID "你的设备ID"
#define CLIENT_ID "你的客户端ID"
#define DEVICE_NAME "我的小智"
```

#### 3. 重新编译和烧录
```bash
# 编译
pio run

# 烧录
pio run --target upload
```

### 方法三：通过 Web 配置界面

#### 1. 连接设备 WiFi
- 设备启动后会创建热点：`Xiaozhi_XXXXXX`
- 密码通常是：`12345678`

#### 2. 访问配置页面
- 连接设备热点后，访问：`http://192.168.4.1`
- 在配置页面设置你的服务器地址

#### 3. 保存配置
- 输入你的服务器IP和端口
- 保存并重启设备

## 配置验证

### 1. 查看设备日志
```bash
# 连接串口查看日志
screen /dev/ttyUSB0 115200
```

应该看到类似日志：
```
连接服务器: ws://你的服务器IP:8000/xiaozhi/v1/?device-id=...
WebSocket 连接成功
发送 hello 消息: {"type":"hello","device_mac":"..."}
```

### 2. 查看服务器日志
在你的服务器上应该看到：
```
缓存mac-session: 你的设备ID <session_id>
缓存client-id-session: 你的客户端ID <session_id>
```

### 3. 测试推送
```bash
curl -X POST http://你的服务器IP:8080/api/push \
  -H "Content-Type: application/json" \
  -d '{"id":"你的客户端ID","text":"你好小智"}'
```

## 常见问题

### 1. 烧录失败
- 检查串口连接
- 确认设备进入下载模式
- 尝试不同的波特率

### 2. 配置不生效
- 确认配置已保存
- 重启设备
- 检查配置文件格式

### 3. 连接失败
- 检查服务器IP和端口
- 确认网络连接
- 查看设备日志

## 推荐配置

```cpp
// 推荐配置示例
#define SERVER_HOST "192.168.1.100"  // 你的服务器IP
#define SERVER_PORT 8000
#define DEVICE_ID "AA:BB:CC:DD:EE:FF"  // 设备唯一ID
#define CLIENT_ID "1111111"             // 客户端ID
``` 