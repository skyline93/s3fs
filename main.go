package main

import (
	"bytes"
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sevlyar/go-daemon"
	"github.com/spf13/cobra"
)

type S3FS struct {
	s3     *s3.S3
	bucket string
}

func (fs *S3FS) Root() (fs.Node, error) {
	log.Println("Accessing root of the filesystem")
	return &S3Dir{fs: fs, path: ""}, nil
}

type S3Dir struct {
	fs   *S3FS
	path string
}

func (d *S3Dir) Attr(ctx context.Context, attr *fuse.Attr) error {
	log.Println("Getting attributes of directory:", d.path)
	attr.Mode = os.ModeDir | 0664
	return nil
}

func (d *S3Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	path := strings.TrimPrefix(d.path+"/"+name, "/")
	log.Println("Looking up path:", path)
	resp, err := d.fs.s3.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:    aws.String(d.fs.bucket),
		Prefix:    aws.String(path),
		Delimiter: aws.String("/"),
		MaxKeys:   aws.Int64(1),
	})
	if err != nil || (*resp.KeyCount == 0 && len(resp.CommonPrefixes) == 0) {
		log.Println("Path not found:", path)
		return nil, os.ErrNotExist
	}
	if len(resp.CommonPrefixes) > 0 {
		log.Println("Path is a directory:", path)
		return &S3Dir{fs: d.fs, path: path}, nil
	}
	log.Println("Path is a file:", path)
	return &S3File{fs: d.fs, s3: d.fs.s3, bucket: d.fs.bucket, path: path, cache: []byte{}}, nil
}

type S3File struct {
	fs     *S3FS
	s3     *s3.S3
	path   string
	bucket string
	cache  []byte
}

func (f *S3File) Attr(ctx context.Context, attr *fuse.Attr) error {
	log.Println("Getting attributes of file:", f.path)
	resp, err := f.fs.s3.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(f.fs.bucket),
		Key:    aws.String(f.path),
	})
	if err != nil {
		log.Println("Failed to get attributes of file:", f.path)
		return err
	}
	attr.Mode = 0666
	attr.Size = uint64(*resp.ContentLength)
	attr.Mtime = *resp.LastModified
	attr.Atime = *resp.LastModified
	return nil
}

func (d *S3Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Println("d.path is:", d.path)
	prefix := strings.TrimPrefix(d.path+"/", "/")
	log.Println("Trimmed prefix is:", prefix)
	log.Println("Reading directory:", prefix)
	resp, err := d.fs.s3.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:    aws.String(d.fs.bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	})
	if err != nil {
		log.Printf("Failed to read directory: %s, err: %s", prefix, err)
		return nil, err
	}
	var dirents []fuse.Dirent
	for _, cp := range resp.CommonPrefixes {
		name := strings.TrimPrefix(*cp.Prefix, prefix)
		name = strings.TrimSuffix(name, "/")
		dirents = append(dirents, fuse.Dirent{Name: name, Type: fuse.DT_Dir})
	}
	for _, obj := range resp.Contents {
		name := strings.TrimPrefix(*obj.Key, prefix)
		dirents = append(dirents, fuse.Dirent{Name: name, Type: fuse.DT_File})
	}
	return dirents, nil
}

func (f *S3File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	log.Println("Starting to write data")

	// 将数据写入到本地缓存
	log.Println("Writing data to cache")
	f.cache = append(f.cache, req.Data...)

	// 如果缓存大小超过5MB，将缓存的数据写入到S3
	if len(f.cache) >= 5*1024*1024 {
		log.Println("Cache size exceeded 5MB, flushing cache")
		err := f.flushCache()
		if err != nil {
			log.Println("Error flushing cache: ", err)
			return err
		}
	}

	// 更新写入的字节数
	log.Println("Updating written bytes size")
	resp.Size = len(req.Data)

	log.Println("Data written successfully")
	return nil
}

func (f *S3File) flushCache() error {
	log.Println("Starting to flush cache")

	// 初始化一个multipart上传
	log.Println("Initializing multipart upload")
	createMultipartUploadInput := &s3.CreateMultipartUploadInput{
		Bucket: aws.String(f.bucket),
		Key:    aws.String(f.path),
	}
	createMultipartUploadOutput, err := f.s3.CreateMultipartUpload(createMultipartUploadInput)
	if err != nil {
		log.Println("Error initializing multipart upload: ", err)
		return err
	}

	// 将缓存的数据上传为一个部分
	log.Println("Uploading part")
	uploadPartInput := &s3.UploadPartInput{
		Bucket:        aws.String(f.bucket),
		Key:           aws.String(f.path),
		PartNumber:    aws.Int64(1),
		UploadId:      createMultipartUploadOutput.UploadId,
		Body:          bytes.NewReader(f.cache),
		ContentLength: aws.Int64(int64(len(f.cache))),
	}
	uploadPartOutput, err := f.s3.UploadPart(uploadPartInput)
	if err != nil {
		log.Println("Error uploading part: ", err)
		return err
	}

	// 完成multipart上传
	log.Println("Completing multipart upload")
	completeMultipartUploadInput := &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(f.bucket),
		Key:      aws.String(f.path),
		UploadId: createMultipartUploadOutput.UploadId,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: []*s3.CompletedPart{
				{
					ETag:       uploadPartOutput.ETag,
					PartNumber: aws.Int64(1),
				},
			},
		},
	}
	_, err = f.s3.CompleteMultipartUpload(completeMultipartUploadInput)
	if err != nil {
		log.Println("Error completing multipart upload: ", err)
		return err
	}

	// 清空缓存
	log.Println("Clearing cache")
	f.cache = nil

	log.Println("Cache flushed successfully")
	return nil
}

