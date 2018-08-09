package shopify

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/src/env"
)

func TestFindAssets(t *testing.T) {
	goodEnv := &env.Env{Directory: filepath.Join("_testdata", "project")}
	badEnv := &env.Env{Directory: "nope"}

	testcases := []struct {
		e      *env.Env
		inputs []string
		err    string
		count  int
	}{
		{e: goodEnv, inputs: []string{filepath.Join("assets", "application.js")}, count: 1},
		{e: goodEnv, count: 7},
		{e: badEnv, count: 7, err: "no such file or directory"},
		{e: goodEnv, inputs: []string{"assets", "config/settings_data.json"}, count: 3},
		{e: goodEnv, inputs: []string{"snippets/nope.txt"}, err: "no such file or directory"},
	}

	for _, testcase := range testcases {
		assets, err := FindAssets(testcase.e, testcase.inputs...)
		if testcase.err == "" {
			assert.Nil(t, err)
			assert.Equal(t, testcase.count, len(assets))
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}
}

func TestAsset_Write(t *testing.T) {
	testDir := filepath.Join("_testdata", "writeto")
	os.Mkdir(testDir, 0755)

	testcases := []struct {
		filename, outdir, err string
	}{
		{outdir: testDir, filename: "blah.txt"},
		{outdir: "nothere", filename: "blah.txt", err: "file or directory"},
		{outdir: testDir, filename: filepath.Join("assets", "test.txt")},
	}

	for _, testcase := range testcases {
		err := Asset{Key: testcase.filename}.Write(testcase.outdir)
		if testcase.err == "" {
			assert.Nil(t, err)
			_, err := os.Stat(filepath.Join(testDir, testcase.filename))
			assert.Nil(t, err)
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}

	os.RemoveAll(testDir)
}

func TestAsset_Contents(t *testing.T) {
	testcases := []struct {
		asset  Asset
		err    string
		length int
	}{
		{asset: Asset{Value: "this is content"}, length: 15},
		{asset: Asset{Attachment: "this is bad content"}, err: "Could not decode"},
		{asset: Asset{Attachment: base64.StdEncoding.EncodeToString([]byte("this is good content"))}, length: 20},
		{asset: Asset{Key: "test.json", Value: "{\"test\":\"one\"}"}, length: 19},
	}

	for _, testcase := range testcases {
		data, err := testcase.asset.contents()
		if testcase.err == "" {
			assert.Nil(t, err)
			assert.Equal(t, testcase.length, len(data))
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}
}

func TestLoadAssetsFromDirectory(t *testing.T) {
	root := filepath.Join("_testdata", "project")
	ignoreNone := func(path string) bool { return strings.Contains(path, ".gitkeep") }
	selectOne := func(path string) bool { return path != filepath.Join("assets", "application.js") }

	testcases := []struct {
		path, err string
		ignore    func(string) bool
		count     int
	}{
		{path: "", ignore: ignoreNone, count: 7},
		{path: "", ignore: selectOne, count: 1},
		{path: "assets", ignore: ignoreNone, count: 2},
		{path: "nope", ignore: ignoreNone, count: 0, err: "no such file or directory"},
	}

	for _, testcase := range testcases {
		assets, err := loadAssetsFromDirectory(root, testcase.path, testcase.ignore)
		if testcase.err == "" {
			assert.Nil(t, err)
			assert.Equal(t, testcase.count, len(assets))
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}
}

func TestReadAsset(t *testing.T) {
	e := &env.Env{Directory: filepath.Join("_testdata", "project")}

	testcases := []struct {
		input    string
		expected Asset
		err      string
	}{
		{input: filepath.Join("assets", "application.js"), expected: Asset{Key: "assets/application.js", Value: "this is js content\n"}},
		{input: filepath.Join(".", "assets", "application.js"), expected: Asset{Key: "assets/application.js", Value: "this is js content\n"}},
		{input: "nope.txt", expected: Asset{}, err: "no such file"},
		{input: "assets", expected: Asset{}, err: ErrAssetIsDir.Error()},
		{input: filepath.Join("assets", "image.png"), expected: Asset{Key: "assets/image.png", Attachment: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAEUlEQVR4nGJiYGBgAAQAAP//AA8AA/6P688AAAAASUVORK5CYII="}},
	}

	for _, testcase := range testcases {
		actual, err := ReadAsset(e, testcase.input)
		assert.Equal(t, testcase.expected, actual)
		if testcase.err == "" {
			assert.Nil(t, err)
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}
}
