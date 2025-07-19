#!/bin/bash

# Linux 打包脚本
echo "开始打包 Linux 版本..."

# 创建打包目录
PACKAGE_DIR="xiaozhi-server-linux"
rm -rf $PACKAGE_DIR
mkdir -p $PACKAGE_DIR

# 复制配置文件
echo "复制配置文件..."
cp config.yaml $PACKAGE_DIR/
cp -r music $PACKAGE_DIR/ 2>/dev/null || mkdir -p $PACKAGE_DIR/music

# 创建必要的目录
mkdir -p $PACKAGE_DIR/logs
mkdir -p $PACKAGE_DIR/tmp
mkdir -p $PACKAGE_DIR/ota_bin

# 复制文档
cp README.md $PACKAGE_DIR/ 2>/dev/null || echo "# 小智服务器" > $PACKAGE_DIR/README.md
cp LICENSE $PACKAGE_DIR/ 2>/dev/null || echo "MIT License" > $PACKAGE_DIR/LICENSE

# 创建启动脚本
cat > $PACKAGE_DIR/start.sh << 'EOF'
#!/bin/bash

# 小智服务器启动脚本
echo "启动小智服务器..."

# 检查配置文件
if [ ! -f "config.yaml" ]; then
    echo "错误: 找不到 config.yaml 配置文件"
    exit 1
fi

# 创建必要目录
mkdir -p logs
mkdir -p tmp
mkdir -p ota_bin

# 启动服务器
./xiaozhi-server-linux
EOF

# 创建停止脚本
cat > $PACKAGE_DIR/stop.sh << 'EOF'
#!/bin/bash

# 停止小智服务器
echo "停止小智服务器..."

# 查找并杀死进程
PID=$(ps aux | grep xiaozhi-server-linux | grep -v grep | awk '{print $2}')
if [ ! -z "$PID" ]; then
    echo "找到进程 PID: $PID"
    kill $PID
    echo "已发送停止信号"
else
    echo "未找到运行中的小智服务器进程"
fi
EOF

# 创建服务文件
cat > $PACKAGE_DIR/xiaozhi.service << 'EOF'
[Unit]
Description=Xiaozhi Server
After=network.target

[Service]
Type=simple
User=xiaozhi
WorkingDirectory=/opt/xiaozhi-server
ExecStart=/opt/xiaozhi-server/xiaozhi-server-linux
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 创建安装脚本
cat > $PACKAGE_DIR/install.sh << 'EOF'
#!/bin/bash

# 小智服务器安装脚本
echo "安装小智服务器..."

# 检查是否为 root 用户
if [ "$EUID" -ne 0 ]; then
    echo "请使用 sudo 运行此脚本"
    exit 1
fi

# 创建用户
useradd -r -s /bin/false xiaozhi 2>/dev/null || echo "用户 xiaozhi 已存在"

# 创建安装目录
INSTALL_DIR="/opt/xiaozhi-server"
mkdir -p $INSTALL_DIR

# 复制文件
cp -r * $INSTALL_DIR/
chown -R xiaozhi:xiaozhi $INSTALL_DIR
chmod +x $INSTALL_DIR/xiaozhi-server-linux
chmod +x $INSTALL_DIR/start.sh
chmod +x $INSTALL_DIR/stop.sh

# 安装服务
cp xiaozhi.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable xiaozhi

echo "安装完成！"
echo "启动服务: sudo systemctl start xiaozhi"
echo "查看状态: sudo systemctl status xiaozhi"
echo "查看日志: sudo journalctl -u xiaozhi -f"
EOF

# 创建卸载脚本
cat > $PACKAGE_DIR/uninstall.sh << 'EOF'
#!/bin/bash

# 小智服务器卸载脚本
echo "卸载小智服务器..."

# 检查是否为 root 用户
if [ "$EUID" -ne 0 ]; then
    echo "请使用 sudo 运行此脚本"
    exit 1
fi

# 停止服务
systemctl stop xiaozhi 2>/dev/null
systemctl disable xiaozhi 2>/dev/null

# 删除服务文件
rm -f /etc/systemd/system/xiaozhi.service
systemctl daemon-reload

# 删除安装目录
rm -rf /opt/xiaozhi-server

# 删除用户（可选）
read -p "是否删除 xiaozhi 用户? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    userdel xiaozhi 2>/dev/null
    echo "已删除用户 xiaozhi"
fi

echo "卸载完成！"
EOF

# 设置权限
chmod +x $PACKAGE_DIR/start.sh
chmod +x $PACKAGE_DIR/stop.sh
chmod +x $PACKAGE_DIR/install.sh
chmod +x $PACKAGE_DIR/uninstall.sh

# 创建压缩包
echo "创建压缩包..."
tar -czf xiaozhi-server-linux.tar.gz $PACKAGE_DIR

echo "打包完成！"
echo "文件: xiaozhi-server-linux.tar.gz"
echo ""
echo "部署说明:"
echo "1. 上传 xiaozhi-server-linux.tar.gz 到 Linux 服务器"
echo "2. 解压: tar -xzf xiaozhi-server-linux.tar.gz"
echo "3. 进入目录: cd xiaozhi-server-linux"
echo "4. 安装: sudo ./install.sh"
echo "5. 启动: sudo systemctl start xiaozhi"
echo ""
echo "或者手动运行:"
echo "1. 解压后直接运行: ./xiaozhi-server-linux"
echo "2. 后台运行: nohup ./xiaozhi-server-linux > logs/server.log 2>&1 &" 