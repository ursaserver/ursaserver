This is a rate limiting proxy written in Go. Under the hood, it uses the [ursa
reverse proxy]. For most users, it's recommended that you use the [ursa reverse
proxy] package yourself to setup a rate limiting server (all that's required is
you provide a configuration object). If you don't want to create a server
binary yourself, this is what this repo provides.

The main thing that this repo does is that it reads the rate limiting
configuration from `json`, `yaml` or `toml` files and converts that to the
`ursa.Conf` object from [ursa reverse proxy] package.

[ursa reverse proxy]: https://github.com/ursaserver/ursa

