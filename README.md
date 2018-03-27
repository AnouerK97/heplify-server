# ![](https://i.imgur.com/QvLYJkC.png)

**heplify-server** is a stand-alone **HOMER** *v5 Capture Server* developed in GO, optimized for speed and simplicity. Distributed as a single binary ready to capture TCP/UDP **HEP** encapsulated packets from [heplify](https://github.com/sipcapture/heplify) or any other [HEP](https://github.com/sipcapture/hep) enabled agent or platform, indexing to database using H5 table format, producing basic usage metrics timeseries and providing users with simple, basic options for correlation and tagging inline.

*TLDR; instant, stand-alone, minimal HOMER without Kamailio or OpenSIPS dependency/options.*

### Notice
**heplify-server** only offers a reduced set of options and is *not* designed for everyone, but should result ideal for those willing to have an *all-in-one* simple capture deployment with minimal complexity and no need for special customization.

### Status 
#### v1 *(current)*
  * Beta Stage - **NOT READY FOR PRODUCTION**
  * HOMER 5 Schema
    * SIP, correlated RTCP, RTCPXR, Logs
  * Testers, Reporters and Contributors [welcome](https://github.com/sipcapture/heplify-server/issues)
#### v2
  * Homer v7 schema
  * *Coming Soon!*

### Installation
* Download a [release](https://github.com/negbie/heplify-server/releases)
* Compile from [sources](https://github.com/negbie/heplify-server/blob/master/docker/heplify-server/Dockerfile)

### Configuration
heplify-server can be configured using command-line options, or by defining a local [configuration file](https://github.com/lmangani/heplify-server/blob/master/example/heplify-server.toml)

------

### Testing
##### Stand-Alone
```
heplify-server -h
```
##### Docker
A sample Docker [compose](https://github.com/sipcapture/heplify-server/tree/master/docker/homer-heplify) file is available providing Heplify-Server, Homer 5 UI, and basic MySQL in seconds!
```
cd heplify-server/docker/homer-heplify
docker-compose up -d
```
##### Service
A sample service file is available under `/example`
```
cp example/heplify-server.service /etc/systemd/system/
systemctl daemon-reload
systemctl start heplify-server
systemctl enable heplify-server
```

----
#### Made by Humans
This Open-Source project is made possible by actual Humans without corporate sponsors, angels or patreons.<br>
If you use this software in production, please consider supporting its development with contributions or [donations](https://www.paypal.com/cgi-bin/webscr?cmd=_donations&business=donation%40sipcapture%2eorg&lc=US&item_name=SIPCAPTURE&no_note=0&currency_code=EUR&bn=PP%2dDonationsBF%3abtn_donateCC_LG%2egif%3aNonHostedGuest)

[![Donate](https://www.paypalobjects.com/en_US/i/btn/btn_donateCC_LG.gif)](https://www.paypal.com/cgi-bin/webscr?cmd=_donations&business=donation%40sipcapture%2eorg&lc=US&item_name=SIPCAPTURE&no_note=0&currency_code=EUR&bn=PP%2dDonationsBF%3abtn_donateCC_LG%2egif%3aNonHostedGuest) 
