package yandex

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/vilasle/backilli/pkg/fs"
	env "github.com/vilasle/backilli/pkg/fs/environment"
	"github.com/vilasle/backilli/pkg/fs/unit"
)

var ErrLoadingConfiguration = fmt.Errorf("failed to load cloud configuration")

// var limit int64 = 536870912

type YandexClient struct {
	s3client   *s3.Client
	bucketName string
	cloudSep   string
	cloudRoot  string
}

func NewClient(conf unit.ClientConfig) (*YandexClient, error) {

	env.Set("AWS_REGION", conf.Region)
	env.Set("AWS_ACCESS_KEY_ID", conf.KeyId)
	env.Set("AWS_SECRET_ACCESS_KEY", conf.KeySecret)

	customResolver := aws.EndpointResolverWithOptionsFunc(yandexResolver)

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithEndpointResolverWithOptions(customResolver))
	if err != nil {
		return nil, ErrLoadingConfiguration
	}

	s3client := s3.NewFromConfig(cfg)

	return &YandexClient{
		s3client:   s3client,
		cloudRoot:  conf.Root,
		cloudSep:   "/",
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
	for {
		num, rerr := resp.Body.Read(buf)
		if num > 0 {
			buffer.Write(buf[:num])
		} else if rerr == io.EOF || rerr != nil {
			break
		}
	}
	return buffer.Bytes(), nil
}

func (c YandexClient) Write(buf *bytes.Buffer, dst string) (string, error) {
	return c.put(buf, dst)
}

func (c YandexClient) put(buf *bytes.Buffer, dst string) (string, error) {
	cloudRoot := c.cloudRoot
	if cloudRoot[len(cloudRoot)-1] == 0x5c ||
		cloudRoot[len(cloudRoot)-1] == 0x2f {
		cloudRoot = cloudRoot[:len(cloudRoot)-1]
	}

	s := bytes.ReplaceAll([]byte(dst), []byte{0x5c}, []byte{0x2f})

	yapath := fmt.Sprintf("%s%s%s", cloudRoot, c.cloudSep, string(s))
	bckpath := fs.GetFullPath("/", c.bucketName, yapath)
	object := &s3.PutObjectInput{
		Bucket:        aws.String(c.bucketName),
		Key:           aws.String(yapath),
		Body:          buf,
		ContentLength: int64(buf.Len()),
	}

	if _, err := c.s3client.PutObject(context.Background(), object); err != nil {
		return "", err
	} else {
		return bckpath, nil
	}
}

func (c YandexClient) Ls(path string) ([]unit.File, error) {
	var ls *s3.ListObjectsV2Output
	var err error
	lp := fs.GetFullPath("/", c.cloudRoot, path)
	params := &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucketName),
		Prefix: aws.String(lp),
	}

	if ls, err = c.s3client.ListObjectsV2(context.Background(), params); err != nil {
		return nil, err
	}

	part := make(map[string]unit.File)
	for _, object := range ls.Contents {
		srp := fs.GetFullPath("/", c.cloudRoot, path)
		key := *object.Key

		spart := strings.Split(srp, "/")
		cpart := strings.Split(key, "/")

		for i := range cpart {
			if len(spart) > i {
				if spart[i] == cpart[i] {
					continue
				}
			}
			if cpart[i] == "" {
				continue
			}

			if _, ok := part[cpart[i]]; !ok {
				part[cpart[i]] = unit.File{
					Date: *object.LastModified,
					Name: strings.Join(cpart[i:i+1], "/"),
				}
			}
			break
		}
	}
	files := make([]unit.File, 0, len(ls.Contents))
	for _, v := range part {
		files = append(files, v)
	}

	return files, nil
}

func (c YandexClient) Remove(path string) error {
	lp := fs.GetFullPath("/", c.cloudRoot, path)
	tpath := path

	fl, err := c.Ls(tpath)
	if err != nil {
		return err
	}
	if len(fl) == 0 {
		deleteParams := &s3.DeleteObjectInput{
			Bucket: aws.String(c.bucketName),
			Key:    aws.String(lp),
		}

		if _, err := c.s3client.DeleteObject(context.Background(), deleteParams); err != nil {
			return err
		}
	} else {
		for _, v := range fl {
			c.Remove(fs.GetFullPath("/", path, v.Name))
		}
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

func (c YandexClient) Description() map[string]any {
	res := make(map[string]any)
	res["name"] = "yandex.cloud"
	res["root"] = c.cloudRoot
	res["bucket"] = c.bucketName

	return res
}