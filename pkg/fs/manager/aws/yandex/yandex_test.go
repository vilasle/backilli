package yandex

import (
	"fmt"
	"os"
	"testing"

	env "github.com/vilamslep/backilli/pkg/fs/environment"
	"github.com/vilamslep/backilli/pkg/fs/unit"
)

func TestWrite(t *testing.T) {
	if err := env.LoadEnvfile("test.env"); err != nil {
		t.Fatal(err)
	}

	conf := unit.ClientConfig{
		Root:  env.Get("ROOT_PLACE"),
		BucketName: env.Get("BUCKET_NAME"),
		Region: env.Get("AWS_REGION"),
		KeyId: env.Get("AWS_ACCESS_KEY_ID"),
		KeySecret: env.Get("AWS_SECRET_ACCESS_KEY"),
	}

	path := env.Get("PATH_SRC")
	wd, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	testSet := getTestSet()

	_, err = wd.Write(testSet)
	if err != nil {
		t.Fatal(err)
	}
	wd.Close()
	defer os.Remove(path)

	client, err := NewClient(conf)
	if err != nil {
		t.Fatal(err)
	}

	if err := client.Write(path, "test.txt"); err != nil {
		t.Fatal(err)
	}
}

func TestRead(t *testing.T) {
	if err := env.LoadEnvfile("test.env"); err != nil {
		t.Fatal(err)
	}
	path := fmt.Sprintf("%s%s", env.Get("ROOT_PLACE"),env.Get("PATH_DST"))
	conf := unit.ClientConfig{
		Root:  env.Get("ROOT_PLACE"),
		BucketName: env.Get("BUCKET_NAME"),
		Region: env.Get("AWS_REGION"),
		KeyId: env.Get("AWS_ACCESS_KEY_ID"),
		KeySecret: env.Get("AWS_SECRET_ACCESS_KEY"),
	}

	client, err := NewClient(conf)
	if err != nil {
		t.Fatal(err)
	}
	testSet := getTestSet()
	c, err := client.Read(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(testSet) != len(c) {
		t.Fatal("lenght data from file does not with test sets")
	}

	for i, s := range testSet {
		c := c[i]
		if c != s {
			t.Fatalf("byte with index %d does not match with test sets. Expected %v. There is %v", i, s, c)
		}
	}
}

func TestLs(t *testing.T) {
	if err := env.LoadEnvfile("test.env"); err != nil {
		t.Fatal(err)
	}

	conf := unit.ClientConfig{
		Root:  env.Get("ROOT_PLACE"),
		BucketName: env.Get("BUCKET_NAME"),
		Region: env.Get("AWS_REGION"),
		KeyId: env.Get("AWS_ACCESS_KEY_ID"),
		KeySecret: env.Get("AWS_SECRET_ACCESS_KEY"),
	}

	client, err := NewClient(conf)
	if err != nil {
		t.Fatal(err)
	}

	ls, err := client.Ls(env.Get("ROOT_PLACE"))
	if err != nil {
		t.Fatal(err)
	}

	pwdLs := []unit.File{
		{Name: "test.txt"},
	}

	for _, v := range ls {
		name := v.Name
		found := false
		for _, vj := range pwdLs {
			if vj.Name == name {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("does not found file %s", v.Name)
		}
	}
}

func TestRemove(t *testing.T) {
	if err := env.LoadEnvfile("test.env"); err != nil {
		t.Fatal(err)
	}

	conf := unit.ClientConfig{
		Root:  env.Get("ROOT_PLACE"),
		BucketName: env.Get("BUCKET_NAME"),
		Region: env.Get("AWS_REGION"),
		KeyId: env.Get("AWS_ACCESS_KEY_ID"),
		KeySecret: env.Get("AWS_SECRET_ACCESS_KEY"),
	}

	client, err := NewClient(conf)
	if err != nil {
		t.Fatal(err)
	}

	if err := client.Remove(env.Get("PATH_CLOUD_SRC")); err != nil {
		t.Fatal(err)
	}
}

func getTestSet() []byte {
	n := 2048
	p := "AaBbCcDdEeFfGgHhIiJjKkLlMmNnOoPpQqRrSsTtUuVvWwXxYyZz"
	tp := make([]byte, len(p))
	for i := 0; i < len(p); i++ {
		tp[i] = p[i]
	}

	testSet := make([]byte, 0)
	for i := 0; i < n; i++ {
		testSet = append(testSet, tp...)
	}

	return testSet
}
