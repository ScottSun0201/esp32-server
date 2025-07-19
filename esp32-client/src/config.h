#ifndef CONFIG_H
#define CONFIG_H

// WiFi 配置
#define WIFI_SSID "你的WiFi名称"
#define WIFI_PASSWORD "你的WiFi密码"

// 服务器配置
#define SERVER_HOST "192.168.1.100"  // 你的服务器IP地址
#define SERVER_PORT 8000
#define SERVER_WS_PATH "/xiaozhi/v1/"

// 设备配置
#define DEVICE_ID "AA:BB:CC:DD:EE:FF"  // 设备唯一标识
#define CLIENT_ID "1111111"             // 客户端ID
#define DEVICE_NAME "ESP32小智"         // 设备名称

// 音频配置
#define AUDIO_SAMPLE_RATE 24000
#define AUDIO_CHANNELS 1
#define AUDIO_FORMAT "opus"
#define AUDIO_FRAME_DURATION 60

// I2S 引脚配置
#define I2S_WS_PIN 15    // Word Select (LRCLK)
#define I2S_SD_PIN 13    // Serial Data
#define I2S_SCK_PIN 2    // Serial Clock (BCLK)

// 音频缓冲区大小
#define AUDIO_BUFFER_SIZE 1024

// 连接配置
#define RECONNECT_INTERVAL 5000    // 重连间隔 (ms)
#define HEARTBEAT_INTERVAL 30000   // 心跳间隔 (ms)

// OTA 配置
#define OTA_CHECK_INTERVAL 3600000 // OTA 检查间隔 (1小时)
#define OTA_SERVER_URL "http://" SERVER_HOST ":8080/api/ota/"

#endif // CONFIG_H 