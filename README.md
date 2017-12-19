# redyfi - dy.fi IP refresher

Simple app to update [dy.fi](https://www.dy.fi) DynDNS IP address for a hostname.

### Install

Clone this repository and install by running ```go get github.com/anerani/redyfi```.

#### As a service

Redyfi can be deployed as a daemon on the background e.g. by using ```systemd``` or ```supervisord```. There are many ways to do this but one way can be by:

* Setting up a separate user to run the service
* Deploy the compiled binary to suitable /*/bin location
* Set up the preferred supervisor or init system to launch the service

```bash
# build and move the file to a common binary location
go build
mv redyfi /usr/local/bin/redyfi
chown redyfi:redyfi /usr/local/bin/redyfi
```

This repository includes an example to set up the updater as a service using ```systemd``` under ```/etc``` directory.

### Configuration

The application supports reading config file out of the box from:
* ```./Redyfi.json```
* ```$HOME/redyfi/Redyfi.json```
* ```/etc/redyfi/Redyfi.json```

You can define configuration for the client in following format in the config file:

```json
{
    "Username": "",
    "Password": "",
    "Hostname": "",
    "Email": ""
}
```

After the configuration file deployment, make sure that the file permissions are restricted to the user and group only.

### Usage

```
[path_to_binary]:
  -configPath string
        path to a configuration file
  -email string
        email address for user agent header
  -hostname string
        hostname to update
  -password string
        dy.fi password
  -username string
        dy.fi username
```
