# 项目名称

汇择捕鱼代码


## 目录结构
- [common](common)
  - [conf](common%2Fconf)
    - [conf.conf](common%2Fconf%2Fconf.conf)
- [game](./game)
  - [common](game%2Fcommon)
    - [config.go](game%2Fcommon%2Fconfig.go)
  - [controllers](game%2Fcontrollers)
    - [enter_public_room.go](game%2Fcontrollers%2Fenter_public_room.go)
  - [main](game%2Fmain)
    - [config.go](game%2Fmain%2Fconfig.go)
    - [init.go](game%2Fmain%2Finit.go)
    - [main.go](game%2Fmain%2Fmain.go)
  - [router](game%2Frouter)
    - [router.go](game%2Frouter%2Frouter.go)
  - [service](game%2Fservice)
    - [client.go](game%2Fservice%2Fclient.go)
    - [define.go](game%2Fservice%2Fdefine.go)
    - [fish.go](game%2Fservice%2Ffish.go)
    - [fish_utils.go.bak](game%2Fservice%2Ffish_utils.go.bak)
    - [request.go](game%2Fservice%2Frequest.go)
    - [room.go](game%2Fservice%2Froom.go)
- [logs](logs)
- [models](./src/models)
    - [user.go](model%2Fuser.go)
- [readme.md](readme.md)
- [LICENSE](./LICENSE)(暂时没有)

## 如何运行

入口文件为game目录下的main.go

## socket.io
有空写

## redis说明

| 端口   | 作用     |
|------|--------|
| 6380 | 游戏房间配置 |
| 端口   | 作用     |
| 端口   | 作用     |


## 许可证

说明项目的许可证信息。

