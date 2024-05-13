# nsh

Noisy Sockets 命令行界面（CLI）。

Noisy Sockets CLI 可用于配置和管理用户空间 WireGuard 连接。随着时间的推移，它将逐渐扩展，包括一系列由 WireGuard 提供支持的应用程序。

这些应用程序中的第一个是 [Noisy Sockets Shell](https://github.com/noisysockets/shell)，这是一个使用 WireGuard 进行认证和加密的安全远程 shell。该 shell 可以通过终端或 Web 浏览器访问。

## 截图

<img src="https://github.com/noisysockets/nsh/raw/main/docs/terminal_screenshot.png"  width="450" alt="终端截图" />

*显示使用浏览器内客户端的终端会话。*

## 开始使用

### 初始化配置

`config init` 命令将生成一个新的私钥，并将配置文件填充提供选项。

```sh
nsh config init -c server.yaml -n server --listen-port=51820 --ip=172.21.248.1
nsh config init -c client.yaml -n client --listen-port=51821 --ip=172.21.248.2
```

### 添加对等方

为了建立连接，服务器和客户端需要相互了解。`peer add` 命令将向配置文件添加一个对等方。

*注意：客户端需要知道服务器的端点才能建立连接。*

```sh
nsh peer add -c server.yaml \
  --name=client \
  --public-key=$(nsh config show -c client.yaml 'public(.privateKey)') \
  --ip=$(nsh config show -c client.yaml '.ips[0]')

nsh peer add -c client.yaml \
  --name=server \
  --public-key=$(nsh config show -c server.yaml 'public(.privateKey)') \
  --endpoint=$(nsh config show -c server.yaml '"localhost:" + (.listenPort|tostring)') \
  --ip=$(nsh config show -c server.yaml '.ips[0]')
```

### 启动服务器

在另一个标签页中，启动服务器。

```sh
nsh shell serve -c server.yaml
```

### 连接到服务器

#### 使用 CLI

您可以通过主机名或 IP 地址连接到 shell 服务器。在以下示例中，我们将使用主机名连接到服务器。

```sh
nsh shell connect -c client.yaml server
```

#### 使用浏览器

当使用 wg 内核模块时，您需要使用 IP 地址连接到 shell 服务器（因为我们尚未实现集成的 DNS 解析器）。

```sh
sudo nsh config export -c client.yaml -o /etc/wireguard/nsh0.conf
sudo wg-quick up nsh0

xdg-open http://172.21.248.1 
```

## 许可证

Noisy Sockets CLI 根据 [Noisy Sockets 源代码许可证 1.0 (NSSL-1.0)](./LICENSE-NSSL-1.0.txt) 和 [Affero 通用公共许可证 v3 (AGPL-3.0)](./LICENSE-AGPL-3.0.txt) 双重许可。

NSSL-1.0 是一个源代码可用许可证，源自 [FSL-1.1](https://fsl.software)，具有以下变更：

- 代码将在 4 年后而不是 2 年后成为 FOSS（自由及开源软件）。
- 未来的许可证将是 MPL-2.0（一个较弱的版权）而不是 Apache-2.0。

您可以在任一许可证的条款下使用 Noisy Sockets CLI。

## 致谢

Noisy Sockets 基于 Jason A. Donenfeld 的 [wireguard-go](https://git.zx2c4.com/wireguard-go) 项目原始代码。

WireGuard 是 Jason A. Donenfeld 的注册商标。
