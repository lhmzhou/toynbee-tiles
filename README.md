# toynbee-tiles

`toynbee-tiles` automates verifying commit hashes in production.

## Install

```
go get -u lhmzhou/toynbee-tiles
```

## Usage

```
$ toynbee-tiles --help

Usage: toynbee-tiles [options] PROJECT [PROJECT ...]

  Helps you verify commit details in production. This will either launch in the browser or extract
  the commit from the following URL patterns.

    "https://{app}-blue1.example.com",
    "https://{app}-green1.example.com",
    "https://{app}-blue-r1.example.com",
    "https://{app}-green-r1.example.com",
    "https://{app}-blue-r2.example.com",
    "https://{app}-green-r2.example.com",

Options:

  -p,--path string       The URL path to either launch or to extract version info from.

  -t,--template string   A Go template that will be printed for each endpoint based on extracting
                         the version information. This is mutually exclusive to the --open option.
                         A {{.dc}} is available with the datacenter name.

  -b,--open              Open every production endpoint in the browser.

```
