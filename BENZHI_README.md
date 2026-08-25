基于 Go 实现的水产养殖循环水系统控制系统项目，一款养殖设备控制服务，完成水质监测、增氧、过滤、循环泵、温控、投喂与告警联动管理。

## 构建与运行

本项目使用 Go 1.23.12，依赖已 vendor 离线固化，构建时使用 `-mod=vendor`。

```bash
go build -mod=vendor ./...
go vet -mod=vendor ./...
go test -mod=vendor ./...
```

启动服务：

```bash
go run -mod=vendor ./cmd/aquarcirc
```

服务默认监听 8080 端口，可用环境变量 `AQUA_PORT` 覆盖。打开 `http://localhost:8080/`
进入养殖循环水控制台，控制台页面由 `web/console.html` 提供。

主要接口：

- `GET /api/health` 服务健康检查
- `GET /api/ponds` 养殖池列表与水质状态
- `POST /api/readings` 上报传感器读数
- `GET /api/alarms` 最近告警
- `POST /api/feed` 触发一轮投喂
- `GET /api/status` 设备运行状态

## 容器镜像

```bash
sh build_benzhi_docker.sh
```

镜像使用离线 vendor 构建，`GOPROXY=off`，最终镜像保留 Go 工具链，便于容器内执行
构建与测试校验。
