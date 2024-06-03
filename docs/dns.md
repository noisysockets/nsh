# DNS

Noisy Sockets includes an embedded DNS server that can be used to resolve peer
names and recursively forward DNS queries to the internet. The DNS server is
intended to be used as a central point for name resolution within a WireGuard
network.

Noisy Sockets DNS uses its own [DNS resolver](https://github.com/noisysockets/resolver) 
implementation for recursive DNS resolution. The custom resolver supports a 
variety of upstream protocols including DNS over TLS.

## Features

* DNS over UDP/TCP
* Recursive DNS Resolver
* DNS64 (IPv6 to IPv4 translation)

## Getting Started

### Initialize Configuration

The `config init` command will generate a new private key and populate the
configuration file with the provided options.

```sh
nsh config init -c resolver.yaml -n resolver
nsh config init -c client.yaml -n client --ip=$(nsh config show -c resolver.yaml 'next(.ips[0])')
```

### Add Peers

The resolver and client will need to be aware of each other in order to establish
a connection. The `peer add` command will add a peer to the configuration file.

*Note: The client will need to know the resolvers public endpoint in order to
establish a connection.*

```sh
nsh peer add -c resolver.yaml \
  --name=client \
  --public-key=$(nsh config show -c client.yaml 'public(.privateKey)') \
  --ip=$(nsh config show -c client.yaml '.ips[0]')

nsh peer add -c client.yaml \
  --name=resolver \
  --public-key=$(nsh config show -c resolver.yaml 'public(.privateKey)') \
  --endpoint=$(nsh config show -c resolver.yaml '"localhost:" + (.listenPort|tostring)') \
  --ip=$(nsh config show -c resolver.yaml '.ips[0]')
```

### Start Resolver

In another terminal window, start the resolver.

```sh
nsh up -c resolver.yaml --enable-dns
```

### Connect to Resolver

#### Export WireGuard Configuration

```sh
nsh config export -c client.yaml --stripped | sudo tee /etc/wireguard/nsh0.conf > /dev/null
```

#### Setup Network Namespace

To avoid conflicts with the host network, for this example we will connect to
the resolver using a network namespace.

```sh
sudo mkdir -p /etc/netns/nsh-client-ns
echo -e "nameserver $(nsh config show -c resolver.yaml '.ips[0]')\nsearch my.nzzy.net.\n" | sudo tee /etc/netns/nsh-client-ns/resolv.conf > /dev/null
sudo ip netns add nsh-client-ns
sudo ip link add nsh0 type wireguard
sudo ip link set dev nsh0 mtu 1280
sudo ip link set nsh0 netns nsh-client-ns
sudo ip netns exec nsh-client-ns wg setconf nsh0 /etc/wireguard/nsh0.conf
sudo ip -n nsh-client-ns addr add "$(nsh config show -c client.yaml '.ips[0]')/64" dev nsh0
sudo ip -n nsh-client-ns link set nsh0 up
```

#### Test DNS Resolution

##### Peer Name

For resolving peer names, we'll need to tell `dig` to use the search domain as 
peer names are suffixed with `.my.nzzy.net.`.

The network domain can be changed by passing the `--domain` flag to the 
`config init` command.

```sh
sudo ip netns exec nsh-client-ns sudo -u $USER dig +search resolver AAAA
```

##### Internet Name

We can also use the resolver to recursively resolve internet names.

```sh
sudo ip netns exec nsh-client-ns sudo -u $USER dig google.com AAAA
```

###### IPv4

By default the resolver implements [DNS64](https://tools.ietf.org/html/rfc6147),
which will translate IPv4 only addresses to IPv6 (using the well known prefix 
`64:ff9b::/96`).

```sh
sudo ip netns exec nsh-client-ns sudo -u $USER dig ipv4.google.com AAAA
```

#### Cleanup

To remove the network namespace and WireGuard interface when you are finished.

```sh
sudo ip -n nsh-client-ns link del nsh0
sudo ip netns del nsh-client-ns
```