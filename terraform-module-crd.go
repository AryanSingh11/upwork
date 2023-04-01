package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func main() {
	// Define your Kubernetes CRD data here
	crd := `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: myresource
spec:
  group: mygroup.example.com
  version: v1alpha1
  names:
    plural: myresources
    singular: myresource
    kind: MyResource
  scope: Namespaced
  validation:
    openAPIV3Schema:
      properties:
        spec:
          type: object
          properties:
            size:
              type: integer
            config:
              type: object
              properties:
                replicas:
                  type: integer
                  default: 1
                resources:
                  type: object
                  properties:
                    cpu:
                      type: string
                      format: "resource"
                      default: "100m"
                    memory:
                      type: string
                      format: "resource"
                      default: "256Mi"
              required:
                - replicas
`

	// Parse the CRD's JSON schema from the YAML definition
	jsonSchema, err := getJSONSchema(crd)
	if err != nil {
		fmt.Println("Error getting JSON schema from CRD YAML:", err)
		os.Exit(1)
	}

	// Generate the Terraform module code
	moduleCode := generateModule(jsonSchema)

	// Print the generated module code to standard output
	fmt.Println(moduleCode)
}

// Generate Terraform module code from a JSON schema
func generateModule(jsonSchema map[string]interface{}) string {
	// Define the Terraform module header
	moduleCode := `variable "namespace" {
  description = "The namespace for the resource"
}

variable "name" {
  description = "The name of the resource"
}

`

	// Generate Terraform code for each field in the JSON schema
	for fieldName, fieldSchema := range jsonSchema["properties"].(map[string]interface{}) {
		moduleCode += generateFieldCode(fieldName, fieldSchema)
	}

	// Define the Terraform resource block for the custom resource
	moduleCode += `resource "myresource" "%[1]s" {
`

	// Add the Terraform variables for each field in the JSON schema
	for fieldName, fieldSchema := range jsonSchema["properties"].(map[string]interface{}) {
		moduleCode += generateVariableCode(fieldName, fieldSchema)
	}

	// Close the Terraform resource block
	moduleCode += `
  metadata {
    namespace = var.namespace
    name = var.name
  }
}
`

	// Replace the %[1]s placeholder with the resource name
	resourceName := jsonSchema["title"].(string)
	moduleCode = fmt.Sprintf(moduleCode, strings.ToLower(resourceName))

	return moduleCode
}

// Generate Terraform code for a field in a JSON schema
func generateFieldCode(fieldName string, fieldSchema interface{}) string {
	// Skip fields that are not objects or have no properties
	if fieldSchema.(map[string]interface{})["type"].(string) != "object" ||
		fieldSchema.(map[string]interface{})["properties"] == nil {
		return ""
