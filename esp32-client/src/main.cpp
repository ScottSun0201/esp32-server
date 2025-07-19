#include <Arduino.h>
#include <WiFi.h>
#include <WebSocketsClient.h>
#include <ArduinoJson.h>
#include <SPIFFS.h>
#include <HTTPClient.h>
#include <Update.h>
#include <driver/i2s.h>

// 配置参数
const char* ssid = "你的WiFi名称";
const char* password = "你的WiFi密码";
const char* serverHost = "你的服务器IP";
const int serverPort = 8000;
const char* deviceId = "AA:BB:CC:DD:EE:FF";
const char* clientId = "1111111";

// 全局变量
WebSocketsClient webSocket;
bool isConnected = false;
bool isListening = false;
String currentSessionId = "";

// I2S 音频配置
#define I2S_WS 15
#define I2S_SD 13
#define I2S_SCK 2
#define I2S_PORT I2S_NUM_0
#define BUFFER_SIZE 1024

// 音频缓冲区
int16_t audioBuffer[BUFFER_SIZE];

// 函数声明
void connectToWiFi();
void connectToServer();
void webSocketEvent(WStype_t type, uint8_t * payload, size_t length);
void sendHelloMessage();
void handleServerMessage(String message);
void startListening();
void stopListening();
void processAudio();
void checkOTAUpdate();

void setup() {
  Serial.begin(115200);
  Serial.println("小智 ESP32 客户端启动...");

  // 初始化 SPIFFS
  if (!SPIFFS.begin(true)) {
    Serial.println("SPIFFS 初始化失败");
    return;
  }

  // 初始化 I2S
  i2s_config_t i2s_config = {
    .mode = (i2s_mode_t)(I2S_MODE_MASTER | I2S_MODE_TX | I2S_MODE_RX),
    .sample_rate = 24000,
    .bits_per_sample = I2S_BITS_PER_SAMPLE_16BIT,
    .channel_format = I2S_CHANNEL_FMT_ONLY_LEFT,
    .communication_format = I2S_COMM_FORMAT_STAND_I2S,
    .intr_alloc_flags = ESP_INTR_FLAG_LEVEL1,
    .dma_buf_count = 8,
    .dma_buf_len = BUFFER_SIZE,
    .use_apll = false,
    .tx_desc_auto_clear = true,
    .fixed_mclk = 0
  };

  i2s_pin_config_t pin_config = {
    .bck_io_num = I2S_SCK,
    .ws_io_num = I2S_WS,
    .data_out_num = I2S_SD,
    .data_in_num = I2S_SD
  };

  i2s_driver_install(I2S_PORT, &i2s_config, 0, NULL);
  i2s_set_pin(I2S_PORT, &pin_config);

  // 连接 WiFi
  connectToWiFi();

  // 检查 OTA 更新
  checkOTAUpdate();

  // 连接服务器
  connectToServer();
}

void loop() {
  webSocket.loop();
  
  // 处理音频
  if (isListening) {
    processAudio();
  }

  // 定期检查连接状态
  static unsigned long lastCheck = 0;
  if (millis() - lastCheck > 30000) { // 每30秒检查一次
    if (!isConnected) {
      Serial.println("重新连接服务器...");
      connectToServer();
    }
    lastCheck = millis();
  }
}

void connectToWiFi() {
  Serial.print("连接 WiFi: ");
  Serial.println(ssid);
  
  WiFi.begin(ssid, password);
  
  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }
  
  Serial.println();
  Serial.print("WiFi 连接成功，IP: ");
  Serial.println(WiFi.localIP());
}

void connectToServer() {
  String wsUrl = "ws://";
  wsUrl += serverHost;
  wsUrl += ":";
  wsUrl += String(serverPort);
  wsUrl += "/xiaozhi/v1/?device-id=";
  wsUrl += deviceId;
  wsUrl += "&client-id=";
  wsUrl += clientId;
  
  Serial.print("连接服务器: ");
  Serial.println(wsUrl);
  
  webSocket.begin(serverHost, serverPort, "/xiaozhi/v1/?device-id=" + String(deviceId) + "&client-id=" + String(clientId));
  webSocket.onEvent(webSocketEvent);
  webSocket.setReconnectInterval(5000);
}

