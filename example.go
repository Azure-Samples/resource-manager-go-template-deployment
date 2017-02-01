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
	"github.com/Azure/go-autorest/autorest/to"
)

// This example requires that the following environment vars are set:
//
// AZURE_TENANT_ID: contains your Azure Active Directory tenant ID or domain
// AZURE_CLIENT_ID: contains your Azure Active Directory Application Client ID
// AZURE_CLIENT_SECRET: contains your Azure Active Directory Application Secret
// AZURE_SUBSCRIPTION_ID: contains your Azure Subscription ID
//

var (
	// This sample will look for your ssh key in HOME/.ssh/id_rsa.pub
	// Specify sshKeyPath if different
	sshKeyPath string

	// Template Deployment allows you to use linked templates, or load a local file.
	// https://docs.microsoft.com/azure/azure-resource-manager/resource-group-linked-templates
	useTemplateLink  bool
	useParameterLink bool

	location       = "westus"
	groupName      = "your-azure-sample-group"
	dnsPrefix      = "sample-dns-prefix"
	deploymentName = "azure-sample"

	groupsClient      resources.GroupsClient
	deploymentsClient resources.DeploymentsClient
)

func init() {
	subscriptionID := getEnvVarOrExit("AZURE_SUBSCRIPTION_ID")
	tenantID := getEnvVarOrExit("AZURE_TENANT_ID")

	oauthConfig, err := azure.PublicCloud.OAuthConfigForTenant(tenantID)
	onErrorFail(err, "OAuthConfigForTenant failed")

	clientID := getEnvVarOrExit("AZURE_CLIENT_ID")
	clientSecret := getEnvVarOrExit("AZURE_CLIENT_SECRET")
	spToken, err := azure.NewServicePrincipalToken(*oauthConfig, clientID, clientSecret, azure.PublicCloud.ResourceManagerEndpoint)
	onErrorFail(err, "NewServicePrincipalToken failed")

	createClients(subscriptionID, spToken)
}

func main() {
	createResourceGroup()
	d := buildDeploymentParameters()
	validateDeployment(d)
	deploy(d)

	fmt.Print("Press enter to delete the resources created in this sample...")
	fmt.Scanln()

	deleteResourceGroup()
}

func createResourceGroup() {
	fmt.Println("Create resource group")
	resourceGroupParameters := resources.ResourceGroup{
		Location: to.StringPtr(location),
	}
	_, err := groupsClient.CreateOrUpdate(groupName, resourceGroupParameters)
	onErrorFail(err, "CreateOrUpdate failed")
}

// buildDeploymentParameters sets the deployment struct to be validated and deployed
func buildDeploymentParameters() resources.Deployment {
	fmt.Println("Build deployment parameters")
	deployment := resources.Deployment{
		Properties: &resources.DeploymentProperties{
			Mode: resources.Incremental,
		},
	}

	fmt.Println("\tGet template")
	if useTemplateLink {
		fmt.Println("\tUsing template link")
		deployment.Properties.TemplateLink = &resources.TemplateLink{
			URI:            to.StringPtr("https://raw.githubusercontent.com/Azure-Samples/resource-manager-go-template-deployment/master/vmDeploymentTemplate.json"),
			ContentVersion: to.StringPtr("1.0.0.0"),
		}
	} else {
		fmt.Println("\tUsing local template")
		template, err := parseJSONFromFile("vmDeploymentTemplate.json")
		onErrorFail(err, "parseJSONFromFile failed")
		deployment.Properties.Template = template
	}

	fmt.Println("\tGet parameters")
	if useParameterLink {
		fmt.Println("\tUsing parameter link")
		deployment.Properties.ParametersLink = &resources.ParametersLink{
			URI:            to.StringPtr("https://raw.githubusercontent.com/Azure-Samples/resource-manager-go-template-deployment/master/vmDeploymentParameter.json"),
			ContentVersion: to.StringPtr("1.0.0.0"),
		}
	} else {
		fmt.Println("\tUsing local parameters")
		parameter := map[string]interface{}{}
		// The paramaters map must have this format {key: {"value": value}}.
		addElementToMap(&parameter, "dnsLabelPrefix", dnsPrefix)
		addElementToMap(&parameter, "vmName", "azure-deployment-sample-vm")
		sshKey, err := getSSHkey(sshKeyPath)
		onErrorFail(err, "getSSHkey failed")
		addElementToMap(&parameter, "sshKeyData", sshKey)
		deployment.Properties.Parameters = &parameter
	}

	return deployment
}

// validateDeployment validates the template
func validateDeployment(deployment resources.Deployment) {
	fmt.Println("Validate deployment template")
	validate, err := deploymentsClient.Validate(groupName, deploymentName, deployment)
	onErrorFail(err, "Validate failed")
	if validate.Error == nil {
		fmt.Println("Deployment is validated! Template is syntactically correct")
	} else {
		printValidationError(validate)
		os.Exit(1)
	}
}

func deploy(deployment resources.Deployment) {
	fmt.Println("Deploy")
	_, err := deploymentsClient.CreateOrUpdate(groupName, deploymentName, deployment, nil)
	onErrorFail(err, "Deploy failed")
	fmt.Println("Finished deployment")
	fmt.Printf("You can connect via ssh azureSample@%v.%v.cloudapp.azure.com\n", dnsPrefix, location)
}

func deleteResourceGroup() {
	fmt.Println("Delete resource group")
	_, err := groupsClient.Delete(groupName, nil)
	onErrorFail(err, "Delete failed")
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
func getSSHkey(sshKeyPath string) (string, error) {
	var path string
	if sshKeyPath == "" {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		path = fmt.Sprintf("%s/.ssh/id_rsa.pub", usr.HomeDir)
	} else {
		path = sshKeyPath
	}

	sshKey, err := ioutil.ReadFile(path)
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

// getEnvVarOrExit returns the value of specified environment variable or terminates if it's not defined.
func getEnvVarOrExit(varName string) string {
	value := os.Getenv(varName)
	if value == "" {
		fmt.Printf("Missing environment variable %s\n", varName)
		os.Exit(1)
	}

	return value
}

// onErrorFail prints a failure message and exits the program if err is not nil.
func onErrorFail(err error, message string) {
	if err != nil {
		fmt.Printf("%s: %s\n", message, err)
		groupsClient.Delete(groupName, nil)
		os.Exit(1)
	}
}

func createClients(subscriptionID string, spToken *azure.ServicePrincipalToken) {
	groupsClient = resources.NewGroupsClient(subscriptionID)
	groupsClient.Authorizer = spToken

	deploymentsClient = resources.NewDeploymentsClient(subscriptionID)
	deploymentsClient.Authorizer = spToken
}

func printValidationError(validate resources.DeploymentValidateResult) {
	fmt.Printf("Error! Code: %s\nMessage: %s\nTarget: %s\n",
		printStringPtr(validate.Error.Code),
		printStringPtr(validate.Error.Message),
		printStringPtr(validate.Error.Target))
}

func printStringPtr(s *string) string {
	if s != nil && *s != "" {
		return *s
	}
	return "-"
}
