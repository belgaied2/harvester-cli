package cmd

import (
	"context"
	"fmt"

	"github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	harvclient "github.com/harvester/harvester/pkg/generated/clientset/versioned"
	rcmd "github.com/rancher/cli/cmd"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	VMv1 "kubevirt.io/client-go/api/v1"
	"kubevirt.io/client-go/kubecli"
	"kubevirt.io/containerized-data-importer/pkg/apis/core/v1alpha1"
)

const (
	vmAnnotationDescription = "field.cattle.io/description"
	vmAnnotationNetworkIps  = "networks.harvesterhci.io/ips"
	dvAnnotationImageID     = "harvesterhci.io/imageId"
	dvSourceHTTPURLPrefix   = "http://minio.harvester-system:9000/vm-images/"
	defaultSSHUser          = "ubuntu"
	defaultVMName           = "test-vm"
	defaultVMDescription    = "Test VM for Harvester"
	defaultDiskSize         = "10Gi"
	defaultMemSize          = "1Gi"
	defaultNbCPUCores       = 1
	defaultNamespace        = "default"
	ubuntuDefaultImage      = "https://cloud-images.ubuntu.com/minimal/daily/focal/current/focal-minimal-cloudimg-amd64.img"
)

var (
	nsFlag = cli.StringFlag{
		Name:   "namespace, n",
		Usage:  "Namespace of the VM",
		EnvVar: "HARVESTER_VM_NAMESPACE",
		Value:  "default",
	}
)

type VirtualMachineData struct {
	State          string
	VirtualMachine VMv1.VirtualMachine
	Name           string
	Node           string
	CPU            uint32
	Memory         string
	IPAddress      string
}

type Client struct {
	HarvesterClient *harvclient.Clientset
	KubevirtClient  *kubecli.KubevirtClient
}

func VMCommand() cli.Command {
	return cli.Command{
		Name:    "virtualmachine",
		Aliases: []string{"vm"},
		Usage:   "Manage Virtual Machines on Harvester",
		Action:  defaultAction(vmLs),
		Subcommands: []cli.Command{
			{
				Name:        "list",
				Usage:       "List VMs",
				Aliases:     []string{"ls"},
				Description: "\nList all VMs in the current Harvester Cluster",
				ArgsUsage:   "None",
				Action:      vmLs,
				Flags: []cli.Flag{
					nsFlag,
				},
			},
			{
				Name: "delete",
				Aliases: []string{
					"del",
					"rm",
				},
				Usage:     "Delete a VM",
				Action:    vmDelete,
				ArgsUsage: "[VM_NAME/VM_ID]",
				Flags: []cli.Flag{
					nsFlag,
				},
			},
			{
				Name: "create",
				Aliases: []string{
					"c",
				},
				Usage:     "Create a VM",
				Action:    vmCreate,
				ArgsUsage: "[VM_NAME]",
				Flags: []cli.Flag{
					nsFlag,
					cli.StringFlag{
						Name:   "vm-description",
						Usage:  "Optional description of your VM",
						EnvVar: "HARVESTER_VM_DESCRIPTION",
						Value:  "",
					},
					cli.StringFlag{
						Name:   "vm-image-id",
						Usage:  "Harvester Image ID of the VM to create",
						EnvVar: "HARVESTER_VM_IMAGE_ID",
						Value:  "",
					},
					cli.StringFlag{
						Name:   "disk-size",
						Usage:  "Size of the primary VM disk",
						EnvVar: "HARVESTER_VM_DISKSIZE",
						Value:  defaultDiskSize,
					},
					cli.StringFlag{
						Name:   "ssh-keyname",
						Usage:  "KeyName of the SSH Key to use with this VM",
						EnvVar: "HARVESTER_VM_KEY",
						Value:  "",
					},
					cli.IntFlag{
						Name:   "cpus",
						Usage:  "Number of CPUs to dedicate to the VM",
						EnvVar: "HARVESTER_VM_CPUS",
						Value:  defaultNbCPUCores,
					},
					cli.StringFlag{
						Name:   "memory",
						Usage:  "Amount of memory in the format XXGi",
						EnvVar: "HARVESTER_VM_MEMORY",
						Value:  defaultMemSize,
					},
				},
			},
			{
				Name:      "stop",
				Usage:     "Stop a VM",
				Action:    vmStop,
				ArgsUsage: "[VM_NAME]",
			},
			{
				Name:      "start",
				Usage:     "Start a VM",
				Action:    vmStart,
				ArgsUsage: "[VM_NAME]",
			},
			{
				Name:      "restart",
				Usage:     "Restart a VM",
				Action:    vmRestart,
				ArgsUsage: "[VM_NAME]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "vm-name, name",
						Usage: "Name of the VM to restart",
					},
				},
			},
		},
	}
}