void webSocketEvent(WStype_t type, uint8_t * payload, size_t length) {
  switch(type) {
    case WStype_DISCONNECTED:
      Serial.println("WebSocket 断开连接");
      isConnected = false;
      break;
      
    case WStype_CONNECTED:
      Serial.println("WebSocket 连接成功");
      isConnected = true;
      sendHelloMessage();
      break;
      
    case WStype_TEXT:
      handleServerMessage(String((char*)payload));
      break;
      
    case WStype_BIN:
      // 处理二进制音频数据
      if (length > 0) {
        // 将音频数据写入 I2S
        size_t bytesWritten = 0;
        i2s_write(I2S_PORT, payload, length, &bytesWritten, 100);
      }
      break;
      
    case WStype_ERROR:
      Serial.println("WebSocket 错误");
      break;
  }
}

void sendHelloMessage() {
  DynamicJsonDocument doc(1024);
  doc["type"] = "hello";
  doc["device_mac"] = deviceId;
  doc["device_name"] = "ESP32小智";
  doc["token"] = "esp32_token";
  
  JsonObject features = doc.createNestedObject("features");
  features["mcp"] = true;
  
  JsonObject audioParams = doc.createNestedObject("audio_params");
  audioParams["format"] = "opus";
  audioParams["sample_rate"] = 24000;
  audioParams["channels"] = 1;
  audioParams["frame_duration"] = 60;
  
  String message;
  serializeJson(doc, message);
  
  Serial.print("发送 hello 消息: ");
  Serial.println(message);
  
  webSocket.sendTXT(message);
}

void handleServerMessage(String message) {
  Serial.print("收到服务器消息: ");
  Serial.println(message);
  
  DynamicJsonDocument doc(2048);
  DeserializationError error = deserializeJson(doc, message);
  
  if (error) {
    Serial.print("JSON 解析失败: ");
    Serial.println(error.c_str());
    return;
  }
  
  String type = doc["type"];
  
  if (type == "hello") {
    // 服务器握手响应
    currentSessionId = doc["session_id"].as<String>();
    Serial.print("会话ID: ");
    Serial.println(currentSessionId);
    
  } else if (type == "listen") {
    // 开始/停止监听
    String state = doc["state"];
    if (state == "start") {
      startListening();
    } else if (state == "stop") {
      stopListening();
    }
    
  } else if (type == "chat") {
    // 聊天消息
    String text = doc["text"];
    Serial.print("收到文本: ");
    Serial.println(text);
    
  } else if (type == "audio") {
    // 音频数据
    Serial.println("收到音频数据");
    
  } else if (type == "abort") {
    // 中止操作
    stopListening();
  }
}

void startListening() {
  Serial.println("开始监听...");
  isListening = true;
}

void stopListening() {
  Serial.println("停止监听...");
  isListening = false;
}

void processAudio() {
  // 从麦克风读取音频数据
  size_t bytesRead = 0;
  esp_err_t result = i2s_read(I2S_PORT, audioBuffer, BUFFER_SIZE * 2, &bytesRead, 100);
  
  if (result == ESP_OK && bytesRead > 0) {
    // 发送音频数据到服务器
    webSocket.sendBIN((uint8_t*)audioBuffer, bytesRead);
  }
}

void checkOTAUpdate() {
  Serial.println("检查 OTA 更新...");
  
  HTTPClient http;
  String url = "http://";
  url += serverHost;
  url += ":8080/api/ota/";
  
  http.begin(url);
  http.addHeader("Content-Type", "application/json");
  http.addHeader("device-id", deviceId);
  
  DynamicJsonDocument doc(512);
  doc["application"]["version"] = "1.0.0";
  
  String jsonString;
  serializeJson(doc, jsonString);
  
  int httpCode = http.POST(jsonString);
  
  if (httpCode == 200) {
    String payload = http.getString();
    Serial.print("OTA 响应: ");
    Serial.println(payload);
    
    // 解析响应，检查是否需要更新
    DynamicJsonDocument responseDoc(1024);
    deserializeJson(responseDoc, payload);
    
    String serverVersion = responseDoc["firmware"]["version"];
    String firmwareUrl = responseDoc["firmware"]["url"];
    
    Serial.print("服务器版本: ");
    Serial.println(serverVersion);
    Serial.print("固件URL: ");
    Serial.println(firmwareUrl);
    
    // 这里可以添加版本比较和下载逻辑
  } else {
    Serial.print("OTA 检查失败，HTTP 代码: ");
    Serial.println(httpCode);
  }
  
  http.end();
} 