# wiretap

![logo](.github/wiretap-hero.png)

A local and pipeline based tool to sniff API request and responses from clients and servers
to detect OpenAPI contract violations and compliance.

A shift left tool, for those who want to know if their applications
are actually compliant with an API.

> This is an early tool and in active development, Why not try it out and give us some feedback?

![](https://github.com/pb33f/wiretap/blob/main/.github/assets/wiretap-preview.gif)

---
# Read the quickstart guide

[🚀 Quick Start Guide 🚀](https://pb33f.io/wiretap/quickstart/)

---
# Install wiretap for your platform

## Installing using homebrew

The easiest way to install `wiretap` is to use **[homebrew](https://brew.sh)** if you're on OSX or Linux.

We have our own tap available that gives the latest and greatest version.

```shell
brew install pb33f/taps/wiretap
```

---

## Installing using npm or yarn

Building a JavaScript / TypeScript application? No problem, grab your copy of `wiretap` using your preference
of **[yarn](https://yarnpkg.com/)** or **[npm](https://npmjs.com)**.

```shell
yarn add global @pb33f/wiretap
```

or...

```shell
npm -i -g @pb33f/wiretap
```

---

## Installing using cURL

Do you want to use `wiretap` in a linux only or CI/CD pipeline or workflow? Or you don't want to/can't use
a package manager like brew?

No problem. Use **cURL** to download and run our installer script.

```shell
curl -fsSL https://pb33f.io/wiretap/install.sh | sh
```

---

## Installing/running using Docker

Love containers? Don't want to install anything? No problem, use our Docker image.

```shell
docker pull pb33f/wiretap
```

```
docker run -p 9090:9090 -p 9091:9091 -p 9092:9092 --rm -v  \
    $PWD:/work:rw pb33f/wiretap -u https://somehostoutthere.com
```

We enable the following default ports `9090`, `9091`, and `9092` for the daemon, monitor, and websockets used
by [ranch](https://github.com/pb33f/ranch) respectively.

---

## Installing on Windows

To grab your copy of `wiretap` for Windows, you can pull it from the
**[latest releases on github](https://github.com/pb33f/wiretap/releases)**
and download the Windows version for your CPU type.

---

# Running wiretap

To get up and running with the absolute defaults (which is to sniff all traffic on port 9090)
and proxy to `https://api.pb33f.com` you can run the following command.

```shell
wiretap -u https://api.pb33f.com
```

## Adding an OpenAPI contract

```shell
wiretap -u https://api.pb33f.com -s my-openapi-spec.yaml
```

# Documentation

- 🚀 [Quick Start](https://pb33f.io/wiretap/quickstart/) 🚀
- [Installing](https://pb33f.io/wiretap/quickstart/)
- [Configuring](https://pb33f.io/wiretap/configuring/)
- [Monitor UI](https://pb33f.io/wiretap/monitor/)
- [Serving static content](https://pb33f.io/wiretap/static-content/)
- [GiftShop example API](https://pb33f.io/wiretap/giftshop-api/)
- [Contributing](https://pb33f.io/wiretap/contributing/)

