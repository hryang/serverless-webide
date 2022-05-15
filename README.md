# serverless-webide



基于 Serverless 架构和 Vscode 的**即开即用，用完即走**的轻量 Web IDE 服务。主要特点：

* 全功能 Vscode Web IDE，支持海量的插件。
* 虚拟机级别的多租安全隔离。
* 数据实时保存。用户可以随时关闭页面而不必担心数据丢失。
* 状态实时恢复。依托于函数计算极致的启动速度，秒级恢复到上次的状态。用户可随时继续。
* 资源利用率高，低成本。绝大多数 IDE 的使用是碎片化的，只在一天中的少部分时间被使用，因此 IDE 实例常驻是不明智的。借助函数计算完全按需付费，忙闲时单独定价的计费策略，成本比常驻型 IDE 实例低 3-10x。

## 快速体验

1. 开通阿里云函数计算，对象存储服务。

2. 在函数计算控制台应用中心部署 Serverless WebIDE 服务。

3. 在浏览器中访问服务的网址。Web IDE 的配置以及 /workspace 下的数据将自动保存。

## 基本流程

本项目主要实现了一个 Reverse Proxy，请求的处理流程如下图所示。

![图片.png](https://cdn.nlark.com/yuque/0/2022/png/995498/1652601830486-bfea1122-433a-49d6-b276-02ab522d8b1e.png)

## 环境配置

在 `configs` 目录下，包含了一些配置文件。请根据需要修改对应的配置文件。

* `dev.yaml`：在本地启动 webide-server 所需的配置文件
* `test.yaml`：运行测试所需的配置文件
* `fc.yaml`：在函数计算（FC） runtime 环境中运行 webide-server 所需的配置文件

在本地启动 webide-server，或者运行测试，还需要配置以下3个环境变量。

* ALI_KEY_ID：您的阿里云 access key id
* ALI_KEY_SECRET：您的阿里云 access key secret
* ALI_REGION：您要运行测试的阿里云区域，例如 cn-hangzhou，cn-beijing 等等

## 开发调试

在项目根目录下按如下步骤执行 shell 命令。

1. 修改 `dev.yaml` 中的配置项，执行下述命令编译项目。成功后，会在项目根目录下新建 target 目录，包含了二进制文件，对应的启动配置文件等交付物。

   ```shell
   make build
   ```

2. 进入 target 目录，在本地环境启动 webide server。

   ```shell
   ./ide-server
   ```

3. 请注意，step 2 只是创建了反向代理 ide-server，后台的 vscode-server 并没有启动。只有执行下述命令后，web ide 才功能就绪。其中端口请后 ide-server 启动时的端口保持一致。

   ```shell
   curl localhost:8080/initialize
   ```

4. Shutdown webide-server，将 vscode-server 的配置数据和 workspace 下的用户数据保存到 oss。

   ```shell
   curl localhost:8080/shutdown
   ```

## 本地测试

在本地运行测试，需要配置以下3个环境变量，以及 `configs` 目录中的 `test.yaml` 中的配置项。

* ALI_KEY_ID：您的阿里云 access key id
* ALI_KEY_SECRET：您的阿里云 access key secret
* ALI_REGION：您要运行测试的阿里云区域，例如 cn-hangzhou，cn-beijing 等等

在项目根目录执行命令运行测试。

```shell
make test
```



## 安装 Serverless Devs 工具

该项目使用 [Serverless Devs](https://docs.serverless-devs.com/serverless-devs/quick_start) 工具部署 FC 应用，请按照文档安装该工具。

## 部署 FC layer

Web IDE 应用依赖三方的 openvscode-server。执行下述命令，将依赖的 openvscode-server 作为 FC 的 layer 发布，这样在部署 Web IDE 应用时，就可以只更新 ide server 交付物。

```shell
make layer
```

在结果输出的最后，能找到如下内容。

![图片.png](https://cdn.nlark.com/yuque/0/2022/png/995498/1652580278643-16a68082-464d-4ad7-95bf-aa1db8cd8fd0.png)

其中需要使用红框中的内容更新项目根目录的 `s.yaml` 文件中的函数 layers 的配置。

![img](https://cdn.nlark.com/yuque/0/2022/png/995498/1652580602698-2abb72d6-bef9-4b7b-a683-4bee1b3c5085.png?x-oss-process=image%2Fresize%2Cw_1500%2Climit_0)

## 部署应用到函数计算（FC）

1. 交叉编译，生成可部署到 FC 的交付物。

   ```shell
   make release
   ```

2. 使用 Serverless Devs 工具部署到 FC。

   ```shell
   s deploy
   ```

## 函数计算（FC）应用调试技巧

### 实时日志查询

可在 FC 控制台可查看函数实时日志。也可使用 Serverless Devs 工具查询实时日志。在项目根目录（s.yaml 所在目录）执行命令：

```shell
s logs --tail
```

### 实例登录

可在 FC 控制台登录实例。也可使用 Serverless Devs 工具登录。在项目根目录（s.yaml 所在目录）执行命令：

1. 首先列出当前函数的实例。

   ```shell
   s instance list
   ```

   

2. 然后登录实例。请将 `your-instance-id` 换成您在 step 1 中列出的实例 id。

   ```shell
   s instance exec -it your-instance-id /bin/bash
   ```

   
