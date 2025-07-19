# ESP32 小智客户端

这是一个基于 ESP32 的小智语音助手客户端，可以连接到小智服务器进行语音交互。

## 硬件要求

- ESP32 开发板
- I2S 音频模块（如 MAX98357A）
- 麦克风模块（如 INMP441）
- 扬声器
- 按钮（可选，用于触发语音）

## 硬件连接

### I2S 音频连接
```
ESP32    ->   I2S 模块
GPIO 15  ->   WS (Word Select)
GPIO 13  ->   SD (Serial Data)  
GPIO 2   ->   SCK (Serial Clock)
GND      ->   GND
3.3V     ->   VCC
```

### 麦克风连接
```
ESP32    ->   INMP441
GPIO 15  ->   WS
GPIO 13  ->   SD
GPIO 2   ->   SCK
GND      ->   GND
3.3V     ->   VDD
```

## 软件配置

### 1. 安装 PlatformIO

```bash
# 安装 PlatformIO Core
pip install platformio

# 或者使用 VS Code 插件
# 搜索 "PlatformIO IDE" 并安装
```

### 2. 配置 WiFi 和服务器

编辑 `src/config.h` 文件：

```cpp
// WiFi 配置
#define WIFI_SSID "你的WiFi名称"
#define WIFI_PASSWORD "你的WiFi密码"

// 服务器配置
#define SERVER_HOST "192.168.1.100"  // 你的服务器IP
#define SERVER_PORT 8000
```

### 3. 编译和上传

```bash
# 编译项目
pio run

# 上传到 ESP32
pio run --target upload

# 监控串口输出
pio device monitor
```

## 功能特性

### 1. 语音交互
- 支持语音输入和输出
- 实时音频处理
- 自动重连机制

### 2. OTA 固件更新
- 自动检查固件更新
- 支持远程固件升级
- 版本管理

### 3. 设备管理
- 唯一设备标识
- 自动注册到服务器
- 状态监控

## 使用说明

### 1. 首次使用
1. 修改 `src/config.h` 中的 WiFi 和服务器配置
2. 编译并上传固件到 ESP32
3. 打开串口监视器查看连接状态

### 2. 语音交互
- 设备连接成功后会自动开始监听
- 说话后等待服务器响应
- 服务器会通过扬声器播放回复

### 3. 固件更新
- 设备会定期检查固件更新
- 有新版本时会自动下载并更新
- 更新完成后自动重启

## 故障排除

### 1. WiFi 连接失败
- 检查 WiFi 名称和密码
- 确保信号强度足够
- 检查网络配置

### 2. 服务器连接失败
- 检查服务器 IP 地址和端口
- 确保服务器正在运行
- 检查网络防火墙设置

### 3. 音频问题
- 检查 I2S 连接
- 确认音频模块工作正常
- 调整音量设置

### 4. 编译错误
- 确保安装了所有依赖库
- 检查 PlatformIO 配置
- 更新库版本

## 开发说明

### 项目结构
```
esp32-client/
├── platformio.ini    # PlatformIO 配置
├── src/
│   ├── main.cpp      # 主程序
│   └── config.h      # 配置文件
└── README.md         # 说明文档
```

### 主要库依赖
- `WebSocketsClient`: WebSocket 通信
- `ArduinoJson`: JSON 处理
- `ESP8266Audio`: 音频处理
- `ESP32AnalogRead/Write`: 模拟输入输出

### 扩展功能
- 添加按钮控制
- 增加 LED 状态指示
- 支持多语言
- 添加传感器集成

## 许可证

MIT License 