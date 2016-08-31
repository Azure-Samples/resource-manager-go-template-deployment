---
services: azure-resource-manager
platforms: go
author: mcardosos
---

#Deploy an SSH Enabled VM with a Template in Go

This example demonstrates how to use Go to deploy an Azure Resource Manager Template. If you don't have a Microsoft Azure subscription you can get a FREE trial account [here](https://azure.microsoft.com/pricing/free-trial).

**On this page**

- [Run this sample](#run)
  - [ParamaterLink instructions](#paramlink)
- [What does deployTemplateExample.go do?](#sample)
- [More information](#info)

<a id="run"></a>
## Run this sample

1. Create a [service principal](https://azure.microsoft.com/documentation/articles/resource-group-authenticate-service-principal-cli/). You will need the Tenant ID, Client ID and Client Secret for [authentication](https://github.com/Azure/azure-sdk-for-go/tree/master/arm#first-a-sidenote-authentication-and-the-azure-resource-manager), so keep them as soon as you get them.
2. Get your Azure Subscription ID using either of the methods mentioned below:
  - Get it through the [portal](portal.azure.com) in the subscriptions section.
  - Get it using the [Azure CLI](https://azure.microsoft.com/documentation/articles/xplat-cli-install/) with command `azure account show`.
  - Get it using [Azure Powershell](https://azure.microsoft.com/documentation/articles/powershell-install-configure/) whit cmdlet `Get-AzureRmSubscription`.
3. Set environment variables `AZURE_TENANT_ID = <TENANT_ID>`, `AZURE_CLIENT_ID = <CLIENT_ID>`, `AZURE_CLIENT_SECRET = <CLIENT_SECRET>` and `AZURE_SUBSCRIPTION_ID = <SUBSCRIPTION_ID>`.
4. Get this sample using command `go get -u github.com/Azure-Samples/resource-manager-go-template-deployment`.
5. Get the [Azure SDK for Go](https://github.com/Azure/azure-sdk-for-go) using command `go get -u github.com/Azure/azure-sdk-for-go`. Or in case that you want to vendor your dependencies using [glide](https://github.com/Masterminds/glide), navigate to this sample's directory and use command `glide install`
6. Compile and run the sample.

<a id="paramlink"></a>
### ParamaterLink instructions

1. Create an [Azure Key Vault](https://azure.microsoft.com/documentation/articles/key-vault-manage-with-cli/) using the [Azure CLI](https://azure.microsoft.com/documentation/articles/xplat-cli-install/) `azure keyvault create --vault-name templateSampleVault --resource-group MyResourceGroup --location westus`
2. Add the public SSH as secret in the vault `azure keyvault secret set --vault-name templateSampleVault --secret-name sshKeyData --value "yourPublicSSHkey"`
3. Reference correctly the [Key Vault in the parameter file](https://azure.microsoft.com/documentation/articles/resource-manager-keyvault-parameter/). In vmDeploymentParameter.json, replace subscription_id, resource_group and vault_name inside the "id" value with the correct values.

<a id="sample"></a>
## What does deployTemplateExample.go do?

First, the sample gets an authentication token using your Azure credentials. This token will be included in all clients (GroupsClient and DeploymentsClient).

```go
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
```

A rescorce group is needed to be able to deploy a template.

```go
	groupClient := resources.NewGroupsClient(credentials["AZURE_SUBSCRIPTION_ID"])
	groupClient.Authorizer = token
	location := "westus"
	var resourceGroupParameters = resources.ResourceGroup{
		Location: &location}
	if _, err = groupClient.CreateOrUpdate(resourceGroupName, resourceGroupParameters); err != nil {
		return err
	}
```

The sample then gets the template and its parameters. Both template and parameters can be set with a `*map[string]interface{}` or with a link to a json file.

```go
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
```

Finally, the template is deployed!

```go
deploymentsClient := resources.NewDeploymentsClient(credentials["AZURE_SUBSCRIPTION_ID"])
deploymentsClient.Authorizer = token
if _, err = deploymentsClient.CreateOrUpdate(resourceGroupName, "azure-sample", parameters, nil); err != nil {
	return err
}
```

<a id="info"></a>
## More information

- [Azure Resource Manager overview - Template deployment](https://azure.microsoft.com/documentation/articles/resource-group-overview/#template-deployment)
- [Create a template deployment](https://msdn.microsoft.com/library/azure/dn790564.aspx)
- [Resource Manager template walkthrough](https://azure.microsoft.com/documentation/articles/resource-manager-template-walkthrough/)
- [Pass secure values during deployment](https://azure.microsoft.com/documentation/articles/resource-manager-keyvault-parameter/)

***

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/). For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.