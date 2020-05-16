# Tunnel SSH Client

SSH client capable of automatically using a proxy


## Git Over SSH integration

Using the following command, we can use Git SSH traffic to pass through the proxy after turning on the proxy:

```shell
TUNNEL_DEBUG=1 git-tunnel clone git@github.com:bailbuild/tunnelssh.git
```