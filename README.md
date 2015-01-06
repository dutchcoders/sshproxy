sshproxy
========

Golang library to proxy ssh connections



## Why

I'm using this library in a honeypot, using this library I can intercept the ssh connections and connect each connection to their own container. Sessions can be recorder using the TypeWriterReadCloser.

## Use cases

* capture the flag
* honeypots
* creating screencasts
* whatever you'd like


## Example

```
go run examples/main.go --dest 172.16.84.182:22 --key examples/conf/id_rsa
```

Screencast of recorded session:

http://jsfiddle.net/qorz0any/1/

## Contributions
Contributions are welcome.

## Creators

**Remco Verhoef**
- <https://twitter.com/remco_verhoef>
- <https://twitter.com/dutchcoders>

## Copyright and license

Code and documentation copyright 2011-2014 Remco Verhoef.

Code released under [the MIT license](LICENSE).

