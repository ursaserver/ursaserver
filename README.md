This is a rate limiting proxy written in Go. Under the hood, it uses the
[ursa]. For advanced users, it's recommended that you use the [ursa] package
yourself to setup a rate limiting server. If you don't want to create a server
binary yourself, this is what this repo provides.

The main thing that this repo does is that it reads the rate limiting
configuration from `json`, files and converts that to the `ursa.Conf` object
from [ursa] package.

[ursa]: https://github.com/ursaserver/ursa
