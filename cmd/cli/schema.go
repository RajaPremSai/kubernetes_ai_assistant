package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

func fetchK8sSchema() (map[string]interface{}, error) {
	var body []byte
	var err error
	if *k8sOpenAPIURL == "" {
		log.Debugf("Fetching schema from kubernertes API server")
		kubeConfig := getKubeConfig()
		body, err = runKubectlCommand("get", "--raw", "/openapi/v2", "--kubeconfig", kubeConfig)
		if err != nil {
			return nil, err
		}
	} else {
		log.Debugf("fetching schema from %s", *k8sOpenAPIURL)
		response, err := http.Get(*k8sOpenAPIURL)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()
		body, err = io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
	}
	var schema map[string]interface{}
	err = json.Unmarshal(body, &schema)
	if err != nil {
		return nil, err
	}
	return schema, nil
}

func fetchResourceNames(resourceName string) ([]string, error) {
	schema, err := fetchK8sSchema()
	if err != nil {
		return nil, err
	}
	log.Debugf("fetching resource name %s", resourceName)
	definitions, ok := schema["definations"].(map[string]interface{})
	if !ok {
		return nil, errors.New("unable to assert schema definations")
	}
	var resourceNames []string
	for k := range definitions {
		if strings.Contains(strings.ToLower(k), strings.ToLower(resourceName)) {
			resourceNames = append(resourceNames, k)
		}
	}
	return resourceNames, nil
}

func fetchSchemaForResource(resourceType string) (map[string]interface{}, error) {
	schema, err := fetchK8sSchema()
	if err != nil {
		return nil, err
	}
	definitions, ok := schema["definations"].(map[string]interface{})
	if !ok {
		return nil, errors.New("unable to assert schema definations")
	}
	log.Debugf("fetching resourceSchema %s", resourceType)
	if resourceSchema, ok := definitions[resourceType]; ok {
		rs, ok := resourceSchema.(map[string]interface{})
		if !ok {
			return nil, errors.New("unbale to assert resource schema")
		}
		return rs, nil
	}
	return nil, errors.New("unable to find resource schema")
}

func runKubectlCommand(args ...string) ([]byte, error) {
	cmd := exec.Command("kubectl", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
