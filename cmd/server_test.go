package cmd

import (
  "net/http"
	"net/http/httptest"
	"testing"
  "io/ioutil"
  "encoding/json"
  "bytes"
  "encoding/base64"
  "fmt"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestHealth(t *testing.T) {
  assert := assert.New(t)

  e := echo.New()

  req := httptest.NewRequest(http.MethodGet, "/health", nil)
  rec := httptest.NewRecorder()

  e.GET("/health", health)
  e.ServeHTTP(rec, req)

  assert.Equal(http.StatusOK, rec.Code)
  assert.Equal("success", rec.Body.String())
}

func TestMutate(t *testing.T) {
  assert := assert.New(t)

  e := echo.New()

  var (
    blob map[string]interface{}
  )

  serverConfig.Sources = "example.org"
  serverConfig.Target = "example.com"

  jsonBlob, err := ioutil.ReadFile("testdata/admissionreview.json")
  if err != nil {
    t.Fatal(err)
  }

  req := httptest.NewRequest(http.MethodPost, "/mutate", bytes.NewReader(jsonBlob))
  rec := httptest.NewRecorder()

  e.POST("/mutate", mutate)
  e.ServeHTTP(rec, req)

  if err = json.Unmarshal([]byte(rec.Body.String()), &blob); err != nil {
		t.Fatal(err)
	}

  patch := fmt.Sprintf("%v", blob["response"].(map[string]interface{})["patch"])
  patchDecoded, err := base64.StdEncoding.DecodeString(patch)
  if err != nil {
    t.Fatal(err)
  }

  assert.Equal(http.StatusOK, rec.Code)
  assert.Equal(`[{"op":"replace","path":"/spec/rules/0/host","value":"muting.example.com"}]`, string(patchDecoded))
}
