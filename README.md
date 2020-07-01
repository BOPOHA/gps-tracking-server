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

# test commands
```shell script
echo -n '*2a48512c373032383131343038322c435223#'       | nc auth.vorona.me 7700 # println(hex.EncodeToString([]byte("*HQ,7028114082,CR#")))
echo -n '*2a48512c373032383131343038322c56342c435223#' | nc auth.vorona.me 7700 # println(hex.EncodeToString([]byte("*HQ,7028114082,V4,CR#")))

```