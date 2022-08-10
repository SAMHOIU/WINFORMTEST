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
* `-dir` 工作目录，默认为`./data`，抓取回来的文件会存储在`./data/fs/`
* `-c` 工作进程数，默认为`1`
* `-period` 抓取周期，以秒为单位，默认`-1`表明只抓取一次

### 主要HTTP接口
* 查看队列头数据（不从队列删除）
    ```
    GET http://localhost:2001/api/data?peek=true
    ```
* 取走队列数据（取走后会从队列删除）
    ```
    GET http://localhost:2001/api/data
    ```
* 查看爬虫状态
    ```
    GET http://localhost:2001/api/status
    ```
### 其他说明
程序运营后在启动目录下会生成两个隐藏目录`.rssl