func (d *S3Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	log.Println("Creating new file:", req.Name)

	// 创建新的文件在S3
	_, err := d.fs.s3.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(d.fs.bucket),
		Key:    aws.String(d.path + "/" + req.Name),
		Body:   bytes.NewReader([]byte{}),
	})
	if err != nil {
		log.Println("Failed to create new file:", req.Name)
		return nil, nil, err
	}

	// 返回一个代表新文件的fs.Node
	f := &S3File{fs: d.fs, path: d.path + "/" + req.Name}
	return f, f, nil
}

func (f *S3File) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	log.Println("Flushing file:", f.path)

	// 将缓存的数据写入到S3
	if len(f.cache) > 0 {
		err := f.flushCache()
		if err != nil {
			log.Println("Error flushing cache: ", err)
			return err
		}
	}

	return nil
}

func main() {
	var endpoint, ak, sk, bucket, mountpoint string
	var daemonize bool

	var rootCmd = &cobra.Command{
		Use:   "s3fs",
		Short: "Mount an S3 bucket as a local filesystem",
		Long: `s3fs is a command line tool that allows you to mount an S3 bucket as a local filesystem.
	
You can use it like this:

s3fs --endpoint your-endpoint --ak your-ak --sk your-sk --bucket your-bucket --mountpoint /path/to/mountpoint

Please replace 'your-endpoint', 'your-ak', 'your-sk', 'your-bucket', and '/path/to/mountpoint' with your actual values.`,
		Run: func(cmd *cobra.Command, args []string) {
			if daemonize {
				cntxt := &daemon.Context{
					PidFileName: "s3fs.pid",
					PidFilePerm: 0644,
					LogFileName: "s3fs.log",
					LogFilePerm: 0640,
					WorkDir:     "./",
					Umask:       027,
					Args:        os.Args,
				}

				d, err := cntxt.Reborn()
				if err != nil {
					log.Fatal("Unable to run: ", err)
				}
				if d != nil {
					return
				}
				defer cntxt.Release()

				log.Print("- - - - - - - - - - - - - - -")
				log.Print("daemon started")
			}

			// 创建AWS会话
			log.Println("Creating AWS session")
			sess := session.Must(session.NewSession(&aws.Config{
				Region:           aws.String("us-east-1"),
				Endpoint:         aws.String(endpoint),
				S3ForcePathStyle: aws.Bool(true),
				Credentials:      credentials.NewStaticCredentials(ak, sk, ""),
			}))

			// 挂载文件系统
			log.Println("Mounting filesystem at:", mountpoint)
			c, err := fuse.Mount(mountpoint, fuse.AllowOther())
			if err != nil {
				log.Fatalf("Failed to mount filesystem: %v", err)
			}
			defer c.Close()

			// 启动文件系统
			log.Println("Starting filesystem")
			err = fs.Serve(c, &S3FS{s3: s3.New(sess), bucket: bucket})
			if err != nil {
				log.Fatalf("Failed to start filesystem: %v", err)
			}

			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs)
			// loop until we receive a SIGINT or SIGTERM
			for {
				select {
				case sig := <-sigs:
					if sig == syscall.SIGINT || sig == syscall.SIGTERM {
						log.Print("exited")
						return
					}
				}
			}
		},
	}

	rootCmd.Flags().StringVarP(&endpoint, "endpoint", "e", "", "S3 endpoint")
	rootCmd.Flags().StringVarP(&ak, "ak", "a", "", "Access key")
	rootCmd.Flags().StringVarP(&sk, "sk", "s", "", "Secret key")
	rootCmd.Flags().StringVarP(&bucket, "bucket", "b", "", "Bucket name")
	rootCmd.Flags().StringVarP(&mountpoint, "mountpoint", "m", "", "Mount point")
	rootCmd.Flags().BoolVarP(&daemonize, "daemon", "d", false, "Run as a daemon")

	rootCmd.Execute()
}
