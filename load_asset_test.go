package phoenix

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

type LoadAssetSuite struct {
	suite.Suite
	allocatedFiles []string
}

func (s *LoadAssetSuite) SetupTest() {
	s.allocatedFiles = []string{}
}

func (s *LoadAssetSuite) TearDownTest() {
	for _, filename := range s.allocatedFiles {
		err := os.Remove(filename)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (s *LoadAssetSuite) TestWhenAFileIsEmpty() {
	root, filename, err := s.allocateFile("")
	if err != nil {
		log.Fatal(err)
	}

	asset, err := LoadAsset(root, filename)
	assert.Nil(s.T(), err, "There should not be an error returned when the file is empty")
	assert.False(s.T(), asset.IsValid(), "The returned asset should not be considered valid")
}

func (s *LoadAssetSuite) TestWhenAFileContainsTextData() {
	root, filename, err := s.allocateFile("hello world")
	if err != nil {
		log.Fatal(err)
	}

	asset, err := LoadAsset(root, filename)
	assert.Nil(s.T(), err, "There should not be an error returned")
	assert.True(s.T(), asset.IsValid(), "Files that contain data should be valid")
	assert.Equal(s.T(), "hello world", asset.Value)
}

func (s *LoadAssetSuite) TestWhenAFileContainsBinaryData() {
	root, filename, err := s.allocateFile(string(BinaryTestData()))
	if err != nil {
		log.Fatal(err)
	}

	asset, err := LoadAsset(root, filename)
	assert.Nil(s.T(), err, "There should not be an error returned")
	assert.True(s.T(), asset.IsValid(), "Files that contain data should be valid")
	assert.True(s.T(), len(asset.Attachment) > 0, "The attachment should not be blank")
}

func (s *LoadAssetSuite) TestWhenFileIsADirectory() {
	root, filename, err := s.allocateDir()
	if err != nil {
		log.Fatal(err)
	}

	asset, err := LoadAsset(root, filename)
	assert.NotNil(s.T(), err, "The error should not be nil if a directory was given to LoadAsset")
	assert.Equal(s.T(), "File is a directory", err.Error())
	assert.False(s.T(), asset.IsValid(), "The asset returned should not be valid")
}

func (s *LoadAssetSuite) allocateFile(content string) (root, filename string, err error) {
	file, err := ioutil.TempFile("", "load-asset-suite-test-file")
	if err != nil {
		return
	}

	if len(content) > 0 {
		file.Write([]byte(content))
		file.Sync()
		file.Seek(0, 0)
	}

	s.noteAllocatedFile(file.Name())

	root = filepath.Dir(file.Name())
	filename = filepath.Base(file.Name())
	return
}

func (s *LoadAssetSuite) allocateDir() (root, filename string, err error) {
	dir, err := ioutil.TempDir("", "load-asset-suite-test-dir")
	if err != nil {
		return
	}

	s.noteAllocatedFile(dir)

	root = filepath.Dir(dir)
	filename = filepath.Base(dir)
	return
}

func (s *LoadAssetSuite) noteAllocatedFile(name string) {
	s.allocatedFiles = append(s.allocatedFiles, name)
}

func TestLoadAssetSuite(t *testing.T) {
	LoadAsset("foo", "bar")
	suite.Run(t, new(LoadAssetSuite))
}
