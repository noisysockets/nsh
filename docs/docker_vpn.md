# Docker VPN Server

For this example we will assume that you are using IPv6 between the client and
the VPN server. If that's not the case, I highly recommend looking to see if 
your ISP supports IPv6 and putting in the effort to get it enabled.

## Prerequisites

* A server with Docker installed.
* A client with WireGuard installed.

## Server

### Docker

First if you haven't already, you'll need to enable IPv6 on your Docker daemon.

Edit `/etc/docker/daemon.json` and add the following:

```json
{
  "ipv6": true,
  "fixed-cidr-v6": "2001:db8:1::/64",
  "experimental": true,
  "ip6tables": true
}
```

Then restart the Docker daemon:

```sh
sudo systemctl restart docker
```

### Config

Create a Docker volume to store the server configuration:

```sh
docker volume create nsh-config
docker run --rm -v nsh-config:/home/nonroot/.config/nsh/ busybox chown -R 65532:65532 /home/nonroot/.config/nsh/
```

Initialize the server configuration:

```sh
docker run --rm -v nsh-config:/home/nonroot/.config/nsh/ ghcr.io/noisysockets/nsh:latest \
  config init -n server --listen-port=51820
```

Retrieve the servers public key and IP address:

```sh
docker run --rm -v nsh-config:/home/nonroot/.config/nsh/ ghcr.io/noisysockets/nsh:latest \
  config show 'public(.privateKey),.ips[0]'
```

## Client

### Config

To calculate the ip address of the client, take the address of the server, which
will look like `fd01:4796:b537::1`, and add one to the last octet. So in this case
the client would have an address of `fd01:4796:b537::2`.

```sh
nsh config init -n $(hostname) --ip=<CLIENT IP>
```

Add the server to the client configuration, use the IP and public key from earlier.
The public address should be a reachable internet address for the server. If
your server is dual stack, I recommend using the IPv6 address here.

```sh
nsh peer add \
  -n server \
  -k "<SERVER PUBLIC KEY>" \
  -e "[<SERVER PUBLIC ADDRESS>]:51820" \
  --ip=<SERVER IP>
```

Configure the client to use the servers DNS resolver:
  
```sh
nsh dns server add <SERVER IP>::1
```

Now set the clients default route to the internet to go through the VPN server:

```sh
nsh route add -d "::/0" --via=server
```

Retrieve the clients name and public key:

```sh
nsh config show '.name,public(.privateKey),.ips[0]'
```

## Back to the Server

### Add the Client

Add the client to the server configuration:

```sh
docker run --rm -v nsh-config:/home/nonroot/.config/nsh/ ghcr.io/noisysockets/nsh:latest peer add \
  -n "<CLIENT NAME>"  \
  -k "<CLIENT PUBLIC KEY>" \
  --ip=<CLIENT IP>
```

### Start the Server

```sh
docker run -d --name=nsh-server --rm -v nsh-config:/home/nonroot/.config/nsh/ -p51820:51820/udp ghcr.io/noisysockets/nsh:latest up --enable-router --enable-dns
```

## Back to the Client

### Export WireGuard Configuration

```sh
nsh config export --stripped | sudo tee /etc/wireguard/nsh0.conf > /dev/null
```

### Setup Network Namespace

To avoid conflicts with the host network, for this example we will connect to
the server using a network namespace.

```sh
sudo mkdir -p /etc/netns/nsh-client-ns
echo -e "nameserver $(nsh config show '.dns.servers[0]')\nsearch my.nzzy.net.\n" | sudo tee /etc/netns/nsh-client-ns/resolv.conf > /dev/null
sudo ip netns add nsh-client-ns
sudo ip link add nsh0 type wireguard
sudo ip link set dev nsh0 mtu 1280
sudo ip link set nsh0 netns nsh-client-ns
sudo ip netns exec nsh-client-ns wg setconf nsh0 /etc/wireguard/nsh0.conf
sudo ip -n nsh-client-ns addr add "$(nsh config show '.ips[0]')/64" dev nsh0
sudo ip -n nsh-client-ns link set nsh0 up
sudo ip -6 -n nsh-client-ns route add default via $(nsh config show '(.peers[]|select(.name == "server")).ips[0]') dev nsh0
```

### Access the Network Namespace

```sh
sudo ip netns exec nsh-client-ns sudo -u $USER bash
```

Congratulations you are now running an IPv6 only network (with NAT64/DNS64 for legacy IPv4 addresses).

### Cleanup

To remove the network namespace and WireGuard interface when you are finished.

```sh
sudo ip -n nsh-client-ns link del nsh0
sudo ip netns del nsh-client-ns
```