package mutator

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	admission "k8s.io/api/admission/v1"
	networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Patch represents a JSON patch
type Patch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

// BadRequest is used to allow the caller to return an appropriate http response
type BadRequest struct {
	err string
}

func (e *BadRequest) Error() string {
	return fmt.Sprintf("Bad Request: %s", e.err)
}

// Mutate receives an http request body (AdmissionReview), and baseDomain.
// It adds an AdmissionResponse to the AdmissionReview and then returns it.
// Its goal is to create a JSON patch to append the baseDomain to the host
// values in a given ingress resource
func Mutate(body []byte, sourceDomains string, targetDomain string) ([]byte, error) {

	// prevent an empty sourceDomains
	if sourceDomains == "" {
		return nil, fmt.Errorf("Received empty source domains")
	}

	// prevent an empty targetDomain
	if targetDomain == "" {
		return nil, fmt.Errorf("Received empty target domain")
	}

	// unmarshal the request
	admReview := admission.AdmissionReview{}
	if err := json.Unmarshal(body, &admReview); err != nil {
		return nil, &BadRequest{fmt.Sprintf("Failed to unmarshal AdmissionReview: %s", err.Error())}
	}
	request := admReview.Request

	// handle an empty request
	if request == nil {
		return nil, &BadRequest{"AdmissionReview.Request is nil"}
	}

	// get the ingress resource from the request
	var ingress *networking.Ingress
	if err := json.Unmarshal(request.Object.Raw, &ingress); err != nil {
		return nil, &BadRequest{fmt.Sprintf("Failed to unmarshal ingress from AdmissionRequest: %s", err.Error())}
	}

	// set the response options
	response := &admission.AdmissionResponse{}
	response.Allowed = true
	response.UID = request.UID
	patchType := admission.PatchTypeJSONPatch
	response.PatchType = &patchType
	response.AuditAnnotations = map[string]string{
		"mutated-host": "true",
	}

	// build a JSONPatch for each host rule
	var patches []*Patch
	for i, rule := range ingress.Spec.Rules {
		patches = append(patches, &Patch{
			Op:    "replace",
			Path:  fmt.Sprintf("/spec/rules/%d/host", i),
			Value: replaceDomain(rule.Host, strings.Split(sourceDomains, ","), targetDomain),
		})
	}

	// add the patches to the response
	jsonPatches, err := json.Marshal(patches)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal patches to JSON: %s", err)
	}
	response.Patch = jsonPatches

	// set the result as success
	response.Result = &metav1.Status{
		Status: "Success",
	}

	// Add the response to the AdmissionReview and return it
	admReview.Response = response
	responseBody, err := json.Marshal(admReview)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal AdmissionReview response to JSON: %s", err)
	}
	return responseBody, nil

}

func replaceDomain(host string, sources []string, target string) string {
	for _, suffix := range sources {
		if strings.HasSuffix(host, suffix) {
			re := regexp.MustCompile("."+suffix+"$")
			result := re.ReplaceAllString(host, "$1."+target)
			return result
		}
	}
	return host
}
