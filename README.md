
# S3FS

[简体中文](README_zh.md)

S3FS is a command line tool that allows you to mount an Amazon S3 bucket as a local filesystem. It is written in Go and uses the FUSE (Filesystem in Userspace) interface to interact with the local filesystem and the AWS SDK for Go to interact with Amazon S3.

## Features

- Mount an Amazon S3 bucket as a local filesystem
- Read and write files directly from and to S3
- Run as a daemon
- Configurable via a TOML file
- Logs are written to a configurable file

## Prerequisites

- Go 1.16 or later
- An Amazon S3 bucket
- AWS Access Key and Secret Key with permissions to the S3 bucket

## Installation

```bash
go install github.com/skyline93/s3fs
```

## Configuration

Create a `config.toml` file in the same directory as the `s3fs` binary with the following content:

```toml
[s3]
endpoint = "your-endpoint"
accessKey = "your-access-key"
secretKey = "your-secret-key"
bucket = "your-bucket"

[core]
logFile = "/path/to/logfile"
pidFile = "/path/to/pidfile"
```

## Usage

```bash
s3fs
```

To run as a daemon, set `daemon = true` in the `config.toml` file.

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

[MIT](https://choosealicense.com/licenses/mit/)
