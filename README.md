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

Use `go build` for now :)

