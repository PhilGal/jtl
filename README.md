[![Go Report Card](https://goreportcard.com/badge/github.com/PhilGal/jtl)](https://goreportcard.com/report/github.com/PhilGal/jtl)
![Build](https://github.com/PhilGal/jtl/workflows/Build/badge.svg)

# jtl - jira time logger
## Description
These are common Jtl commands used in various situations:

  * log your work as you go to a local file (see: `jtl help log`),
  * display summary report (see: `jtl help report`),
  * finally, push all data from file to your company's remote server (see: `jtl help push`)
  
For better experience, it is recommended to add a valid configuration file `$HOME/.jtl/config.yaml`. Type 'jtl help push' for more details.

When you call any command, a programm is trying to locate a data file `$HOME/data/<month-year>.csv` Thus, each month you'll have a new data file.
There is, however, a possibility to force programm to use a particular data file with `--data` global option. If you use decide to use `--data`, use it with every command, because it is a runtime option.
Same goes for the config file with `--config` option.

## Installation

### Download executable

Go to [releases](https://github.com/PhilGal/jtl/releases) and download a latest executable for your OS.
Once you get your executable, install/run it as you would usually install or run any command line program: run from where it is, add to PATH, etc.

### Build from source

Install go >= 1.14.x

```
❯ git clone https://github.com/PhilGal/jtl.git
❯ cd jtl
❯ go install
```

Running `go install` will put an executable into your `$GOPATH` directory. Make sure you have `$GOPATH/bin` in your `$PATH` and once it is there you'll be able to run `jtl` from your terminal/command prompt.

```
❯ echo $GOPATH
/usr/local/Cellar/go/1.14.1
❯ which jtl
/usr/local/Cellar/go/1.14.1/bin/jtl
```

 
