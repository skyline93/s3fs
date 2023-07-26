# S3FS

S3FS是一个命令行工具，允许你将Amazon S3存储桶挂载为本地文件系统。它使用Go编写，并使用FUSE（用户空间文件系统）接口与本地文件系统交互，使用AWS SDK for Go与Amazon S3交互。

## 功能

- 将Amazon S3存储桶挂载为本地文件系统
- 直接从S3读取和写入文件
- 作为守护进程运行

## 先决条件

- Go 1.16或更高版本
- 一个Amazon S3存储桶
- 具有对S3存储桶权限的AWS访问密钥和秘密密钥

## 安装

```bash
go get github.com/skyline93/s3fs
```

## 使用

```bash
s3fs --endpoint your-endpoint --ak your-ak --sk your-sk --bucket your-bucket --mountpoint /path/to/mountpoint
```

将`your-endpoint`，`your-ak`，`your-sk`，`your-bucket`和`/path/to/mountpoint`替换为你的实际值。

作为守护进程运行：

```bash
s3fs --endpoint your-endpoint --ak your-ak --sk your-sk --bucket your-bucket --mountpoint /path/to/mountpoint --daemon
```

## 贡献

欢迎提出拉取请求。对于重大变更，请先开启一个问题，讨论你想要改变的内容。

## 许可证

[MIT](https://choosealicense.com/licenses/mit/)
