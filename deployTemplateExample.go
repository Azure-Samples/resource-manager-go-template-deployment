// An example illustrating how to use Go to deploy an Azure Resource Manager Template.
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"

	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/go-autorest/autorest/azure"
)

func main() {
	fmt.Println("Azure Resource Manager Template Deployment Sample")
	err := deploy("exampleresourcegroup")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

// deploy creates a resource group, and deploys a template there.
func deploy(resourceGroupName string, sshKeyPath ...string) error {
	dnsPrefix := "sample-dns-prefix"
	fmt.Println("Get credentials and token...")
	credentials := map[string]string{
		"AZURE_CLIENT_ID":       os.Getenv("AZURE_CLIENT_ID"),
		"AZURE_CLIENT_SECRET":   os.Getenv("AZURE_CLIENT_SECRET"),
		"AZURE_SUBSCRIPTION_ID": os.Getenv("AZURE_SUBSCRIPTION_ID"),
		"AZURE_TENANT_ID":       os.Getenv("AZURE_TENANT_ID")}
	if err := checkEnvVar(&credentials); err != nil {
		return err
	}
	oauthConfig, err := azure.PublicCloud.OAuthConfigForTenant(credentials["AZURE_TENANT_ID"])
	if err != nil {
		return err
	}
	token, err := azure.NewServicePrincipalToken(*oauthConfig, credentials["AZURE_CLIENT_ID"], credentials["AZURE_CLIENT_SECRET"], azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return err
	}

	// Ensure the resource group is created.
	fmt.Println("Create resource group...")
	groupClient := resources.NewGroupsClient(credentials["AZURE_SUBSCRIPTION_ID"])
	groupClient.Authorizer = token
	location := "westus"
	var resourceGroupParameters = resources.ResourceGroup{
		Location: &location}
	if _, err = groupClient.CreateOrUpdate(resourceGroupName, resourceGroupParameters); err != nil {
		return err
	}

	// Map template
	fmt.Println("Get template...")
	template, err := parseJSONFromFile("vmDeploymentTemplate.json")
	if err != nil {
		return err
	}

	//TemplateLink
	/*
		templateURI := "https://raw.githubusercontent.com/Azure-Samples/resource-manager-go-template-deployment/master/vmDeploymentTemplate.json"
		var templateLink = resources.TemplateLink{
			URI:            &templateURI,
			ContentVersion: nil}
	*/

	fmt.Println("Get parameters...")
	parameter := map[string]interface{}{}
	// The paramaters map must have this format {key: {"value": value}}.
	addElementToMap(&parameter, "dnsLabelPrefix", dnsPrefix)
	addElementToMap(&parameter, "vmName", "azure-deployment-sample-vm")
	sshKey, err := getSSHkey(&sshKeyPath)
	if err != nil {
		return err
	}
	addElementToMap(&parameter, "sshKeyData", sshKey)

	// ParameterLink
	/*
		parameterURI := "https://raw.githubusercontent.com/Azure-Samples/resource-manager-go-template-deployment/master/vmDeploymentParameter.json"
		var parameterLink = resources.ParametersLink{
			URI:            &parameterURI,
			ContentVersion: nil}
	*/

	// Set the template or templateLink, and parameters or parametersLink to use.
	var properties = resources.DeploymentProperties{
		Template:       template,
		TemplateLink:   nil,
		Parameters:     &parameter,
		ParametersLink: nil,
		Mode:           resources.Incremental}
	var parameters = resources.Deployment{
		Properties: &properties}

	deploymentsClient := resources.NewDeploymentsClient(credentials["AZURE_SUBSCRIPTION_ID"])
	deploymentsClient.Authorizer = token

	fmt.Println("Deploying...")
	if _, err = deploymentsClient.CreateOrUpdate(resourceGroupName, "azure-sample", parameters, nil); err != nil {
		return err
	}
	fmt.Println("Finished deployment")
	fmt.Printf("You can connect via ssh azureSample@%v.%v.cloudapp.azure.com\n", dnsPrefix, location)

	// Clean up after the deployment.
	/*
		fmt.Println("Deleting resource group...")
		if _, err = groupClient.Delete(resourceGroupName, nil); err != nil {
			return err
		}
		fmt.Println("Finished deleting resource group")
	*/

	return nil
}

// parseJSONFromFile recieves a JSON file path, and Unmarshals the file into a map[string]interface{}.
func parseJSONFromFile(filePath string) (*map[string]interface{}, error) {
	text, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	fileMap := map[string]interface{}{}
	if err = json.Unmarshal(text, &fileMap); err != nil {
		return nil, err
	}
	return &fileMap, nil
}

// getSSHkey receives a SSH key file path, and returns the key as a string.
func getSSHkey(sshKeyPath *[]string) (string, error) {
	path := ""
	if len(*sshKeyPath) == 0 {
		path = ".ssh/id_rsa.pub"
	} else {
		path = (*sshKeyPath)[0]
	}
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	sshKey, err := ioutil.ReadFile(usr.HomeDir + "/" + path)
	if err != nil {
		return "", err
	}
	return string(sshKey), nil
}

// addElementToMap adds the key and value to the map specific format required for Azure template deployment parameters.
func addElementToMap(parameter *map[string]interface{}, key string, value interface{}) {
	(*parameter)[key] = map[string]interface{}{
		"value": value,
	}
}

// checkEnvVar checks if the environment variables are actually set.
func checkEnvVar(envVars *map[string]string) error {
	var missingVars []string
	for varName, value := range *envVars {
		if value == "" {
			missingVars = append(missingVars, varName)
		}
	}
	if len(missingVars) > 0 {
		return fmt.Errorf("Missing environment variables %v", missingVars)
	}
	return nil
}
