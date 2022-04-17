package mutator

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/admission/v1"
)

func getTestData(t *testing.T, file string) []byte {
	t.Helper()
	data, err := ioutil.ReadFile(filepath.Join("testdata", file))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestMutate(t *testing.T) {

	tc := []struct {
		name          string
		testdata      string
		sourceDomains string
		targetDomain  string
		patches       []*Patch
		err           bool
		errType       interface{}
	}{
		{
			name:          "valid request single rule",
			testdata:      "valid-request-single-rule.json",
			sourceDomains: "test.one",
			targetDomain:  "test.two",
			patches: []*Patch{
				{"replace", "/spec/rules/0/host", "muting.test.two"},
			},
			err: false,
		},
		{
			name:          "valid request multi rule",
			testdata:      "valid-request-multi-rule.json",
			sourceDomains: "test.one",
			targetDomain:  "test.two",
			patches: []*Patch{
				{"replace", "/spec/rules/0/host", "muting-a.test.two"},
				{"replace", "/spec/rules/1/host", "muting-b.test.two"},
			},
			err: false,
		},
		{
			name:          "invalid request empty AdmissionReview.Request",
			testdata:      "invalid-request-empty-request.json",
			sourceDomains: "test.one",
			targetDomain:  "test.two",
			patches:       []*Patch{},
			err:           true,
			errType:       &BadRequest{},
		},
		{
			name:          "invalid request invalid ingress",
			testdata:      "invalid-request-empty-request.json",
			sourceDomains: "test.one",
			targetDomain:  "test.two",
			patches:       []*Patch{},
			err:           true,
			errType:       &BadRequest{},
		},
		{
			name:          "invalid request json",
			testdata:      "invalid-request-json.json",
			sourceDomains: "test.one",
			targetDomain:  "test.two",
			patches:       []*Patch{},
			err:           true,
			errType:       &BadRequest{},
		},
		{
			name:     "empty base domain",
			testdata: "valid-request-single-rule.json",
			patches:  []*Patch{},
			err:      true,
			errType:  errors.New(""),
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {

			// execute the test
			request := getTestData(t, test.testdata)
			respBody, err := Mutate(request, test.sourceDomains, test.targetDomain)

			// validate error if error expected
			if test.err {
				assert.Error(t, err)
				assert.IsType(t, test.errType, err)
				return
			}

			// fail the test if we have an unexpected error
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			// validate response
			admReview := v1.AdmissionReview{}
			err = json.Unmarshal(respBody, &admReview)
			assert.NoError(t, err)
			resp := admReview.Response
			expectedPatch, err := json.Marshal(test.patches)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, string(expectedPatch), string(resp.Patch))
			assert.Equal(t, resp.AuditAnnotations["mutated-host"], "true")

		})
	}

}
