# gps-tracking-server

This demo application was written in March 2016 to accept and store binary data from Ruptela (http://www.ruptela.com) and Teltonika (http://teltonika.lt) GPS tracking devices.

P.S. I'm noob in Go programming :)

![preview](https://github.com/nenadvasic/gps-tracking-server/blob/master/preview.png?raw=true)

# build

```shell script
go mod download
go build -o ~/bin/gps-gatewayd cmd/gatewayd/*go
go build -o ~/bin/gps-frontend cmd/frontend/*go
```
