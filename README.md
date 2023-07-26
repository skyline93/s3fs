# S3FS

[简体中文](README_zh.md)

S3FS is a command line tool that allows you to mount an Amazon S3 bucket as a local filesystem. It is written in Go and uses the FUSE (Filesystem in Userspace) interface to interact with the local filesystem and the AWS SDK for Go to interact with Amazon S3.

## Features

- Mount an Amazon S3 bucket as a local filesystem
- Read and write files directly from and to S3
- Run as a daemon

## Prerequisites

- Go 1.16 or later
- An Amazon S3 bucket
- AWS Access Key and Secret Key with permissions to the S3 bucket

## Installation

```bash
go get github.com/yourusername/s3fs
```

Replace `yourusername` with your actual GitHub username.

## Usage

```bash
s3fs --endpoint your-endpoint --ak your-ak --sk your-sk --bucket your-bucket --mountpoint /path/to/mountpoint
```

Replace `your-endpoint`, `your-ak`, `your-sk`, `your-bucket`, and `/path/to/mountpoint` with your actual values.

To run as a daemon:

```bash
s3fs --endpoint your-endpoint --ak your-ak --sk your-sk --bucket your-bucket --mountpoint /path/to/mountpoint --daemon
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

[MIT](https://choosealicense.com/licenses/mit/)
