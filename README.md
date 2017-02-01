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
- [What does example.go do?](#sample)
- [More information](#info)

<a id="run"></a>

## Run this sample

1. If you don't already have it, [install Go 1.7](https://golang.org/dl/).

1. Clone the repository.

    ```
    git clone https://github.com:Azure-Samples/virtual-machines-go-manage.git
    ```

1. Install the dependencies using glide.

    ```
    cd virtual-machines-go-manage
    glide install
    ```

1. Create an Azure service principal either through
    [Azure CLI](https://azure.microsoft.com/documentation/articles/resource-group-authenticate-service-principal-cli/),
    [PowerShell](https://azure.microsoft.com/documentation/articles/resource-group-authenticate-service-principal/)
    or [the portal](https://azure.microsoft.com/documentation/articles/resource-group-create-service-principal-portal/).

1. Set the following environment variables using the information from the service principle that you created.

    ```
    export AZURE_TENANT_ID={your tenant id}
    export AZURE_CLIENT_ID={your client id}
    export AZURE_CLIENT_SECRET={your client secret}
    export AZURE_SUBSCRIPTION_ID={your subscription id}
    ```

    > [AZURE.NOTE] On Windows, use `set` instead of `export`.

1. Run the sample.

    ```
    go run example.go
    ```


<a id="paramlink"></a>

### ParamaterLink instructions

1. Create an [Azure Key Vault](https://azure.microsoft.com/documentation/articles/key-vault-manage-with-cli/) using the [Azure CLI](https://azure.microsoft.com/documentation/articles/xplat-cli-install/) `azure keyvault create --vault-name templateSampleVault --resource-group MyResourceGroup --location westus`

1. Add the public SSH as secret in the vault using the Azure CLI

	```
	azure keyvault secret set --vault-name templateSampleVault --secret-name sshKeyData --value "yourPublicSSHkey"
	```	

1. Reference correctly the [Key Vault in the parameter file](https://azure.microsoft.com/documentation/articles/resource-manager-keyvault-parameter/). In vmDeploymentParameter.json, replace subscription_id, resource_group and vault_name inside the "id" value with the correct values.

<a id="sample"></a>

## What does example.go do?

A rescorce group is needed to be able to deploy a template.

```go
	resourceGroupParameters := resources.ResourceGroup{
		Location: to.StringPtr(location),
	}
	_, err := groupsClient.CreateOrUpdate(groupName, resourceGroupParameters)
```

The sample then gets the template and its parameters. Both template and parameters can be set with a `*map[string]interface{}` or with a link to a json file.

```go
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
```

The template is validated...

```go
	validate, err := deploymentsClient.Validate(groupName, deploymentName, deployment)
	onErrorFail(err, "Validate failed")
	if validate.Error == nil {
		fmt.Println("Deployment is validated! Template is syntactically correct")
	} else {
		printValidationError(validate)
		os.Exit(1)
	}
```

Finally, the template is deployed!

```go
	_, err := deploymentsClient.CreateOrUpdate(groupName, deploymentName, deployment, nil)
```

<a id="info"></a>

## More information

- [Azure Resource Manager overview - Template deployment](https://azure.microsoft.com/documentation/articles/resource-group-overview/#template-deployment)
- [Create a template deployment](https://msdn.microsoft.com/library/azure/dn790564.aspx)
- [Resource Manager template walkthrough](https://azure.microsoft.com/documentation/articles/resource-manager-template-walkthrough/)
- [Pass secure values during deployment](https://azure.microsoft.com/documentation/articles/resource-manager-keyvault-parameter/)
- [Using linked templates with Azure Resource Manager](https://docs.microsoft.com/en-us/azure/azure-resource-manager/resource-group-linked-templates)

***

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/). For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.