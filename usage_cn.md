# Crawler使用方法

## 使用说明

### 安装golang开发环境（如果不安装也可以直接下载release包）。

### 通过go来安装`crawler.club/crawler`。

```sh
go install crawler.club/crawler
crawler --help
```

### 主要参数说明
* `-api` 打开通过http取数据的接口
* `-addr` http服务地址，查看爬虫状态、取数据等http接口，默认为`2001`
* `-fs` 打开本地文件存储，默认开启
* `-dir` 工作目录，默认为