func defaultAction(fn func(ctx *cli.Context) error) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		if ctx.Bool("help") {
			cli.ShowAppHelp(ctx)
			return nil
		}
		return fn(ctx)
	}
}

func vmLs(ctx *cli.Context) error {

	c, err := GetHarvesterClient()

	if err != nil {
		return err
	}

	vmList, err := (*c.KubevirtClient).VirtualMachine("default").List(&k8smetav1.ListOptions{})

	if err != nil {
		return err
	}

	vmiList, err := (*c.KubevirtClient).VirtualMachineInstance("default").List(&k8smetav1.ListOptions{})

	if err != nil {
		return err
	}

	vmiMap := map[string]VMv1.VirtualMachineInstance{}
	for _, vmi := range vmiList.Items {
		vmiMap[vmi.Name] = vmi
	}

	writer := rcmd.NewTableWriter([][]string{
		{"STATE", "State"},
		{"NAME", "Name"},
		{"NODE", "Node"},
		{"CPU", "CPU"},
		{"RAM", "Memory"},
		{"IP Address", "IPAddress"},
	},
		ctx)

	defer writer.Close()

	for _, vm := range vmList.Items {

		running := *vm.Spec.Running
		var state string
		if running {
			state = "Running"
		} else {
			state = "Not Running"
		}

		var IP string
		if vmiMap[vm.Name].Status.Interfaces == nil {
			IP = ""
		} else {
			IP = vmiMap[vm.Name].Status.Interfaces[0].IP
		}

		writer.Write(&VirtualMachineData{
			State:          state,
			VirtualMachine: vm,
			Name:           vm.Name,
			Node:           vmiMap[vm.Name].Status.NodeName,
			CPU:            vm.Spec.Template.Spec.Domain.CPU.Cores,
			Memory:         vm.Spec.Template.Spec.Domain.Resources.Requests.Memory().String(),
			IPAddress:      IP,
		})

	}

	return writer.Err()
}

func vmDelete(ctx *cli.Context) error {
	c, err := GetHarvesterClient()

	if err != nil {
		return err
	}

	vmName := ctx.Args().First()

	return (*c.KubevirtClient).VirtualMachine(ctx.String("namespace")).Delete(vmName, &k8smetav1.DeleteOptions{})
}

