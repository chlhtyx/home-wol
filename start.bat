@echo off
chcp 65001 >nul
cls

echo ==================================
echo   Home WOL 服务启动脚本
echo ==================================
echo.

REM 检查 config.json 是否存在
if not exist "config\config.json" (
    echo ⚠️  配置文件不存在，正在创建...
    copy config\config.json.template config\config.json >nul
    echo ✅ 已创建配置文件: config\config.json
    echo ⚠️  请编辑配置文件添加你的设备信息
    echo.
)

REM 检查 .env 是否存在
if not exist ".env" (
    echo ⚠️  环境变量文件不存在，正在创建...
    copy .env.example .env >nul
    echo ✅ 已创建环境变量文件: .env
    echo ⚠️  请根据你的网络环境修改 .env 文件中的 BROADCAST_ADDRESS
    echo.
    pause
)

echo 🚀 启动服务...
docker-compose up -d

echo.
echo ==================================
echo   服务状态
echo ==================================
docker-compose ps

echo.
echo ==================================
echo   查看日志
echo ==================================
echo 运行以下命令查看日志：
echo   docker-compose logs -f
echo.

echo ==================================
echo   服务地址
echo ==================================
echo WOL 服务: http://localhost:1323
echo.
echo 测试命令：
echo   # 生成验证码
echo   curl "http://localhost:1323/generate_code?mac=YOUR_MAC"
echo.
echo   # 唤醒设备
echo   curl "http://localhost:1323/wol?mac=YOUR_MAC^&auth_code=YOUR_CODE"
echo.

pause
