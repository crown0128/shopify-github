package kit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/Shopify/themekit/theme"
)

type HttpClientTestSuite struct {
	suite.Suite
	config Configuration
	client *httpClient
}

func (suite *HttpClientTestSuite) SetupTest() {
	suite.config, _ = NewConfiguration()
	suite.config.Domain = "test.myshopify.com"
	suite.config.ThemeID = "123"
	suite.config.Password = "sharknado"
	suite.client, _ = newHTTPClient(suite.config)
}

func (suite *HttpClientTestSuite) TestNewHttpClient() {
	assert.Equal(suite.T(), suite.config, suite.client.config)
	assert.Equal(suite.T(), suite.config.Timeout, suite.client.client.Timeout)

	config, _ := NewConfiguration()
	config.Proxy = "://abc!21@"
	client, err := newHTTPClient(config)
	assert.NotNil(suite.T(), err)

	config, _ = NewConfiguration()
	config.Proxy = "http://localhost:3000"
	client, err = newHTTPClient(config)
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), client.client.Transport)
}

func (suite *HttpClientTestSuite) TestAdminURL() {
	assert.Equal(suite.T(),
		fmt.Sprintf("https://%s/admin/themes/%v", suite.config.Domain, suite.config.ThemeID),
		suite.client.AdminURL())

	suite.client.config.ThemeID = "live"

	assert.Equal(suite.T(),
		fmt.Sprintf("https://%s/admin", suite.config.Domain),
		suite.client.AdminURL())
}

func (suite *HttpClientTestSuite) TestAssetPath() {
	assert.Equal(suite.T(),
		fmt.Sprintf("%s/assets.json", suite.client.AdminURL()),
		suite.client.AssetPath())
}

func (suite *HttpClientTestSuite) TestThemesPath() {
	assert.Equal(suite.T(),
		fmt.Sprintf("%s/themes.json", suite.client.AdminURL()),
		suite.client.ThemesPath())
}

func (suite *HttpClientTestSuite) TestThemePath() {
	assert.Equal(suite.T(),
		fmt.Sprintf("%s/themes/456.json", suite.client.AdminURL()),
		suite.client.ThemePath(456))
}

func (suite *HttpClientTestSuite) TestAssetQuery() {
	server := suite.NewTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "GET", r.Method)
		assert.Equal(suite.T(), "fields=key,attachment,value", r.URL.RawQuery)
		fmt.Fprintf(w, jsonFixture("http_client/multi_asset"))
	})
	resp, err := suite.client.AssetQuery(Retrieve, map[string]string{})
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), listRequest, resp.Type)
	assert.Equal(suite.T(), 2, len(resp.Assets))
	server.Close()

	server = suite.NewTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "GET", r.Method)
		assert.Equal(suite.T(), "fields=key,attachment,value&asset[key]=file.txt", r.URL.RawQuery)
		fmt.Fprintf(w, jsonFixture("http_client/single_asset"))
	})
	resp, err = suite.client.AssetQuery(Retrieve, map[string]string{"asset[key]": "file.txt"})
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), assetRequest, resp.Type)
	assert.Equal(suite.T(), "assets/hello.txt", resp.Asset.Key)
	server.Close()
}

func (suite *HttpClientTestSuite) TestNewTheme() {
	server := suite.NewTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "POST", r.Method)

		decoder := json.NewDecoder(r.Body)
		var t map[string]theme.Theme
		decoder.Decode(&t)
		defer r.Body.Close()

		assert.Equal(suite.T(), theme.Theme{Name: "name", Source: "source", Role: "unpublished"}, t["theme"])
		fmt.Fprintf(w, jsonFixture("http_client/theme"))
	})
	defer server.Close()
	resp, err := suite.client.NewTheme("name", "source")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), themeRequest, resp.Type)
	assert.Equal(suite.T(), "timberland", resp.Theme.Name)
}

func (suite *HttpClientTestSuite) TestGetTheme() {
	server := suite.NewTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "GET", r.Method)
		fmt.Fprintf(w, jsonFixture("http_client/theme"))
	})
	resp, err := suite.client.GetTheme(123)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), themeRequest, resp.Type)
	assert.Equal(suite.T(), "timberland", resp.Theme.Name)
	server.Close()
}

func (suite *HttpClientTestSuite) TestAssetAction() {
	server := suite.NewTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "PUT", r.Method)

		decoder := json.NewDecoder(r.Body)
		var t map[string]theme.Asset
		decoder.Decode(&t)
		defer r.Body.Close()

		assert.Equal(suite.T(), theme.Asset{Key: "key", Value: "value"}, t["asset"])
		fmt.Fprintf(w, jsonFixture("http_client/single_asset"))
	})
	defer server.Close()
	resp, err := suite.client.AssetAction(Update, theme.Asset{Key: "key", Value: "value"})
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), assetRequest, resp.Type)
	assert.Equal(suite.T(), "key", resp.Asset.Key)
}

func (suite *HttpClientTestSuite) TestNewRequest() {
	req, err := suite.client.newRequest(Update, suite.config.Domain, nil)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.config.Password, req.Header.Get("X-Shopify-Access-Token"))
	assert.Equal(suite.T(), "application/json", req.Header.Get("Content-Type"))
	assert.Equal(suite.T(), "application/json", req.Header.Get("Accept"))
	assert.Equal(suite.T(), fmt.Sprintf("go/themekit (%s; %s)", runtime.GOOS, runtime.GOARCH), req.Header.Get("User-Agent"))

	_, err = suite.client.newRequest(Update, "://#nksd", nil)
	assert.NotNil(suite.T(), err)
}

func (suite *HttpClientTestSuite) TestSendJSON() {
	server := suite.NewTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "PUT", r.Method)

		decoder := json.NewDecoder(r.Body)
		var t map[string]string
		decoder.Decode(&t)
		defer r.Body.Close()

		assert.Equal(suite.T(), "mystring", t["test"])
	})
	defer server.Close()
	resp, err := suite.client.sendJSON(assetRequest, Update, server.URL, map[string]interface{}{"test": "mystring"})
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), assetRequest, resp.Type)
}

func (suite *HttpClientTestSuite) TestSendRequest() {
	server := suite.NewTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "PUT", r.Method)

		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)

		assert.Equal(suite.T(), "my string", buf.String())
	})
	defer server.Close()
	resp, err := suite.client.sendRequest(assetRequest, Update, server.URL, bytes.NewBufferString("my string"))
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), assetRequest, resp.Type)
}

func (suite *HttpClientTestSuite) NewTestServer(handler http.HandlerFunc) *httptest.Server {
	server := httptest.NewServer(handler)
	suite.client.config.Domain = server.URL
	suite.client.insecure = true
	return server
}

func TestHttpClientTestSuite(t *testing.T) {
	suite.Run(t, new(HttpClientTestSuite))
}

func jsonFixture(name string) string {
	path := fmt.Sprintf("../fixtures/%s.json", name)
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	return string(bytes)
}
