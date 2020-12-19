# scanme

ScanMe allows easy continuous scan of an IP (or list of IPs).

![release](https://github.com/fopina/scanme/workflows/release/badge.svg)
[![](https://images.microbadger.com/badges/version/fopina/scanme.svg)](https://microbadger.com/images/fopina/scanme "Get your own version badge on microbadger.com")
[![](https://images.microbadger.com/badges/image/fopina/scanme.svg)](https://microbadger.com/images/fopina/scanme "Get your own image badge on microbadger.com")

Ready to use in a docker image.

### Usage

```bash
$ docker run --rm fopina/scanme -h
Usage: scanme [options] target [target ...]

Continuously scan one (or more) targets

Options:
  -path string
    	path to masscan binary (default "/usr/bin/masscan")
  -rate string
    	masscan rate (default "100")
  -show
    	show masscan output
  -sleep int
    	number of seconds to sleep between re-scans, set to 0 to disable (default 1800)
  -token string
    	PushItBot token for notifications
```

So get a token from [@PushItBot](http://fopina.github.io/tgbot-pushitbot/) (if you want the Telegram notifications) and just run it

```bash
$ docker run --rm fopina/scanme -rate 10000 -token XXAABBDD scanme.nmap.org
2019/01/13 01:36:51 Scanning [scanme.nmap.org]
2019/01/13 01:37:13 Scan finished in 22.7225257s
2019/01/13 01:37:13 45.33.32.156 - *NEW*: `22,40961` | *FINAL*: `22,40961`
...
2019/01/13 02:37:51 Scanning [scanme.nmap.org]
2019/01/13 02:38:13 Scan finished in 24.9848842s
2019/01/13 02:38:13 45.33.32.156 - *NEW*: `80,443` | *FINAL*: `22,80,443,40961`
...
```
