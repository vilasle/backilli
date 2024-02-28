package smb

import (
	"os"
	"strconv"
	"testing"

	env "github.com/vilasle/backilli/pkg/fs/environment"
	"github.com/vilasle/backilli/pkg/fs/unit"
)

func TestWrite(t *testing.T) {
	if err := env.LoadEnvfile("test.env"); err != nil {
		t.Fatal(err)
	}
	// TODO need to fix tests
	_ = env.Get("PATH_SRC")
	pathSrc := env.Get("PATH_DST")
	address := env.Get("ADDRESS")

	port, err := strconv.Atoi(env.Get("PORT"))
	if err != nil {
		t.Fatal(err)
	}
	domain := env.Get("DOMAIN")
	user := env.Get("USER")
	password := env.Get("PASSWORD")
	mountPoint := env.Get("MOUNT_POINT")

	wd, err := os.OpenFile(pathSrc, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	testSet := getTestSet()

	_, err = wd.Write(testSet)
	if err != nil {
		t.Fatal(err)
	}
	wd.Close()
	defer os.Remove(pathSrc)

	client, err := NewClient(unit.ClientConfig{
		Host:       address,
		Port:       port,
		Domain:     domain,
		User:       user,
		Password:   password,
		MountPoint: mountPoint,
		Root:       "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	// TODO need to fix tests
	// _, err = client.Write(pathSrc, path)
	// if err != nil {
	// 	t.Fatal(err)
	// }
}

func TestRead(t *testing.T) {
	if err := env.LoadEnvfile("test.env"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatal(r)
		}
	}()

	testSet := getTestSet()

	path := env.Get("PATH_SRC")
	address := env.Get("ADDRESS")

	port, err := strconv.Atoi(env.Get("PORT"))
	if err != nil {
		t.Fatal(err)
	}
	domain := env.Get("DOMAIN")
	user := env.Get("USER")
	password := env.Get("PASSWORD")
	mountPoint := env.Get("MOUNT_POINT")

	client, err := NewClient(unit.ClientConfig{
		Host:       address,
		Port:       port,
		Domain:     domain,
		User:       user,
		Password:   password,
		MountPoint: mountPoint,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	res, err := client.Read(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(res) != len(testSet) {
		t.Fatal("lenght data from file does not with test sets")
	}

	for i, s := range testSet {
		c := res[i]
		if c != s {
			t.Fatalf("byte with index %d does not match with test sets. Expected %v. There is %v", i, s, c)
		}
	}
}

func TestLs(t *testing.T) {
	if err := env.LoadEnvfile("test.env"); err != nil {
		t.Fatal(err)
	}

	path := env.Get("PATH_DIR")
	address := env.Get("ADDRESS")

	port, err := strconv.Atoi(env.Get("PORT"))
	if err != nil {
		t.Fatal(err)
	}
	domain := env.Get("DOMAIN")
	user := env.Get("USER")
	password := env.Get("PASSWORD")
	mountPoint := env.Get("MOUNT_POINT")

	client, err := NewClient(unit.ClientConfig{
		Host:       address,
		Port:       port,
		Domain:     domain,
		User:       user,
		Password:   password,
		MountPoint: mountPoint,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	pwdLs := []unit.File{
		{Name: "test.txt"},
		{Name: "test1.txt"},
	}

	ls, err := client.Ls(path)

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

	path := env.Get("PATH_SRC")

	address := env.Get("ADDRESS")

	port, err := strconv.Atoi(env.Get("PORT"))
	if err != nil {
		t.Fatal(err)
	}
	domain := env.Get("DOMAIN")
	user := env.Get("USER")
	password := env.Get("PASSWORD")
	mountPoint := env.Get("MOUNT_POINT")

	defer func() {
		if r := recover(); r != nil {
			t.Fatal(r)
		}
	}()

	client, err := NewClient(unit.ClientConfig{
		Host:       address,
		Port:       port,
		Domain:     domain,
		User:       user,
		Password:   password,
		MountPoint: mountPoint,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	if err := client.Remove(path); err != nil {
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
