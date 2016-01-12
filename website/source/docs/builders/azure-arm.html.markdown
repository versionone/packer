---
description: |
    The `azure-arm` Packer builder creates Azure Virtual Machines using Azure Resource Manager.
layout: docs
page_title: 'Azure Resource Manager'
...

# Azure Resource Manager

Type: `azure-arm`

The `azure-arm` Packer builder creates Azure Virtual Machines using [Azure Resource Manager](https://azure.microsoft.com/en-us/documentation/articles/resource-group-overview/), in conjunction with their associated network and storage resources. Essentially, the following resources are encapsulated:

- Compute Resource Provider
- Storage Resource Provider
- Network Resource Provider

Because Packer is focused on building and provisioning VMs rather than deploying them, some options from these providers (like load balancers) are omitted. Once Packer has created a virtual machine it may be further customized and deployed using [Terraform](https://www.terraform.io/docs/providers/azure/index.html) or the Azure web interface.

**Note vs. Classic Azure**: Resource Manager introduces a new model for managing machines which associates storage and network resources with the VM. This model is incompatible with Classic Azure. For more information please refer to [Understanding Resource Manager deployment and classic deployment](https://azure.microsoft.com/en-us/documentation/articles/resource-manager-deployment-model/)

## Configuration Reference

There are many configuration options available for the builder. They are
segmented below into two categories: required and optional parameters.

In addition to the options listed here, a
[communicator](/docs/templates/communicator.html) can be configured for this
builder.

### Required:

-   `auth_token` (string) - OAuth token used to authenticate against the API.
-   `image` (string) - The base image to build from, such as Ubuntu or Windows.
-   `size` (string) - The type of instance class to use, such as `Basic_A1` or `Standard_D4`.
-   `location` (string) - Which region the VM will be built in, such as West US

### Optional:

-   `volume_size` (integer) - The size of the volume, in GiB. Required if not
    specifying a `snapshot_id`
-   `delete_on_termination` (boolean) - Indicates whether the EBS volume is
    deleted on instance termination

## Basic Example

Here is a basic example. It is completely valid except for the access keys:

``` {.javascript}
"builders": [
    {
        "type":"azure-arm",
        "image": "",
        "size": "Basic_A1",
        "location": "West US",
        "username":"",
        "password":"",
        "subnet":"",
        "virtual_network":"",
        "storage_service_name":"",
        // Additional OS-specific options
        "communicator":"ssh",
        // Additional communicator options
    }
]
```

TODO

## Accessing the Instance to Debug

If you need to access the instance to debug for some reason, run the builder
with the `-debug` flag.

TODO