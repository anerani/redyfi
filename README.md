# redyfi - dy.fi IP refresher

Simple app to update [dy.fi](https://www.dy.fi) DynDNS IP address for a hostname.

### Install

Clone this repository and install by running ```go get github.com/anerani/redyfi```.

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
