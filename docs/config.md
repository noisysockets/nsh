# Config

Noisy Sockets uses a custom, versioned YAML configuration file to store details 
about the underlying WireGuard network.

The choice to use a custom config format over just extending the WireGuard INI
format was multi-faceted. But it mostly came down to wanting something with more
structure, and support for versioning.

In order to retain compatibility with existing WireGuard tools, the YAML 
configuration can be lossly converted to and from the WireGuard INI format using 
the `config import` and `config export` commands.

## Configuration File

The configuration file is by default stored in the `$XDG_CONFIG_HOME` directory.

Depending on your operating system, this will be something like:

- Linux: `~/.config/nsh/noisysockets.yaml`
- MacOS: `~/Library/Application Support/nsh/noisysockets.yaml`
- Windows: `%LOCALAPPDATA%\nsh\noisysockets.yaml`

The default configuration path can be overridden using the `--config` flag.

## Config Show

In order to make it easier to work with the configuration file, Noisy Sockets
provides a `config show` command that supports [jq](https://stedolan.github.io/jq/)
style queries.

```bash
nsh config show '.peers[0].publicKey'
```

We've added a few extra built-in functions to make it easier to work with 
WireGuard configuration.

### Built-in Functions

#### public

The `public()` function can be used to get the base64 encoded Curve25519 
public key from a base64 encoded Curve25519 private key.

Eg. to show the current peers public key:

```bash
nsh config show 'public(.privateKey)'
```

#### next

The `next()` function can be used to get the next available IP address from a
CIDR block.

Eg. to show the next available IP address:

```bash
nsh config show 'next(.ips[0])'
```
