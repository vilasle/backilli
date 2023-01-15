package yandex

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/vilamslep/backilli/pkg/fs/unit"
)

var ErrLoadingConfiguration = fmt.Errorf("failed to load cloud configuration")

type YandexClient struct {
	s3client   *s3.Client
	bucketName string
	osSep      string
	cloudSep   string
	cloudRoot  string
}

func NewClient(conf unit.ClientConfig) (*YandexClient, error) {
	customResolver := aws.EndpointResolverWithOptionsFunc(yandexResolver)

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithEndpointResolverWithOptions(customResolver))
	if err != nil {
		return nil, ErrLoadingConfiguration
	}

	s3client := s3.NewFromConfig(cfg)

	osSep := "/"
	if runtime.GOOS == "windows" {
		osSep = "\\"
	}

	return &YandexClient{
		s3client:   s3client,
		cloudRoot:  conf.RootCloud,
		cloudSep:   "/",
		osSep:      osSep,
		bucketName: conf.BucketName,
	}, nil
}

func (c YandexClient) Read(path string) ([]byte, error) {
	object := &s3.GetObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(path),
	}
	resp, err := c.s3client.GetObject(context.Background(), object)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, int(resp.ContentLength))
	defer resp.Body.Close()
	var buffer bytes.Buffer
	for true {
		num, rerr := resp.Body.Read(buf)
		if num > 0 {
			buffer.Write(buf[:num])
		} else if rerr == io.EOF || rerr != nil {
			break
		}
	}
	return buffer.Bytes(), nil
}

func (c YandexClient) Write(src string, dst string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	slpath := strings.Split(src, c.osSep)
	start := 0
	if len(slpath) > 1 {
		for i := range slpath {
			if slpath[i] == dst {
				start = (i + 1)
				break
			}
		}
	}

	yapath := fmt.Sprintf("%s%s%s", c.cloudRoot, c.cloudSep, strings.Join(slpath[start:], c.cloudRoot))

	object := &s3.PutObjectInput{
		Bucket:        aws.String(c.bucketName),
		Key:           aws.String(yapath),
		Body:          file,
		ContentLength: info.Size(),
	}

	if _, err = c.s3client.PutObject(context.Background(), object); err != nil {
		return err
	} else {
		return nil
	}
}

func (c YandexClient) Ls(path string) ([]unit.File, error) {
	var ls *s3.ListObjectsV2Output
	var err error

	params := &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucketName),
		Prefix: aws.String(path),
	}

	if ls, err = c.s3client.ListObjectsV2(context.TODO(), params); err != nil {
		return nil, err
	}

	files := make([]unit.File, 0, len(ls.Contents))
	for _, object := range ls.Contents {
		path := strings.Split(*object.Key, "/")

		name := path[len(path)-1]
		if name != ""  {
			files = append(files, unit.File{
				Date: *object.LastModified,
				Name: name,
			})
		}
	}

	return files, nil
}

func (c YandexClient) Remove(path string) error {
	deleteParams := &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(path),
	}

	if _, err := c.s3client.DeleteObject(context.TODO(), deleteParams); err != nil {
		return err
	}
	return nil
}

func (c YandexClient) Close() error {
	return nil
}

func yandexResolver(service string, region string, options ...interface{}) (aws.Endpoint, error) {
	if service == s3.ServiceID && region == "ru-central1" {
		return aws.Endpoint{
			PartitionID:   "yc",
			URL:           "https://storage.yandexcloud.net",
			SigningRegion: "ru-central1",
		}, nil
	}
	return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
}
