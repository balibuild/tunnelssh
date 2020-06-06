# TunnelSSH 一个机智的 SSH 客户端

我们逐渐依赖上了高速互联网，每一天，浏览社交媒体，玩游戏，看视频，收发邮件，远程视频，都离不开网络。对于程序员而言，网络能解决我们的协调开发问题，我们可以通过分布式版本控制系统远程与人协作，但这一切建立在稳定的网络连接的基础上。

网络又是复杂的，多变的，硬件可能出现故障，企业可能添加企业策略限制，运营商可能相应链路带宽不足，也许有其他因素让网络连接无法建立或者中断。所以也就有了特定的机制绕过这些限制，建立特殊的网络连接。

当企业隔离外网时，只允许特定的机器与外网连接，这个时候，我们可以通过代理经由特定的机器建立连接。在 HTTP 协议中，我们可以使用 [CONNECT](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/CONNECT) 方法建立隧道，从而实现网络的互通。

在 OpenSSH 中，我们可以使用 ProxyCommand 设置让 OpenSSH 客户端经由 Proxy command 建立连接，但配置是固定的，一旦用户网络情况调整，比如关闭代理就需要重新设置，这有一些麻烦，于是我便决定编写 TunnelSSH 解决这个问题。实际上开发一个全功能的 SSH 客户端稍显麻烦，而 OpenSSH 可以设置 ProxyCommand，因此在 TunnelSSH 中，我们还提供了 necat 命令，当 OpenSSH 通过设置 ProxyCommand 为 netcat 时，netcat 将解决自动按照系统设置建立网络连接。除此之外，在 TunnelSSH 项目中，我们还提供了 TunnelSSH 的 git 包装，SSH AskPass Utility 工具，这就是 TunnelSSH 的基本内容。

## TunnelSSH 介绍

TunnelSSH 既是此项目的名称也是其 SSH 客户端的名称，TunnelSSH 借鉴了 [tatsushid/minssh](https://github.com/tatsushid/minssh) 相关代码，是一个有限的 SSH 客户端，主要是为了解决作者在使用 SSH 协议管理 git 存储库时，无法使用系统代理的问题。一个简单的截图如下：

![](./docs/images/snapshot.png)

请注意，在这里我使用了 [baulk](https://github.com/baulk/baulk.git) 安装了 TunnelSSH，并且启动了 baulk 终端环境，因此可以直接使用 `git tunnel -V push` 将存储库推送到 Github 上。

TunnelSSH 的实现没有多少技术含量，简单而言，就是基于 Golang 的 TunnelSSH 在建立 SSH 连接时，如果代理可用就使用代理建立 `net.Conn`，在此基础上建立 SSH 连接，然后再 git 包装命令中设置 `GIT_SSH` 和 `GIT_SSH_VARIANT` 环境变量，支持解析 git 使用的 SSH 命令行参数，这就行了。

## TunnelSSH NetCat

TunnelSSH NetCat 出现的目的很简单，既然 TunnelSSH 暂时不想成为一个强大的 SSH 客户端，那么，NetCat 可以帮助 OpenSSH 变得更加强大，NetCat 命令和 TunnelSSH 使用了同样的 `tunnel` 包，能够读取系统配置（Windows 注册表项）和环境变量，通过代理建立网络连接，当代理不可用时，回退到直接连接，未开启代理时，建立直接连接，同样也非常简单。

## git-tunnel TunnelSSH 的 Git 包装

git-tunnel 通过检查环境配置初始化 GIT_SSH 设置和环境变量后启动 git 相应命令，支持相关操作，处理能够使 git 使用 TunnelSSH 建立 SSH 连接，也可以在 Windows 上设置环境变量，让 Git Over HTTP 通过代理建立网络连接，这样可以无需用户通过运行 `git config` 设置存储库的网络连接方式。

类似命令操作如下（这里使用了 [baulk](https://github.com/baulk/baulk.git) 工具将 git/git-tunnel 所在目录加入到环境变量中了）：

```shell
git tunnel clone git@github.com:balibuild/tunnelssh.git
cd tunnelssh
git tunnel -V fetch
```

未开代理的时候，你也可以这样使用，没有任何其他影响。在 Linux/macOS 可以设置 Bash/Zsh Alias，在 PowerShell 中同样可以设置别名。

## SSH AskPass Utility

在 TunnelSSH 中，我们使用 ssh-askpass 命令实现在标准输入重定向后的密码信息读取和提示确认功能，在这里我们使用了 Windows 现代的密码凭据输入界面，如下图：

![](./docs/images/ssh-askpass.png)

需要注意的是，由于使用了 Vista 的 GUI 风格，因此程序需要嵌入应用程序清单，所以这个项目需要使用 [bali](https://github.com/balibuild/bali) 进行构建。

## 已知问题

由于作者无 macOS，并未测试读取系统代理设置，因此，在 macOS 上只能读取环境变量中的设置。如果有人想要帮助作者实现该功能，欢迎提交 PR。

## 其他

欢迎用户提交 PR

## 感谢

ssh RunInteractive borrows from [tatsushid/minssh](https://github.com/tatsushid/minssh). Thanks here
