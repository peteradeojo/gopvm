# Golang PHP Version Manager

## Install

```sh
$ git clone https://github.com/peteradeojo/gopvm
$ cd gopvm
$ go build -o pvm
```

## Usage
```sh
$ ./pvm -version # List versions 
$ ./pvm -install 8.x # Install & build version 8.x
$ ./pvm -use 8.x # Use version 8.x
```

## libiconv
Getting libiconv errors in build stage?
```sh
# MacOs
$ brew install libiconv
$ export ICONV_DIR="/usr/local/opt/libiconv/"
```

# TODO:
- [ ] Manage specific version extensions
- [ ] Fix error with reading config file