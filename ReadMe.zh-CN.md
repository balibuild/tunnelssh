# TunnelSSH 一个机智的 SSH 客户端

我们逐渐依赖上了高速互联网，每一天，浏览社交媒体，玩游戏，看视频，收发邮件，远程视频，都离不开网络。对于程序员而言，网络能解决我们的协调开发问题，我们可以通过分布式版本控制系统远程与人协作，但这一切建立在稳定的网络连接的基础上。

网络又是复杂的，多变的，硬件可能出现故障，企业可能添加企业策略限制，运营商可能相应链路带宽不足，也许有其他因素让网络连接无法建立或者中断。所以也就有了特定的机制绕过这些限制，建立特殊的网络连接。

当企业隔离外网时，只允许特定的机器与外网连接，这个时候，我们可以通过代理经由特定的机器建立连接。在 HTTP 协议中，我们可以使用 [CONNECT](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/CONNECT) 方法建立隧道，从而实现网络的互通。

在 OpenSSH 中，我们可以使用 ProxyCommand 设置让 OpenSSH 客户端经由 Proxy command 建立连接，但配置是固定的，一旦用户网络情况调整，比如关闭代理就需要重新设置，这有一些麻烦，于是我便决定编写 TunnelSSH 解决这个问题。实际上开发一个全功能的 SSH 客户端稍显麻烦，而 OpenSSH 可以设置 ProxyCommand，因此在 TunnelSSH 中，我们还提供了 necat 命令，当 OpenSSH 通过设置 ProxyCommand 为 netcat 时，netcat 将解决自动按照系统设置建立网络连接。除此之外，在 TunnelSSH 项目中，我们还提供了 TunnelSSH 的 git 包装，SSH AskPass Utility 工具，这就是 TunnelSSH 的基本内容。

## TunnelSSH 介绍

TunnelSSH 既是此项目的名称也是其 SSH 客户端的名称，TunnelSSH 借鉴了 [tatsushid/minssh](https://github.com/tatsushid/minssh) 相关代码，是一个有限的 SSH 客户端，主要是为了解决作者在使用 SSH 协议管理 git 存储库时，无法使用系统代理的问题。一个简单的截图如下：

![](./docs/images/snapshot.png)

