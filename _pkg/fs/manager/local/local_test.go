package local

import (
	"fmt"
	"os"
	"testing"

	env "github.com/vilasle/backilli/pkg/fs/environment"
	"github.com/vilasle/backilli/pkg/fs/unit"
)

func TestWrite(t *testing.T) {
	if err := env.LoadEnvFile("test.env"); err != nil {
		t.Fatal(err)
	}
	pathSrc := env.Get("PATH_SRC")
	pathDst := env.Get("PATH_DST")

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
	defer os.Remove(pathDst)

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// TODO need to fix tests
	_ = LocalClient{root: pwd}

	// _, err = localClient.Write(pathSrc, pathDst)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// cnt, err := os.ReadFile(pathDst)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// if len(cnt) != len(testSet) {
	// 	t.Fatal("lenght data from file does not with test sets")
	// }

	// for i, s := range testSet {
	// 	c := cnt[i]
	// 	if c != s {
	// 		t.Fatalf("byte with index %d does not match with test sets. Expected %v. There is %v", i, s, c)
	// 	}
	// }
}

func TestRead(t *testing.T) {
	if err := env.LoadEnvFile("test.env"); err != nil {
		t.Fatal(err)
	}
	path := env.Get("PATH_SRC")

	fd, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := os.Remove(path)
		if err != nil {
			fmt.Println(err)
		}
		if r := recover(); r != nil {
			t.Fatal(r)
		}
	}()

	testSet := getTestSet()

	_, err = fd.Write(testSet)
	if err != nil {
		t.Fatal(err)
	}
	fd.Close()

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	localClient := LocalClient{root: pwd}

	res, err := localClient.Read(path)
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
	if err := env.LoadEnvFile("test.env"); err != nil {
		t.Fatal(err)
	}
	path := env.Get("PATH_DIR")
	localClient := LocalClient{}
	ls, err := localClient.Ls(path)
	if err != nil {
		t.Fatal(err)
	}

	pwdLs := []unit.File{
		{Name: "test.env"},
		{Name: "local.go"},
		{Name: "local_test.go"},
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
	if err := env.LoadEnvFile("test.env"); err != nil {
		t.Fatal(err)
	}
	path := env.Get("PATH_SRC")

	fd, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatal(r)
		}
	}()

	testSet := getTestSet()

	_, err = fd.Write(testSet)
	if err != nil {
		t.Fatal(err)
	}
	fd.Close()

	localClient := LocalClient{}
	if err := localClient.Remove(path); err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(path)
	if !os.IsNotExist(err) {
		t.Fatal("file is exist but can be removed")
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