func vmCreate(ctx *cli.Context) error {

	c, err := GetHarvesterClient()

	if err != nil {
		return err
	}

	// Checking existence of Image ID and if not, using default ubuntu image.
	imageID := ctx.String("vm-image-id")
	if imageID != "" {
		_, err := (*c.HarvesterClient).HarvesterhciV1beta1().VirtualMachineImages(ctx.String("namespace")).Get(context.TODO(), imageID, k8smetav1.GetOptions{})
		if err != nil {
			return err
		}
		logrus.Debug("Image ID %s given does exist!", ctx.String("vm-image-id"))
	} else {
		setDefaultVMImage(ctx)
	}

	// Checking existing of the SSH KeyPair
	keyName := ctx.String("ssh-keyname")
	if keyName != "" {
		_, err := (*c.HarvesterClient).HarvesterhciV1beta1().KeyPairs(ctx.String("namespace")).Get(context.TODO(), keyName, k8smetav1.GetOptions{})
		if err != nil {
			return err
		}
		logrus.Debug("Image ID %s given does exist!", ctx.String("ssh-keyname"))
	} else {
		setDefaultSSHKey(ctx)
	}

	sc := "longhorn-" + ctx.String("vm-image-id")
	dsAPIGroup := "storage.k8s.io"
	diskRandomID := RandomID()
	vmName := ctx.Args().First()
	pvcName := vmName + "-disk-0-" + diskRandomID

	vmLabels := map[string]string{
		"harvesterhci.io/creator": "harvester",
	}
	vmiLabels := vmLabels
	vmiLabels["harvesterhci.io/vmName"] = vmName
	volumeMode := v1.PersistentVolumeBlock

	cloudInitUserData, err := (*c.KubevirtClient).CoreV1().ConfigMaps(ctx.String("namespace")).Get(context.TODO(), "standard-ubuntu", k8smetav1.GetOptions{})

	if err != nil {
		return fmt.Errorf("the cloud init template specified does not exist")
	}

	sshKey, err := (*c.HarvesterClient).HarvesterhciV1beta1().KeyPairs(ctx.String("namespace")).Get(context.TODO(), ctx.String("ssh-keyname"), k8smetav1.GetOptions{})

	if err != nil {
		return err
	}

	cloudInitSSHSection := "\nssh_authorized_keys:\n  - " + sshKey.Spec.PublicKey + "\n"

	cloudInitNetworkData, err := (*c.KubevirtClient).CoreV1().ConfigMaps(ctx.String("namespace")).Get(context.TODO(), "ubuntu-std-network", k8smetav1.GetOptions{})

	if err != nil {
		return fmt.Errorf("the cloud init template specified does not exist")
	}
	logrus.Debug("CloudInit: ")

	ubuntuVM := &VMv1.VirtualMachine{
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      vmName,
			Namespace: ctx.String("namespace"),
			Annotations: map[string]string{
				vmAnnotationDescription: ctx.String("vm-description"),
				vmAnnotationNetworkIps:  "[]",
			},
			Labels: vmLabels,
		},
		Spec: VMv1.VirtualMachineSpec{
			Running: NewTrue(),
			DataVolumeTemplates: []VMv1.DataVolumeTemplateSpec{
				{
					ObjectMeta: k8smetav1.ObjectMeta{
						Name: pvcName,
						Annotations: map[string]string{
							dvAnnotationImageID: ctx.String("namespace") + "/" + ctx.String("vm-image-id"),
						},
					},
					Spec: v1alpha1.DataVolumeSpec{
						Source: v1alpha1.DataVolumeSource{
							Blank: &v1alpha1.DataVolumeBlankImage{},
						},
						PVC: &v1.PersistentVolumeClaimSpec{
							AccessModes: []v1.PersistentVolumeAccessMode{
								"ReadWriteOnce",
							},
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									"storage": resource.MustParse(ctx.String("disk-size")),
								},
							},

							StorageClassName: &sc,
							VolumeMode:       &volumeMode,

							DataSource: &v1.TypedLocalObjectReference{
								APIGroup: &dsAPIGroup,
							},
						},
					},
					// DataSource: nil,

				},
			},

			Template: &VMv1.VirtualMachineInstanceTemplateSpec{
				ObjectMeta: k8smetav1.ObjectMeta{
					Annotations: vmiAnnotations(pvcName, ctx.String("ssh-keyname")),
					Labels:      vmiLabels,
				},
				Spec: VMv1.VirtualMachineInstanceSpec{
					Hostname: vmName,
					Networks: []VMv1.Network{

						{
							Name: "nic-1",

							NetworkSource: VMv1.NetworkSource{
								Multus: &VMv1.MultusNetwork{
									NetworkName: "vlan1",
								},
							},
						},
					},
					Volumes: []VMv1.Volume{
						{
							Name: "disk-0",
							VolumeSource: VMv1.VolumeSource{
								DataVolume: &VMv1.DataVolumeSource{
									Name: pvcName,
								},
							},
						},
						{
							Name: "cloudinitdisk",
							VolumeSource: VMv1.VolumeSource{
								CloudInitNoCloud: &VMv1.CloudInitNoCloudSource{
									UserData:    cloudInitUserData.Data["cloudInit"] + cloudInitSSHSection,
									NetworkData: cloudInitNetworkData.Data["cloudInit"],
								},
							},
						},
					},
					Domain: VMv1.DomainSpec{
						CPU: &VMv1.CPU{
							Cores:   uint32(ctx.Int("cpus")),
							Sockets: uint32(ctx.Int("cpus")),
							Threads: uint32(ctx.Int("cpus")),
						},
						Devices: VMv1.Devices{
							Inputs: []VMv1.Input{
								{
									Bus:  "usb",
									Type: "tablet",
									Name: "tablet",
								},
							},
							Interfaces: []VMv1.Interface{
								{
									Name:                   "nic-1",
									Model:                  "virtio",
									InterfaceBindingMethod: VMv1.DefaultBridgeNetworkInterface().InterfaceBindingMethod,
								},
							},
							Disks: []VMv1.Disk{
								{
									Name: "disk-0",
									DiskDevice: VMv1.DiskDevice{
										Disk: &VMv1.DiskTarget{
											Bus: "virtio",
										},
									},
								},
								{
									Name: "cloudinitdisk",
									DiskDevice: VMv1.DiskDevice{
										Disk: &VMv1.DiskTarget{
											Bus: "virtio",
										},
									},
								},
							},
						},
						Resources: VMv1.ResourceRequirements{
							Requests: v1.ResourceList{
								"memory": resource.MustParse(ctx.String("memory")),
							},
						},
					},
				},
			},
		},
	}

	if err != nil {
		return err
	}

	_, err = (*c.KubevirtClient).VirtualMachine("default").Create(ubuntuVM)

	if err != nil {
		return err
	}

	return nil
}

