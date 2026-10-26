# 软路由登录并拉取部署

适用 Linux x86_64 / amd64 软路由，已安装 Docker 和 Docker Compose。使用项目根目录的 `docker-compose.router.yml`，软路由端无需源码构建。

## 准备部署文件

在软路由的数据盘创建部署目录，将 `docker-compose.router.yml` 放入其中，然后进入该目录。首次部署可以创建空设备配置；已有 `config.json` 时直接使用原文件，不要覆盖：

```sh
[ -f config.json ] || printf '{"devices":[]}\n' > config.json
```

也可以将现有 `config.json` 一起复制到部署目录，以保留设备与广播地址。

管理端登录认证属于新的源码变更，当前已发布的 v1.2 不包含此功能。部署认证版本前，先构建并发布包含此变更的新镜像，再将 `docker-compose.router.yml` 的 `image` 修改为该版本。新的健康检查使用 `/healthz`，也需要新版服务支持。

在部署目录创建 `.env`，设置 `ADMIN_USERNAME=admin` 和 `ADMIN_PASSWORD=你的强密码`，或通过 DockerManager 的环境变量配置这两项。密码必填，没有默认值。此账号用于 Web 管理端登录，与下文的镜像仓库账号独立。

## 在软路由上登录仓库

使用之后执行 Docker 命令的同一系统用户登录。例如平时使用 root 管理 Docker，就在软路由 root 终端执行：

```sh
docker login --username='你的阿里云镜像仓库用户名' registry.cn-hangzhou.aliyuncs.com
```

提示输入密码时，输入阿里云容器镜像服务“访问凭证”中设置的仓库登录密码。用户名和密码以仓库控制台给出的登录信息为准。不要将密码填写在 Compose 文件里。电脑上的 Docker 登录凭据不会自动同步到软路由。

阿里云官方说明：[个人版实例推送与拉取镜像](https://www.alibabacloud.com/help/en/acr/user-guide/use-a-container-registry-personal-edition-instance-to-push-and-pull-images)。

## 拉取并启动

```sh
docker compose -f docker-compose.router.yml pull
docker compose -f docker-compose.router.yml up -d
docker compose -f docker-compose.router.yml ps
```

当前使用的镜像为 `registry.cn-hangzhou.aliyuncs.com/chldemo/home-wol:v1.2`，架构为 `linux/amd64`。

浏览器访问 `http://软路由的局域网IP:1323/admin/`，认证版本先使用配置的管理端账号密码登录。在页面顶部“网络设置”填写目标设备子网的广播地址并保存，设备信息和网络设置保存在部署目录的 `config.json` 中。

配置采用 host 网络，以便发送局域网 WOL 广播。服务使用宿主机 1323 端口，须确保该端口未被其他服务占用。

如果系统只安装了旧版独立 Compose，可将上面的 `docker compose` 替换为 `docker-compose`；若提示不支持 `pull_policy`，需升级 Compose 后再使用本配置。

## 仍然提示 denied

确认登录结果为 `Login Succeeded`，且登录与拉取使用同一系统用户。核对登录域名为 `registry.cn-hangzhou.aliyuncs.com`，仓库路径为 `chldemo/home-wol`，版本为 `v1.2`；登录账号还需拥有此仓库的拉取权限。

查看服务日志：

```sh
docker compose -f docker-compose.router.yml logs --tail 100
```

## 页面新增设备后获取验证码失败

旧版 v1.1 存在认证配置缓存未刷新的问题：页面新增设备或修改密钥后，验证码接口仍可能读取启动时的数据。可以先在软路由执行以下命令重新加载配置，然后再测试：

```sh
docker restart home-wol
```

修复镜像 v1.2 已发布，设备保存后会立即刷新认证缓存。将 Compose 的镜像地址改为 `registry.cn-hangzhou.aliyuncs.com/chldemo/home-wol:v1.2`，重新拉取并创建容器后生效：

```sh
docker compose -f docker-compose.router.yml pull
docker compose -f docker-compose.router.yml up -d
```

通过 iStoreOS DockerManager 部署时，在原项目中修改 `image` 的版本为 `v1.2`，然后选择拉取镜像并重新创建容器，保留现有配置文件挂载。

如果重启后仍报错，先通过 `http://软路由的局域网IP:1323/admin/` 直连测试，以区分应用问题与 Nginx 反向代理问题。新版页面会显示验证码接口返回的具体错误或 HTTP 状态。
