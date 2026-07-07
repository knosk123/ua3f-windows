# ua3f-win

Windows 11 amd64 上的 UA3F 重写版，生成单个 `ua3f-win.exe`。

## 使用

双击 `ua3f-win.exe` 后会自动请求管理员权限，并用 Microsoft Edge 打开本地控制页面。

页面可配置：

- UA 预设：`wechat` / `pc`
- TTL
- HTTP 端口列表，例如 `80,8080`
- 日志级别：`info` / `debug` / `warn`

不再支持用户自定义 UA。`wechat` 当前为：

```text
Mozilla/5.0 (Linux; Android 15; RMX6688 Build/AP3A.240617.008; wv) AppleWebKit/537.36
```

## 命令行

```powershell
.\ua3f-win.exe -ua wechat
.\ua3f-win.exe -ua pc
.\ua3f-win.exe -ua wechat -log debug
.\ua3f-win.exe -ttl 64 -ports "80,8080"
```

`-ua` 只接受 `wechat` 或 `pc`。

## 限制

- 只处理 IPv4 明文 HTTP。
- 不处理 HTTPS 内容。
- 不做完整 TCP 流重组。
- UA 只在已有 `User-Agent` 请求头内等长覆盖，不插入缺失的 `User-Agent` 行。
- WinDivert 需要管理员权限。

## 编译

```powershell
cd "path\to\ua3f-win"
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:GOPROXY = "https://goproxy.cn,direct"
go mod tidy
go build -ldflags "-s -w" -o ua3f-win.exe
```

也可以运行：

```powershell
powershell -ExecutionPolicy Bypass -File .\build.ps1
```
