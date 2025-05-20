package helpers

import (
	"os"
	"path"
	"runtime"
)

var fixturesFolder = "" // nolint:gochecknoglobals

func init() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("error recovering caller information in test helper")
	}
	fixturesFolder = path.Join(path.Dir(filename), "../fixtures")
}

func GetFixturePath(name string) string {
	return path.Join(fixturesFolder, name)
}

func ReadFixture(name string) []byte {
	data, err := os.ReadFile(GetFixturePath(name))
	if err != nil {
		panic(err)
	}
	return data
}

func ReadFixtureString(name string) string {
	return string(ReadFixture(name))
}
