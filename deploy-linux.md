# Linux 部署指南

## 文件准备

你已经生成了部署包：`xiaozhi-server-linux.tar.gz`

## 部署步骤

### 1. 上传到 Linux 服务器

```bash
# 使用 scp 上传
scp xiaozhi-server-linux.tar.gz user@your-server:/tmp/

# 或使用 rsync
rsync -avz xiaozhi-server-linux.tar.gz user@your-server:/tmp/
```

### 2. 在 Linux 服务器上解压

```bash
# 登录到服务器
ssh user@your-server

# 解压文件
cd /tmp
tar -xzf xiaozhi-server-linux.tar.gz
cd xiaozhi-server-linux
```

### 3. 配置服务器

编辑 `config.yaml` 文件：

```yaml
# 服务器基础配置
server:
  ip: 0.0.0.0
  port: 8000
  token: "1234567890"

# Web界面配置
web:
  enabled: true
  port: 8080
  websocket: ws://你的服务器IP:8000
  vision: http://你的服务器IP:8080/api/vision
```

### 4. 安装和启动

#### 方法一：使用安装脚本（推荐）

```bash
# 安装为系统服务
sudo ./install.sh

# 启动服务
sudo systemctl start xiaozhi

# 查看状态
sudo systemctl status xiaozhi

# 查看日志
sudo journalctl -u xiaozhi -f
```

#### 方法二：手动运行

```bash
# 直接运行
./xiaozhi-server-linux

# 后台运行
nohup ./xiaozhi-server-linux > logs/server.log 2>&1 &

# 查看日志
tail -f logs/server.log
```

### 5. 验证部署

#### 检查服务状态
```bash
# 检查端口监听
netstat -tlnp | grep :8000
netstat -tlnp | grep :8080

# 检查进程
ps aux | grep xiaozhi
```

#### 测试 API 接口
```bash
# 测试 WebSocket 服务
curl http://localhost:8080/api/devices

# 测试推送接口
curl -X POST http://localhost:8080/api/push \
  -H "Content-Type: application/json" \
  -d '{"id":"1111111","text":"测试消息"}'
```

## 配置说明

### 1. 防火墙配置

```bash
# Ubuntu/Debian
sudo ufw allow 8000
sudo ufw allow 8080

# CentOS/RHEL
sudo firewall-cmd --permanent --add-port=8000/tcp
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload
```

### 2. 系统服务管理

```bash
# 启动服务
sudo systemctl start xiaozhi

# 停止服务
sudo systemctl stop xiaozhi

# 重启服务
sudo systemctl restart xiaozhi

# 查看状态
sudo systemctl status xiaozhi

# 开机自启
sudo systemctl enable xiaozhi

# 禁用开机自启
sudo systemctl disable xiaozhi
```

### 3. 日志管理

```bash
# 查看实时日志
sudo journalctl -u xiaozhi -f

# 查看最近日志
sudo journalctl -u xiaozhi -n 100

# 查看错误日志
sudo journalctl -u xiaozhi -p err
```

## 故障排除

### 1. 服务启动失败

```bash
# 检查配置文件
cat config.yaml

# 检查端口占用
sudo netstat -tlnp | grep :8000

# 检查权限
ls -la xiaozhi-server-linux
```

### 2. 连接问题

```bash
# 检查网络连接
ping your-server-ip

# 检查防火墙
sudo ufw status
```

### 3. 性能优化

```bash
# 查看资源使用
top -p $(pgrep xiaozhi-server-linux)

# 查看内存使用
free -h

# 查看磁盘使用
df -h
```

## 更新部署

### 1. 停止服务
```bash
sudo systemctl stop xiaozhi
```

### 2. 备份配置
```bash
sudo cp /opt/xiaozhi-server/config.yaml /opt/xiaozhi-server/config.yaml.backup
```

### 3. 更新文件
```bash
# 上传新版本
scp new-xiaozhi-server-linux.tar.gz user@your-server:/tmp/

# 解压并替换
cd /tmp
tar -xzf new-xiaozhi-server-linux.tar.gz
sudo cp -r xiaozhi-server-linux/* /opt/xiaozhi-server/
```

### 4. 重启服务
```bash
sudo systemctl start xiaozhi
sudo systemctl status xiaozhi
```

## 监控和维护

### 1. 创建监控脚本
```bash
#!/bin/bash
# 检查服务状态
if ! systemctl is-active --quiet xiaozhi; then
    echo "小智服务已停止，正在重启..."
    systemctl restart xiaozhi
fi
```

### 2. 设置定时任务
```bash
# 编辑 crontab
crontab -e

# 添加监控任务（每5分钟检查一次）
*/5 * * * * /path/to/monitor-script.sh
```

## 安全建议

1. **修改默认端口**：避免使用默认端口 8000 和 8080
2. **设置防火墙**：只开放必要端口
3. **使用 HTTPS**：生产环境建议使用 SSL 证书
4. **定期备份**：备份配置文件和日志
5. **监控日志**：定期检查错误日志

## 联系信息

如有问题，请检查：
- 服务状态：`sudo systemctl status xiaozhi`
- 错误日志：`sudo journalctl -u xiaozhi -p err`
- 配置文件：`cat /opt/xiaozhi-server/config.yaml` 