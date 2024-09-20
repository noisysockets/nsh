# nsh

Noisy Sockets is a layer 3 service mesh for the edge. Built top of WireGuard, 
and with the guiding question "What if endpoints could be processes instead of 
machines?".

## Features

* Wire-compatible with WireGuard.
* Userspace and unprivileged.
* Integrated service discovery.
* First class support for IPv6.
* Cross platform.

## Components

For more information on the individual components and how to get started, see the following:

* [Configuration](./docs/config.md)
* [DNS Server](./docs/dns.md)
* [Router](./docs/router.md)

## Examples

For some example use cases, see the following: 

* [Docker VPN Server](./docs/docker_vpn.md)

## Telemetry

By default the Noisy Sockets CLI gathers anonymous crash and usage statistics. 
This anonymized data is processed on our servers within the EU and is not shared 
with third parties. You can opt out of telemetry by setting the `DO_NOT_TRACK=1` 
environment variable.

## Credits

Noisy Sockets is based on code originally from the [wireguard-go](https://git.zx2c4.com/wireguard-go) project by Jason A. Donenfeld.

WireGuard is a registered trademark of Jason A. Donenfeld.