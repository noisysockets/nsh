# Router

With Noisy Sockets you can run a userspace router/exit node, this userspace 
router can be used to forward traffic between the WireGuard network and the 
internet (or other private networks).

## Features

* TCP/UDP Forwarding
* Limited ICMPv4/ICMPv6 forwarding (ping)
* NAT64 (IPv6 to IPv4 translation)
* Recursive DNS Resolver

## Getting Started

### Initialize Configuration

The `config init` command will generate a new private key and populate the
configuration file with the provided options.

```sh
nsh config init -c router.yaml -n router
nsh config init -c client.yaml -n client --ip=$(nsh config show -c router.yaml 'next(.ips[0])')
```

### Add Peers

The router and client will need to be aware of each other in order to establish
a connection. The `peer add` command will add a peer to the configuration file.

*Note: The client will need to know the routers public endpoint in order to 
establish a connection.*

```sh
nsh peer add -c router.yaml \
  --name=client \
  --public-key=$(nsh config show -c client.yaml 'public(.privateKey)') \
  --ip=$(nsh config show -c client.yaml '.ips[0]')

nsh peer add -c client.yaml \
  --name=router \
  --public-key=$(nsh config show -c router.yaml 'public(.privateKey)') \
  --endpoint=$(nsh config show -c router.yaml '"localhost:" + (.listenPort|tostring)') \
  --ip=$(nsh config show -c router.yaml '.ips[0]')
```

### Add Route

The client will need to know where to send internet bound traffic (eg. which 
peer is acting as a router).

```sh
nsh route add -c client.yaml --destination=::/0 --via=router
```

### Start Router

In another terminal window, start the router.

```sh
nsh up -c router.yaml --enable-router --enable-dns
```

*Note: Userspace routers do not require any elevated permissions.*

### Use the Router

#### Export WireGuard Configuration

```sh
nsh config export -c client.yaml --stripped | sudo tee /etc/wireguard/nsh0.conf > /dev/null
```

#### Setup Network Namespace

To avoid conflicts with the host network, for this example we will connect to
the router using a network namespace.

```sh
sudo mkdir -p /etc/netns/nsh-client-ns
echo -e "nameserver $(nsh config show -c router.yaml '.ips[0]')\nsearch my.nzzy.net.\n" | sudo tee /etc/netns/nsh-client-ns/resolv.conf > /dev/null
sudo ip netns add nsh-client-ns
sudo ip link add nsh0 type wireguard
sudo ip link set dev nsh0 mtu 1280
sudo ip link set nsh0 netns nsh-client-ns
sudo ip netns exec nsh-client-ns wg setconf nsh0 /etc/wireguard/nsh0.conf
sudo ip -n nsh-client-ns addr add "$(nsh config show -c client.yaml '.ips[0]')/64" dev nsh0
sudo ip -n nsh-client-ns link set nsh0 up
sudo ip -6 -n nsh-client-ns route add default via "$(nsh config show -c router.yaml '.ips[0]')" dev nsh0
```

#### Make a Request

You can now attempt to access the internet using the router as a gateway.

##### IPv6

The following will return the public IP address of the router.

```sh
sudo ip netns exec nsh-client-ns sudo -u $USER curl https://ipv6.icanhazip.com
```

##### IPv4

By default the router implements [NAT64](https://tools.ietf.org/html/rfc6146),
which will allow you to access IPv4 resources on IPv6 only networks.

```sh
sudo ip netns exec nsh-client-ns sudo -u $USER curl https://ipv4.icanhazip.com
```

#### Cleanup

To remove the network namespace and WireGuard interface when you are finished.

```sh
sudo ip -n nsh-client-ns link del nsh0
sudo ip netns del nsh-client-ns
```