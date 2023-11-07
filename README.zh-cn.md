## Hyperdot 是什么
Hyperdot 是一个区块链分析平台，可以访问和消费链上的数据，更像区块链的 Github，但 hyperdot 核心在数据分析和图表上。在此工具的帮助下，Hyperdot 的用户可以查看和交换来自以下来源的数据：Polkadot, Kusama, Moonbeam, Acala, HydraDX, Astar, Hashed 等数百个平行链。

## Hyperdot-node 是什么

hyperdot-node 是 hyperdot 项目的一部分。hyperdot-node 集成了 [substrate-etl](https://github.com/colorfulnotion/substrate-etl) 提供的 bigquery 的 polkadot 的公共数仓，hyperdot-node 为 hyperdot 提供了所需的 API，编排了基础设施。

hyperdot-node 可以：

- 为 hyperdot-fronted 提供所需的 API。
- 提供数据库、缓存、对象存储等基础设施层的编排和管理。
- 集成了 substrate-etl 的 Google query 的 polkadot 公共数仓。
- 提供通过 SQL 进行查询并创作图表、仪表盘功能。
- 提供图标、仪表盘的分享。

hyperdot-node 组成图下图所示

![hyperdot-node architcture](docs/imgs/hyperdot-node-arch.png)

## 安装

有两个选项用户安装和运行 hyperdot-node server:
1. [Docker 安装指南](#Docker-安装指南)
2. [源码模式安装指南](#源码模式安装指南)


## 安装前准备
为了尽快运行 hyperdot-node，需要做以下准备
1. 确保您安装了 docker 环境。
2. 确保您安装了 docker-compose 编排工具。
3. 您需要准备 Google Bigquery 账号。
4. 您需要准备 [polkaholic](https://polkaholic.io/) 的 [api key](https://polkaholic.io/login)。


### 准备 Google 应用默认凭据

Hyperdot-node 需要从 Google Bigquery 上查询数据，因此在本地运行 hyperdot 时，您需要设置 Google 应用默认凭据，您可以参考 [设置应用默认凭据](https://cloud.google.com/docs/authentication/provide-credentials-adc?hl=zh-cn#on-prem) 以获取更多信息。

当准备好 Google 应用凭据之后，你可以按照以下步骤尽快准备

1. 创建一个新的项目，当然您也可以使用已有的项目。
2. 为账号分配权限，确保可以访问 Bigquery。
3. 进入到 Google 控制台的[服务账号](https://console.cloud.google.com/iam-admin/serviceaccounts?hl=zh-cn) 页面，创建或使用应用凭据。
4. 获得应用凭据，将其放置在项目根目录 `config` 文件夹中，并将其重新命名为 `hyperdot-gcloud-iam.json`。

## Docker 安装指南

使用 Docker 运行应用程序可以实现最少的安装和快速设置。建议仅将其用于评估用例，例如本地开发构建。

1. 克隆该项目到你的本地 `git clone https://github.com/Infra3-Network/hyperdot-node.git`。
2. 复制 `config/hyperdot-sample.json` 为 `config/hyperdot.json`。
3. 修改您在准备工作第4步获得的应用凭据文件名称为 `hyperdot-gcloud-iam.json`，并将其复制到 `config` 目录中。
4. 修改 `config/hyperdot.json` 的配置，例如您在准备工作第1步中使用的项目为 `foo`，你需要修改配置中
   ```json
    {
       "bigquery": {
        "projectId": "foo"
        }
    }
   ```
5. 编译 docker 镜像
   ```shell
    make build/docker
   ```
   如果您时 Macos m1/2/3 chip, 可以编译 arm的镜像
   ```shell
   make build/docker-arm
   ```
6. 通过下面的命令启动或停止 hyperdot-node
   ```shell
   make up-docker
   make stop-docker
   ```
   如果您本地没有 hyperdot-node 镜像，该命令首先会编译 hyperdot-node 镜像，并启动 `postgres`, `redis`, `minio` 基础层服务。
7. 现在您应该启动了 hyperdot-node 服务，试着访问 http://localhost:3030/apis/v1/swager/index.html 看看吧！



## 源码模式安装指南

源码模式安装需要
- Docker Engine
- Makefile 工具链
- Unix-based 操作系统
- Go 1.20
- Postgres
- Redis
- Minio


根据以下步骤，您可以快速运行 hyperdot-node server
1. 编译 go 源码
   ```shell
   go mod tidy && go mod vendor

   go build -o /path/to/hyperdot-node cmd/node/main.go 
   ```
   如果您想关闭 CGO，可以
   ```shell
    CGO_ENABLED=0 GOOS=linux go mod tidy 
    CGO_ENABLED=0 GOOS=linux go mod vendor 
    CGO_ENABLED=0 GOOS=linux go build -o /path/to/hyperdot-node cmd/node/main.go
   ```
2. 要运行 hyperdot-node，您需要更改您的配置来集成基础层系统，您可以参考[配置](#配置) 部分查看如何更改。
3. 运行程序，你可以通过执行编译好的 golang 程序
    ```shell
    /path/to/hyperdot-node -config=/path/to/hyperdot.json
    ```
4. 试着访问 http://localhost:3030/apis/v1/swager/index.html 看看吧！

## 配置
我们提供了一个示例配置 `config/hyperdot-sample.json`，你可以参考该配置进行修改

```json
{
    "polkaholic": {
        "apiKey": "<YOU API_KEY>",
        "baseUrl": "https://api.polkaholic.io"
    },
    "apiserver": {
        "addr": ":3030"
    },
    "bigquery": {
        "projectId": "hyperdot"
    },
    "localStore": {
        "bolt": {
            "path": "hyperdot.db"
        }
    },
    "postgres": {
        "host": "postgres",
        "port": 5432,
        "user": "hyperdot",
        "password": "hyperdot",
        "db": "hyperdot",
        "tz": "Asia/ShangHai"
    },
    "s3": {
        "endpoint": "minio:9000",
        "useSSL": false,
        "accessKey": "hyperdot",
        "secretKey": "hyperdot"
    },
    "redis": {
        "addr": "redis:6379"
    }

}

```

- `polkaholic`
  - `apiKey`: 我们从[polkaholic.io](https://polkaholic.io/)获取链配置信息，因此，您需要修改为您自己的 key。
  - `baseUrl`: 通常您不需要修改该配置。
- `apiserver`
   - `addr`: hyperdot 默认运行在 3030 端口，如果您需要可以修改该配置。
- `bigquery`
  - `projectId`: 您需要配置 `projectId` 来查询 bigquery，你可以参考[安装前准备](#安装前准备)
- `localStore`
  - `blot`
    - `path`: 我们使用 `bblot` 数据库存储一些数据和元数据，如果需要，您可以修改存储路径
- `postgres`: 我们使用 `postgres` 存储用户数据，如果需要，您可以修改 postgres 的配置。
- `s3`: 我们默认使用 `minio` 对象存储来存储用户 blob 数据，如果需要，您也可以使用兼容 `s3` 协议的其他对象存储。
- `redis`：我们使用 redis 来缓存链上和用户的数据，如果需要，您也可以修改 redis 的配置。


## 测试
这将引导您完成测试发行者节点各个方面的步骤。


### 启动测试环境

```shell
make up-test
# docker-compose -f tests/docker-compose.yaml up -d
# Creating network "hyperdot-node-test_default" with the default driver
# Creating hyperdot-test-postgres ... done
# Creating hyperdot-test-redis    ... done
# Creating hyperdot-test-minio    ... done
```

### 运行测试

在运行测试前，你同样需要配置 google 应用凭据和 polkaholic 的 api key，如果您已经完成了配置可以参考下面的命令运行测试，否则您可以参考[安装前准备](#安装前准备) 来查看如何配置。

创建好凭据之后，你需要修改配置文件

```shell
cp tests/hyperdot-sample.test.json  tests/hyperdot.test.json
```

修改 `tests/hyperdot.test.json` 配置中的 `polkaholic` 和 `bigquery` 配置项中的 `apiKey` 和 `projectId` 

```json
{
    "polkaholic": {
        "apiKey": "<YOU_API_KEY>",
        "baseUrl": "https://api.polkaholic.io"
    },

    "bigquery": {
        "projectId": "<YOU_PROJECT_ID>"
    },
}
```

然后，通过下面的命令运行测试。

```shell
make tests
```

### 运行 Lint

如果您想运行 lint 来检查你的代码，您需要
- 安装 golangci-lint，您可以参考 [Install](https://golangci-lint.run/usage/install/)


```shell
make lint
```
## API 参考
我们为内部所有的 API 都生成了 Openapi 供您参考，你可以将 `docs/swager.[json|yaml]` 导入到兼容 swager 的在线浏览服务中。也可以启动 hyperdot-node 后，访问 `/apis/v1/swager/index.html` 查看

![hyperdot-node api文档](docs/imgs/hyperdot-node-api.png)


## 故障排查

1. 对于中国用户，如果您在编译程序时遇到 timeout 问题，可以尝试配置 golang 代理 https://goproxy.io/zh/。
2. 对于中国用户，如果您在运行 hyperdot-node 的镜像或者已经编译好的程序时，出现访问 Google Bigquery TLS timeout 问题时，你可以配置代理
   ```shell
   # 对于 docker 镜像，将以下内容写入到 ~/.docker/config.json，然后重启运行docker
   {
    "proxies":
        {
        "default":
            {
                "httpProxy": "http://proxy.example.com:8080",
                "httpsProxy": "http://proxy.example.com:8080",
                "noProxy": "localhost,127.0.0.1,.example.com"
            }
        }
    }

   # 对于编译好的程序，可以在终端设置dialing
   export http_proxy=<you proxy>
   export https_proxy=<you proxy>
   ```
   

## 许可证

请查看 [LICENSE](../LICENSE)