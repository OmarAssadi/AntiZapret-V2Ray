# AntiZapret-V2Ray

A utility written in Go to generate a proxy ruleset matching the list of domains blocked 
in the Russian Federation for use with V2Ray or XRay to circumvent government censorship 
and evade detection.

## Zapret

***Noun*** 

запре́т • (zaprét) m inan

Definition: 
* prohibition, interdiction, ban

## How it works

The Russian government provides a list of all domains and IP addresses on the blocklist 
maintained by The Federal Service for Supervision of Communications, Information 
Technology, and Mass Media (Roskomnadzor) to telecommunications providers, like 
datacenters, cloud operators, and ISPs.

The intent behind providing the list is to allow those companies and services to
enforce the blocklist, as the government relies on the individual ISPs to do the
filtering.

However, this also makes it relatively easy to generate a ruleset to proxy
the banned services and only the banned services by making use of the same list. 

This tool makes use of a dump of the blocklist provided by Zapret-Info, which can be 
found below.

* [GitHub](https://github.com/zapret-info/z-i)
* [Assembla](https://app.assembla.com/spaces/z-i/git/source)
* [SourceForge](https://sourceforge.net/p/zapret-info/code/)
* [I2P](http://7pk2wryitzvsym36hgrsl66fp6fwfqs4ul7bb6abkinn66xi36ta.b32.i2p/zapret-info/)

## Usage

1. Obtain a CSV dump of the Roskomnadzor list from Zapret-Info, the main `dump.csv` file,
   and place it in `z-i/dump.csv` in the root directory of the project
2. Build the project using `go build`
3. Run the built executable
4. Set your proxy rules to use generated `geosite.dat` file located in `publish/` 
to route traffic. The default "geosite" country code of the list is `ZAPRETINFO`.

### Sample V2Ray configurations

Included in the [examples/](examples/) directory are a few starter V2Ray starter
configurations. In each example, only government-blocked traffic is proxied; non-blocked
traffic is routed directly over the client's default outbound connection.

* [VMess over KCP](examples/kcp-vmess)
* [VLESS over QUIC with ChaCha20 and dynamically rotating ports, masquerading as UDP 
torrent traffic.](examples/quic-chacha20-vless-dynamic-port-utp-masquerade)
* [VLESS over QUIC](examples/quic-vless)
* [VMess over WebSockets with TLS, using Caddy as a reverse proxy](examples/websocket-tls-vmess)

*NOTE:* *please replace any  security keys and UUIDs with your own locally generated ones!*

#### Which Configuration To Pick

Depending on your needs, different setups may be more appropriate. In terms of security,
a traditional VPN setup with more mature, peer-reviewed protocols is likely the safest
option. 

However, while the traffic going through the tunnel is secure when using WireGuard or 
OpenVPN, it is also much easier to detect the fact that a VPN is in use. If evasion of
DPI and anti-circumvention technology is required, or a more covert solution is necessary
for some other reason, one of the VMess/VLESS options is the best choice currently.

##### VMess vs. VLESS

VLESS is a newer stateless transport protocol that does not require system time to be 
in sync nor the usage of alter ids. However, it does not implement its own encryption. 
And thus, it would be inappropriate to VLESS without some sort of layer, like TLS, 
sitting on top (e.g., VLESS + KCP *without* TLS would be ill-advised).

VMess, on the other hand, does depend on the server and client system times to be no more
than 90 seconds out of sync. For clarity, this is timezone independent, i.e., the client
and server may both have distinct non-UTC timezones, but the time difference must be no 
more than 90 seconds from each other once UTC offsets have been taken into account.

A pro and a con of VMess, however, is that it implements its own encryption layer, which
can be either AES-128-GCM or ChaCha20-Poly1305. This adds additional overhead when used 
with TLS due to the double encryption, but it also allows VMess to be used without TLS.

##### TCP vs. KCP vs. QUIC vs. WebSockets

[KCP](https://github.com/skywind3000/kcp/blob/master/README.en.md) is a reliable 
transport protocol built on-top of UDP with the intention of minimizing latency 
at the cost of lower throughput. It is well-supported and more mature than the QUIC
options provided in V2Ray. It is suitable for general use and can be paired with VMess
for encryption and masquerading. TCP may be a better choice if throughput is a bigger
concern than latency.

The TCP transport layer is the oldest and most mature option in V2Ray, and can be paired
with VMess or VLESS and TLS.

V2Ray also has a WebSocket transport layer, which can be used with VMess or VLESS and
TLS. However, it will have more overhead than TCP or KCP and thus is generally not
advisable unless there is the need to use a standard webserver like NGINX or Caddy
as a reverse proxy. Another use case for the WebSocket transport is for use behind 
WebSocket-enabled CDNs, like CloudFlare.

QUIC is the least mature transport option out of the ones list thus far, but should 
result in better throughput and latency than TCP with TLS.

When in doubt, use TCP and VLESS with TLS or KCP and VMess, unless CloudFlare is in use,
in which case WebSockets or gRPC and VLESS would be more appropriate.

## Acknowledgements

Special thanks to the following groups and individuals for their efforts.

* [Zapret-Info](https://github.com/zapret-info/z-i) - Providing regular dumps of the
Russian government blocklist
* [Shadowsocks](https://github.com/shadowsocks/) - Proxy encryption protocol designed to
circumvent China's great firewall
* [Clowwindy](https://github.com/Clowwindy) - Original development of Shadowsocks
* [V2Ray](https://github.com/v2ray) - A platform for building proxies, and the VMess
protocol, designed to bypass anti-circumvention techniques
* [XRay](https://github.com/XTLS) - An alternative v2ray-core with support for XTLS
* [Loyalsoldier](https://github.com/loyalsoldier) - References for generating V2Ray rules