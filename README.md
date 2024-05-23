# nsh

The [Noisy Sockets](https://github.com/noisysockets/noisysockets) CLI.

The Noisy Sockets CLI can be used to configure and manage userspace WireGuard connections. Over time it will grow to include a collection of WireGuard powered apps.

The first of these apps is the Noisy Sockets Shell which is a secure, remote shell that uses WireGuard for authentication and encryption. The shell is accessible via a terminal or a web browser.

## Screenshot

<img src="https://github.com/noisysockets/nsh/raw/main/docs/terminal_screenshot.png" width="450" alt="Terminal Screenshot" />

*Showing a terminal session using the in-browser client.*

## Getting Started

### Initialize Configuration

The `config init` command will generate a new private key and populate the
configuration file with the provided options.

```sh
nsh config init -c server.yaml -n server --listen-port=51820 --ip=100.64.0.1
nsh config init -c client.yaml -n client --listen-port=51821 --ip=100.64.0.2
```

### Add Peers

The server and client will need to be aware of each other in order to establish 
a connection. The `peer add` command will add a peer to the configuration file.

*Note: The client will need to know the servers endpoint in order to establish a connection.*

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

### Start Server

In another tab, start the server.

```sh
nsh serve -c server.yaml --enable-dns --enable-shell
```

### Connect to Server

#### Using CLI

You can connect to the shell server by its hostname, or the IP address. In the 
following example, we will connect to the server using the hostname.

```sh
nsh shell -c client.yaml server
```

#### Using Browser

##### Create WireGuard interface

```sh
sudo nsh config export -c client.yaml -o /etc/wireguard/nsh0.conf
sudo wg-quick up nsh0
```

##### Configure DNS

To resolve the server hostname, you will need to configure the DNS resolver.

For resolvconf (Debian/Ubuntu):

```sh
sudo grep -q "nameserver 100.64.0.1" /etc/resolvconf/resolv.conf.d/tail || echo "nameserver 100.64.0.1" | sudo tee -a /etc/resolvconf/resolv.conf.d/tail > /dev/null
sudo resolvconf -u
```

For systemd-resolved (RHEL/CentOS):

```sh
sudo resolvectl --interface=nsh0 --set-dns=100.64.0.1 --set-domain=my.nzzy.net.
```

##### Open Shell

```sh
xdg-open http://server.my.nzzy.net/shell/
```

The network domain can be customized via the `--domain` flag of the `config init` command.

## Credits

Noisy Sockets is based on code originally from the [wireguard-go](https://git.zx2c4.com/wireguard-go) project by Jason A. Donenfeld.

WireGuard is a registered trademark of Jason A. Donenfeld.