[![Go Report Card](https://goreportcard.com/badge/github.com/belgaied2/harvester-cli)](https://goreportcard.com/report/github.com/belgaied2/harvester-cli)
[![Lint Go Code](https://github.com/belgaied2/harvester-cli/actions/workflows/lint.yml/badge.svg)](https://github.com/belgaied2/harvester-cli/actions/workflows/lint.yml)

# CLI Tool for [Harvester HCI](https://harvesterhci.io)
Harvester CLI is a command line tool that can manage some aspects of Virtual Machine management on a remote Harvester Cluster, using simplicity principles similar to [multipass](https://multipass.run/) or [k3d](https://k3d.io).


# Quick setup
Harvester CLI needs a KUBECONFIG to access remotely the Harvester Cluster. There are two ways to achieve that:

## Inline configuration path
Harvester KUBECONFIG files can be manually given to Harvester using a CLI flag or a by setting an environment variable:
- CLI Flag: `--config`
- Environment variable : `HARVESTER_CONFIG`

## Automatic download from Rancher Server
Since a Harvester Cluster can be imported into Rancher like any other Kubernetes cluster, it is possible to use the Rancher API to automatically download the KUBECONFIG file for Harvester. This approach is particularly useful if multiple Harvester Clusters are managed by a single Rancher instances.

Two commands are needed for this.
- First login to Rancher Server: `harvester login https://<RANCHER_URL>` -t <RANCHER_API_TOKEN>
- Then, get the KUBECONFIG for the target Harvester Cluster: `harvester get-config <HARVESTER_CLUSTER_NAME>`, this will download the KUBECONFIG file for target Harvester cluster to `$HOME/.harvester/config`, which is the default location used by the CLI.

Next, a check can be done by listing the available VMs:
```bash
harvester vm list
```

# Default behavior when creating VMs
Please be aware that Harvester CLI offers an opinionated approach to creating VMs, it is supposed to be a way to easily create and destroy test VMs for the purpose of conducting tests.
For instance, *if no VM image is provided* to the `harvester vm create` command, `harvester` CLI will go ahead and use the *first image it finds in Harvester*. If Harvester has no image, it will go ahead and download Ubuntu Focal `20.04` in its minimal version.

Similarly, if no Cloud-Init template is provided, it will go ahead and create a standard Cloud-Init template for Ubuntu.

If you consider the behavior to be problematic, please suggest a detailed proposal in a GitHub Issue.

# Features implemented
At the moment, features implemented in Harvester CLI are:
- Automatic Harvester Configuration Download from Rancher API
- VM Lifecycle Management: List, Create, Delete, Start, Stop, Restart
- Direct Shell access to VMs

Many aspects might be implemented in the future, like Network Management or VM Image Management, please feel free to contribute or suggest features by creating issues.

# VM Management


### harvester virtualmachine (alias vm)

The `harvester vm` command has a number of possible sub-commands for VM Lifecycle management including `list`, `create`, `delete`, `start`, `stop` and `restart`. The default behavior if no sub-command is given is the `list` sub-command. 

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
The `list` sub-command lists all VMs available in the Harvester cluster. At the moment, no filtering feature is implemented.

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

### harvester vm create
The `create` sub-command creates a VM based on some input preferences, the most important of which are: 
- `cpus` flag: Number of CPUs
- `memory` flag: Memory size using a notation such as : 4G ( 4 x 10^9 Bytes ) or 8Gi ( 8 x 1024^3 Bytes)
- `disk-size` flag: Disk size using the same notation as above
- `vm-image-id` flag: references the VM image (should be a Cloud Image type of image) that already exists on Harvester. *NOTE: At this time, it is necessary to give the VM ID and not the image name. The ID can be found in the Harvester UI in the YAML description of the VM Image*

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

### harvester vm delete
The `delete` sub-command deletes the VM which name corresponds to the first argument that follows. For the moment, if multiple arguments are given, only the first one will be used.

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

### harvester vm stop
The `stop` sub-command stops the VM which name corresponds to the first argument that follows. For the moment, if multiple arguments are given, only the first one will be used.

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
The `start` sub-command starts the VM which name corresponds to the first argument that follows. For the moment, if multiple arguments are given, only the first one will be used.

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
The `restart` sub-command restarts the VM which name corresponds to the first argument that follows. For the moment, if multiple arguments are given, only the first one will be used.

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
