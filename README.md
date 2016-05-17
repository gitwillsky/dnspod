## dnspod
A simply dnspod.cn DDNS implementation, using go language。

### Usage

Command way:

    $ dnspod your_email your_password

Config way:

    $ dnspod config

Before using config way,you must create a config file named `ddns.conf` like this:

```
email = xxx@xx.com       # your dnspod login email
password = xxxx          # your dnspod login password
domains = 1,2,3          # your domain sequence number in dnspod.cn domain list
records = 5,6,7          # your record sequence number in dnspod.cn domain's record list
```

### Found bug

Use Pull－Request.

### Need help

Email ： hdu_willsky@foxmail.com
