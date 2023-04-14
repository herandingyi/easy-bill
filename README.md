# how to run easy-bill

```shell
$ git clone https://github.com/herandingyi/easy-bill.git
$ cd easy-bill
# replace the telegram-robot-token with your own
$ sed -i s#1234567890:ABCD-1234abcdefghijklmnopqrstuvwxyz#your-telegram-robot-token#g docker-compose.yaml
$ docker-compose up -d
```

# how to upgrade easy-bill

```shell
$ docker-compose build --no-cache
$ docker-compose up -d
```
