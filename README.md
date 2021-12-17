[![Go Report Card](https://goreportcard.com/badge/github.com/belgaied2/harvester-cli)](https://goreportcard.com/report/github.com/belgaied2/harvester-cli)
[![Lint Go Code](https://github.com/belgaied2/harvester-cli/actions/workflows/lint.yml/badge.svg)](https://github.com/belgaied2/harvester-cli/actions/workflows/lint.yml)

# CLI Tool for [Harvester HCI](https://harvesterhci.io)
This repository aims at providing a CLI tool to easily create VMs on Harvester, using simplicity principals similar to [multipass](https://multipass.run/) or [k3d](https://k3d.io).

# Usage
Harvester CLI is a command line tool that can manage some aspects of Virtual Machine management on a remote Harvester Cluster. The number of features implemented at the moment are still limited, but most of the handling of VMs themselves are already done.

Please be aware that Harvester CLI offers an opinionated approach to creating VMs, it is supposed to be a way to easily create and destroy test VMs for the purpose of conducting tests.
For instance, if no VM image is provided to the `harvester vm create` command, Harvester will go ahead and use the first image it finds in Harvester. If Harvester has no image, it will go ahead and download Ubuntu Focal `20.04` in its minimal version.

Similarly, if no Cloud-Init template is provided, it will go ahead and create a standard Cloud-Init template for Ubuntu.

If you consider the behavior to be problematic, please suggest a detailed proposal in a GitHub Issue.

## Features implemented
At the moment, features implemented in Harvester CLI are:
- Automatic Harvester Configuration Download from Rancher API
- VM Lifecycle Management: List, Create, Delete, Start, Stop, Restart
- Direct Shell access to VMs

Many aspects might be implemented in the future, like Network Management or VM Image Management, please feel free to contribute.

## VM Management


### harvester virtualmachine (alias vm)

> Manage Virtual Machines on Harvester
>
> name: harvester vm

```
NAME:
   harvester vm - Manage Virtual Machines on Harvester

USAGE:
   harvester vm command [command options] [arguments...]

COMMANDS:
   list, ls         List VMs
   delete, del, rm  Delete a VM
   create, c        Create a VM
   stop             Stop a VM
   start            Start a VM
   restart          Restart a VM

OPTIONS:
   --help, -h  show help


```

### harvester vm list (alias l)

> List VMs
>
> name: harvester vm list

List all VMs in the current Harvester Cluster

```
NAME:
   harvester vm list - List VMs

USAGE:
   harvester vm list [command options] None

DESCRIPTION:

List all VMs in the current Harvester Cluster

OPTIONS:
   --namespace value, -n value  Namespace of the VM (default: "default") [%HARVESTER_VM_NAMESPACE%]


```

### harvester vm delete

> Delete a VM
>
> name: harvester vm delete

```
NAME:
   harvester vm delete - Delete a VM

USAGE:
   harvester vm delete [command options] [VM_NAME/VM_ID]

OPTIONS:
   --namespace value, -n value  Namespace of the VM (default: "default") [%HARVESTER_VM_NAMESPACE%]


```

### harvester vm create

> Create a VM
>
> name: harvester vm create

```
NAME:
   harvester vm create - Create a VM

USAGE:
   harvester vm create [command options] [VM_NAME]

OPTIONS:
   --namespace value, -n value      Namespace of the VM (default: "default") [%HARVESTER_VM_NAMESPACE%]
   --vm-description value           Optional description of your VM [%HARVESTER_VM_DESCRIPTION%]
   --vm-image-id value              Harvester Image ID of the VM to create [%HARVESTER_VM_IMAGE_ID%]
   --disk-size value                Size of the primary VM disk (default: "10Gi") [%HARVESTER_VM_DISKSIZE%]
   --ssh-keyname value              KeyName of the SSH Key to use with this VM [%HARVESTER_VM_KEY%]
   --cpus value                     Number of CPUs to dedicate to the VM (default: 1) [%HARVESTER_VM_CPUS%]
   --memory value                   Amount of memory in the format XXGi (default: "1Gi") [%HARVESTER_VM_MEMORY%]
   --cloud-init-user-data value     Cloud Init User Data in yaml format [%HARVESTER_USER_DATA%]
   --cloud-init-network-data value  Cloud Init Network Data in yaml format [%HARVESTER_NETWORK_DATA%]


```

### harvester vm stop

> Stop a VM
>
> name: harvester vm stop

```
NAME:
   harvester vm stop - Stop a VM

USAGE:
   harvester vm stop [VM_NAME]

```

### harvester vm start

> Start a VM
>
> name: harvester vm start

```
NAME:
   harvester vm start - Start a VM

USAGE:
   harvester vm start [VM_NAME]

```

### harvester vm restart

> Restart a VM
>
> name: harvester vm restart

```
NAME:
   harvester vm restart - Restart a VM

USAGE:
   harvester vm restart [command options] [VM_NAME]

OPTIONS:
   --vm-name value, --name value  Name of the VM to restart


```

## Automatic Configuration download from Rancher
In order to get Harvester's Kubeconfig to be able to manage your particular Harvester Cluster, you have :
- The manual way: get the KUBECONFIG file from the underlying RKE2 Cluster and put it on your client, then reference it in the `harvester` commands 
- Download the KUBECONFIG file from Rancher using a Rancher API token: this is done using the `harvester get-config` command.

*TODO* : Improve this section


# DISCLAIMER
This is still in a very early stage of development.