// vmStart issues a power on for the virtual machine instance.
func vmStart(ctx *cli.Context) error {

	c, err := GetHarvesterClient()
	if err != nil {
		return err
	}

	vm, err := (*c.KubevirtClient).VirtualMachine("default").Get(ctx.Args().First(), &k8smetav1.GetOptions{})

	*vm.Spec.Running = true

	if err != nil {
		return err
	}

	_, err = (*c.KubevirtClient).VirtualMachine("default").Update(vm)
	return err
}

// Stop issues a power off for the virtual machine instance.
func vmStop(ctx *cli.Context) error {

	c, err := GetHarvesterClient()
	if err != nil {
		return err
	}

	vm, err := (*c.KubevirtClient).VirtualMachine("default").Get(ctx.Args().First(), &k8smetav1.GetOptions{})
	*vm.Spec.Running = false

	if err != nil {
		return err
	}

	_, err = (*c.KubevirtClient).VirtualMachine("default").Update(vm)
	return err
}

// Restart reboots the virtual machine instance.
func vmRestart(ctx *cli.Context) error {

	err := vmStop(ctx)
	if err != nil {
		return err
	}
	return vmStart(ctx)
}

func vmiAnnotations(pvcName string, sshKeyName string) map[string]string {
	return map[string]string{
		"harvesterhci.io/diskNames": "[\"" + pvcName + "\"]",
		"harvesterhci.io/sshNames":  "[\"" + sshKeyName + "\"]",
	}
}

func setDefaultVMImage(ctx *cli.Context) error {
	c, err := GetHarvesterClient()

	if err != nil {
		return err
	}

	vmImages, err := (*c.HarvesterClient).HarvesterhciV1beta1().VirtualMachineImages("default").List(context.TODO(), k8smetav1.ListOptions{})

	if err != nil {
		return err
	}

	var vmImage *v1beta1.VirtualMachineImage

	if len(vmImages.Items) == 0 {
		vmImage, err = CreateVMImage("ubuntu-default-image", ubuntuDefaultImage)
		if err != nil {
			return fmt.Errorf("impossible to create a default VM Image")
		}
	} else {
		vmImage = &vmImages.Items[0]
	}

	imageID := vmImage.ObjectMeta.Name
	ctx.Set("vm-image-id", imageID)

	return nil
}

func setDefaultSSHKey(ctx *cli.Context) error {
	c, err := GetHarvesterClient()

	if err != nil {
		return err
	}

	sshKeys, err := (*c.HarvesterClient).HarvesterhciV1beta1().KeyPairs("default").List(context.TODO(), k8smetav1.ListOptions{})

	if err != nil {
		return err
	}

	if len(sshKeys.Items) == 0 {

		return fmt.Errorf("no ssh keys exists in harvester, please add a new ssh key")
	} else {
		sshKey := &sshKeys.Items[0]
		ctx.Set("ssh-keyname", sshKey.Name)
	}

	return nil
}

func CreateVMImage(imageName string, url string) (*v1beta1.VirtualMachineImage, error) {
	c, err := GetHarvesterClient()

	if err != nil {
		return &v1beta1.VirtualMachineImage{}, err
	}

	vmImage, err := (*c.HarvesterClient).HarvesterhciV1beta1().VirtualMachineImages("default").Create(
		context.TODO(),
		&v1beta1.VirtualMachineImage{
			ObjectMeta: k8smetav1.ObjectMeta{
				Name: "ubuntu-default",
			},
			Spec: v1beta1.VirtualMachineImageSpec{
				DisplayName: imageName,
				URL:         url,
			},
		},
		k8smetav1.CreateOptions{})

	if err != nil {
		return &v1beta1.VirtualMachineImage{}, err
	}

	return vmImage, nil
}
