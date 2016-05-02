# Gdlsc

A simple Golang dependency license checker.

# Installation
```sh
$ go get github.com/andela-ssunday/gdlsc
```
# Usage
Enter project folder and run:
```sh
$ gdlsc
```
This assumes that you have `$GOPATH/bin` added to your `$PATH`.

**Note:** Github api is used to get license. However, Gihub has a limit for api requests made from an ip adrress in an hour. To increase the limit, you need to add an access token to your environment variable called `ACCESS_TOKEN`. For more information about getting your access token, check here: [Creating an access token](https://help.github.com/articles/creating-an-access-token-for-command-line-use/).


License
----

MIT


**Free Software!**
