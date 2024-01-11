THIS IS A WORK IN PROGRESS

# Ursaserver

ursaserver is a command line rate limiter. Under the hood, it uses the [ursa].
For advanced users, it's recommended that you use the [ursa] package yourself
to setup a rate limiting server. If you don't want to create a server binary
yourself, this is what this repo provides.

The main thing that this repo does is that it reads the rate limiting
configuration from `json`, files and converts that to the `ursa.Conf` object
from [ursa] package.

## Installation

If you have go command install the binary by running

```bash
go install github.com/ursaserver/ursaserver/@latest
```

options to download the binary if you don't have go installed may be provided
in the future.

## Usage

```
Usage of ursaserver:
  -file string
    	configuration json file (default "conf.json")
  -port int
    	server port (default 3333)
```

The command line server takes a configuration file and the port to run on.
You can start the server at port, say 3111, if you have the configuration file
in the current directory as `ursaconf.json` as:

```
ursaserver -file ursaconf.json -port 3111
```

As noted, port 3333 is the default if you don't specify the `-port` option and the 
file `conf.json` is the default if you don't specify the `-file` option


## Example

I have a backend running at port 8000, and I want to set up a rate limiting
server that rate limits two of my APIs, product search API and product detail API.
Product search API is public and currently not rate limited so we'll set up a rate
limit by IP for that API and product detail API is available only for use by
frontend applications (not general users) that provide a valid API key as one 
of the http methods. 

**Goals**

1. 5 requests per minute for location API for public.
1. 100 requests per minute for product detail API for special frontends.


**Here's my JSON**
```json
{
	"Upstream": "http://localhost:8000",
	"Routes": [
		{
			"Methods": ["GET"],
			"Pattern": "^\/api\/product\/[[:alpha:]-]+\/$",
			"Rates": {"FrontendApp": "10/minute"}
		}
		{
			"Methods": ["GET"],
			"Pattern": "^\/api\/product\/$",
			"Rates": {"IP": "5/minute"}
		},
	],
	"CustomRates": {
		"FrontendApp": {
			"header": "Rebuild-Secret-Key", 
			"validIfIn": ["SOME_SECRET_HERE"], 
			"failCode": 400, 
			"failMsg": "Unauthorized" 
			}
		}
}
```

With this json saved at `conf.json` I can run `ursaserver` and the rate limiter
will start at port 3333. You can of course set up rate limit on any http method
as you like.


## Beware
1. Order of routes matters. Django has this issues too (in urlpatterns)


## Issues
1. Secrets are exposed in JSON, no comments.
1. Don't have a nice way to know what's wrong with configuration json when it doesn't work.


[ursa]: https://github.com/ursaserver/ursa
