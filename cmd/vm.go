package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	harvclient "github.com/harvester/harvester/pkg/generated/clientset/versioned"
	rcmd "github.com/rancher/cli/cmd"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	VMv1 "kubevirt.io/client-go/api/v1"
)

const (
	vmAnnotationPVC             = "harvesterhci.io/volumeClaimTemplates"
	vmAnnotationNetworkIps      = "networks.harvesterhci.io/ips"
	defaultDiskSize             = "10Gi"
	defaultMemSize              = "1Gi"
	defaultNbCPUCores           = 1
	defaultNamespace            = "default"
	ubuntuDefaultImage          = "https://cloud-images.ubuntu.com/minimal/daily/focal/current/focal-minimal-cloudimg-amd64.img"
	defaultCloudInitUserData    = "#cloud-config\npassword: password\nchpasswd: { expire: False}\nssh_pwauth: True\npackages:\n  - qemu-guest-agent\nruncmd:\n  - [ systemctl, daemon-reload ]\n  - [ systemctl, enable, qemu-guest-agent.service ]\n  - [ systemctl, start, --no-block, qemu-guest-agent.service ]"
	defaultCloudInitNetworkData = "version: 2\nrenderer: networkd\nethernets:\n  enp1s0:\n    dhcp4: true\n  enp2s0:\n    dhcp4: true"
	defaultCloudInitCmPrefix    = "default-ubuntu-"
)

var (
	nsFlag = cli.StringFlag{
		Name:   "namespace, n",
		Usage:  "Namespace of the VM",
		EnvVar: "HARVESTER_VM_NAMESPACE",
		Value:  defaultNamespace,
	}
)

// VirtualMachineData type is a Data Structure that holds information to display for VM
type VirtualMachineData struct {
	State          string
	VirtualMachine VMv1.VirtualMachine
	Name           string
	Node           string
	CPU            uint32
	Memory         string
	IPAddress      string
}

// VMCommand defines the CLI command that manages VMs
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
						Name:   "disk-size, disk, d",
						Usage:  "Size of the primary VM disk",
						EnvVar: "HARVESTER_VM_DISKSIZE",
						Value:  defaultDiskSize,
					},
					cli.StringFlag{
						Name:   "ssh-keyname, i",
						Usage:  "KeyName of the SSH Key to use with this VM",
						EnvVar: "HARVESTER_VM_KEY",
						Value:  "",
					},
					cli.IntFlag{
						Name:   "cpus, c",
						Usage:  "Number of CPUs to dedicate to the VM",
						EnvVar: "HARVESTER_VM_CPUS",
						Value:  defaultNbCPUCores,
					},
					cli.StringFlag{
						Name:   "memory, m",
						Usage:  "Amount of memory in the format XXGi",
						EnvVar: "HARVESTER_VM_MEMORY",
						Value:  defaultMemSize,
					},
					cli.StringFlag{
						Name:   "cloud-init-user-data, user-data",
						Usage:  "Name of the Cloud Init User Data Template to be used",
						EnvVar: "HARVESTER_USER_DATA",
						Value:  "",
					},
					cli.StringFlag{
						Name:   "cloud-init-network-data, network-data",
						Usage:  "Name of the Cloud Init Network Data Template to be used",
						EnvVar: "HARVESTER_NETWORK_DATA",
						Value:  "",
					},
					cli.StringFlag{
						Name:   "template, from-template",
						Usage:  "Harvester VM Template to use for creating the VM in the format <template_name>:<version> or <template> in which case the latest version will be used",
						EnvVar: "HARVESTER_VM_TEMPLATE",
						Value:  "",
					},
				},
			},
			{
				Name:      "stop",
				Usage:     "Stop a VM",
				Action:    vmStop,
				ArgsUsage: "[VM_NAME]",
				Flags: []cli.Flag{
					nsFlag,
				},
			},
			{
				Name:      "start",
				Usage:     "Start a VM",
				Action:    vmStart,
				ArgsUsage: "[VM_NAME]",
				Flags: []cli.Flag{
					nsFlag,
				},
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
					nsFlag,
				},
			},
		},
	}
}

//vmLs lists the VMs available in Harvester
func vmLs(ctx *cli.Context) error {

	c, err := GetHarvesterClient(ctx)

	if err != nil {
		return err
	}

	vmList, err := c.KubevirtV1().VirtualMachines(ctx.String("namespace")).List(context.TODO(), k8smetav1.ListOptions{})

	if err != nil {
		return err
	}

	vmiList, err := c.KubevirtV1().VirtualMachineInstances(ctx.String("namespace")).List(context.TODO(), k8smetav1.ListOptions{})

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

//vmDelete deletes a VM which name is given in argument from Harvester
func vmDelete(ctx *cli.Context) error {
	c, err := GetHarvesterClient(ctx)

	if err != nil {
		return err
	}

	vmName := ctx.Args().First()

	return c.KubevirtV1().VirtualMachines(ctx.String("namespace")).Delete(context.TODO(), vmName, k8smetav1.DeleteOptions{})
}

// vmCreate implements the CLI *vm create* command, there are two options, either to create a VM from a Harvester VM template or from a VM image
func vmCreate(ctx *cli.Context) error {
	c, err := GetHarvesterClient(ctx)

	if err != nil {
		return err
	}

	if ctx.String("template") != "" {
		return vmCreateFromTemplate(ctx, c)
	} else {
		return vmCreateFromImage(ctx, c, nil)
	}
}

//vmCreateFromTemplate creates a VM from a VM template provided in the CLI command
func vmCreateFromTemplate(ctx *cli.Context, c *harvclient.Clientset) error {
	template := ctx.String("template")

	// noFlagList := []string{"cpus", "memory", "disk", "vm-image-id", "ssk-keyname", "cloud-init-user-data", "cloud-init-network-data"}

	// for _, flag := range noFlagList {
	// 	fmt.Printf("Flag %s has a value %s", flag, ctx.String(flag))
	// 	if ctx.String(flag) != "" {
	// 		return fmt.Errorf("the flag %s was given when using template flag, this is not permitted", flag)
	// 	}
	// }

	logrus.Warnf("You are using a template flag, please be aware that any other flag will be IGNORED!")

	// checking template format
	subCompTemplate := SplitOnColon(template)

	if len(subCompTemplate) > 2 {
		return fmt.Errorf("given template flag does not have the format <template_name> or <template_name>:<version>")
	}

	templateName := subCompTemplate[0]
	var version int
	var err error
	if len(subCompTemplate) == 1 {
		version = 0
	} else if len(subCompTemplate) == 2 {
		version, err = strconv.Atoi(subCompTemplate[1])
	}

	if err != nil {
		return fmt.Errorf("version given in template flag %s is not an integer", subCompTemplate[1])
	}

	// checking if template exists
	templateContent, err := c.HarvesterhciV1beta1().VirtualMachineTemplates(ctx.String("namespace")).Get(context.TODO(), templateName, k8smetav1.GetOptions{})

	if err != nil {
		return fmt.Errorf("template %s was not found on the Harvester Cluster", subCompTemplate[0])
	}

	// Picking the templateVersion
	var templateVersion *v1beta1.VirtualMachineTemplateVersion
	if version == 0 {
		templateVersionString := strings.Split(templateContent.Spec.DefaultVersionID, "/")[1]
		templateVersionNamespace := strings.Split(templateContent.Spec.DefaultVersionID, "/")[0]
		logrus.Debugf("templateVersion found is :%s\n", templateContent.Spec.DefaultVersionID)

		templateVersion, err = c.HarvesterhciV1beta1().VirtualMachineTemplateVersions(templateVersionNamespace).Get(context.TODO(), templateVersionString, k8smetav1.GetOptions{})
		// templateVersion, err := c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineTemplates(templateVersionNamespace).Get(context.TODO(), "ubuntu-template", k8smetav1.GetOptions{})

		if err != nil {
			return err
		}
		templateVersion.ManagedFields = []k8smetav1.ManagedFieldsEntry{}
		marshalledTemplateVersion, err := json.Marshal(templateVersion)

		if err != nil {
			return err
		}
		logrus.Debugf("template version: %s\n", string(marshalledTemplateVersion))
	} else {
		templateVersion, err = fetchTemplateVersionFromInt(ctx, c, version, templateName)
		if err != nil {
			return err
		}
	}

	templateVersionAnnot := templateVersion.Spec.VM.ObjectMeta.Annotations[vmAnnotationPVC]
	logrus.Debugf("VM Annotation for PVC (should be JSON format): %s", templateVersionAnnot)
	var pvcList []v1.PersistentVolumeClaim
	err = json.Unmarshal([]byte(templateVersionAnnot), &pvcList)

	if err != nil {
		return err
	}

	pvc := pvcList[0]

	vmImageIdWithNamespace := pvc.ObjectMeta.Annotations["harvesterhci.io/imageId"]
	vmImageId := strings.Split(vmImageIdWithNamespace, "/")[1]

	err = ctx.Set("vm-image-id", vmImageId)

	if err != nil {
		return fmt.Errorf("error during setting flag to context: %w", err)
	}

	err = ctx.Set("disk-size", pvc.Spec.Resources.Requests.Storage().String())

	if err != nil {
		return fmt.Errorf("error during setting flag to context: %w", err)
	}

	vmTemplate := templateVersion.Spec.VM.Spec.Template

	err = vmCreateFromImage(ctx, c, vmTemplate)

	if err != nil {
		return err
	}

	return nil
}

// fetchTemplateVersionFromInt gets the Template with the right version given the context (containing template name) and the version as an integer
func fetchTemplateVersionFromInt(ctx *cli.Context, c *harvclient.Clientset, version int, templateName string) (*v1beta1.VirtualMachineTemplateVersion, error) {

	templateSelector := "template.harvesterhci.io/templateID=" + templateName

	allTemplateVersions, err := c.HarvesterhciV1beta1().VirtualMachineTemplateVersions(ctx.String("namespace")).List(context.TODO(), k8smetav1.ListOptions{
		LabelSelector: templateSelector,
	})

	if err != nil {
		return nil, err
	}

	for _, serverTemplateVersion := range allTemplateVersions.Items {
		if version == serverTemplateVersion.Status.Version {
			return &serverTemplateVersion, nil
		}

	}
	return nil, fmt.Errorf("no template with the same version found")
}

//vmCreateFromImage creates a VM from a VM Image using the CLI command context to get information
func vmCreateFromImage(ctx *cli.Context, c *harvclient.Clientset, vmTemplate *VMv1.VirtualMachineInstanceTemplateSpec) error {

	var err error
	// Checking existence of Image ID and if not, using default ubuntu image.
	imageID := ctx.String("vm-image-id")
	var vmImage *v1beta1.VirtualMachineImage
	if imageID != "" {
		vmImage, err = c.HarvesterhciV1beta1().VirtualMachineImages(ctx.String("namespace")).Get(context.TODO(), imageID, k8smetav1.GetOptions{})
		if err != nil {
			return err
		}
		logrus.Debugf("Image ID %s given does exist!", ctx.String("vm-image-id"))
	} else {
		vmImage, err = setDefaultVMImage(c, ctx)
		if err != nil {
			return err
		}
	}
	storageClassName := vmImage.Status.StorageClassName
	vmName := ctx.Args().First()
	diskRandomID := RandomID()

	pvcName := vmName + "-disk-0-" + diskRandomID
	pvcAnnotation := "[{\"metadata\":{\"name\":\"" + pvcName + "\",\"annotations\":{\"harvesterhci.io/imageId\":\"" + ctx.String("namespace") + "/" + ctx.String("vm-image-id") + "\"}},\"spec\":{\"accessModes\":[\"ReadWriteMany\"],\"resources\":{\"requests\":{\"storage\":\"" + ctx.String("disk-size") + "\"}},\"volumeMode\":\"Block\",\"storageClassName\":\"" + storageClassName + "\"}}]"
	vmLabels := map[string]string{
		"harvesterhci.io/creator": "harvester",
	}
	vmiLabels := vmLabels
	vmiLabels["harvesterhci.io/vmName"] = vmName

	if vmTemplate == nil {

		cloudInitUserData, err := getCloudInitData(ctx, "user")

		if err != nil {
			return err
		}

		var sshKey *v1beta1.KeyPair

		// Checking existence of the SSH KeyPair
		keyName := ctx.String("ssh-keyname")
		if keyName != "" {
			sshKey, err = c.HarvesterhciV1beta1().KeyPairs(ctx.String("namespace")).Get(context.TODO(), keyName, k8smetav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("error during getting keypair from Harvester: %w", err)
			}
			logrus.Debugf("SSH Key Name %s given does exist!", ctx.String("ssh-keyname"))

		} else {
			sshKey, err = setDefaultSSHKey(c, ctx)
			if err != nil {
				return fmt.Errorf("error during setting default SSH key: %w", err)
			}
		}

		if sshKey == nil || sshKey == (&v1beta1.KeyPair{}) {
			return fmt.Errorf("no keypair could be defined")
		}

		cloudInitSSHSection := "\nssh_authorized_keys:\n  - " + sshKey.Spec.PublicKey + "\n"

		cloudInitNetworkData, err := getCloudInitData(ctx, "network")
		if err != nil {
			return fmt.Errorf("error during getting cloud-init for networking: %w", err)
		}
		logrus.Debug("CloudInit: ")

		vmTemplate = &VMv1.VirtualMachineInstanceTemplateSpec{
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
							PersistentVolumeClaim: &VMv1.PersistentVolumeClaimVolumeSource{
								PersistentVolumeClaimVolumeSource: v1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvcName,
								},
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
		}
	} else {
		vmTemplate.Spec.Volumes[0].PersistentVolumeClaim.ClaimName = pvcName
	}

	ubuntuVM := &VMv1.VirtualMachine{
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      vmName,
			Namespace: ctx.String("namespace"),
			Annotations: map[string]string{

				vmAnnotationPVC:        pvcAnnotation,
				vmAnnotationNetworkIps: "[]",
			},
			Labels: vmLabels,
		},
		Spec: VMv1.VirtualMachineSpec{
			Running: NewTrue(),

			Template: vmTemplate,
		},
	}

	if err != nil {
		return err
	}

	_, err = c.KubevirtV1().VirtualMachines(ctx.String("namespace")).Create(context.TODO(), ubuntuVM, k8smetav1.CreateOptions{})

	if err != nil {
		return err
	}

	return nil
}

// vmStart issues a power on for the virtual machine instance.
func vmStart(ctx *cli.Context) error {

	c, err := GetHarvesterClient(ctx)
	if err != nil {
		return err
	}

	vm, err := c.KubevirtV1().VirtualMachines(ctx.String("namespace")).Get(context.TODO(), ctx.Args().First(), k8smetav1.GetOptions{})

	*vm.Spec.Running = true

	if err != nil {
		return err
	}

	_, err = c.KubevirtV1().VirtualMachines(ctx.String("namespace")).Update(context.TODO(), vm, k8smetav1.UpdateOptions{})
	return err
}

// Stop issues a power off for the virtual machine instance.
func vmStop(ctx *cli.Context) error {

	c, err := GetHarvesterClient(ctx)
	if err != nil {
		return err
	}

	vm, err := c.KubevirtV1().VirtualMachines(ctx.String("namespace")).Get(context.TODO(), ctx.Args().First(), k8smetav1.GetOptions{})
	*vm.Spec.Running = false

	if err != nil {
		return err
	}

	_, err = c.KubevirtV1().VirtualMachines(ctx.String("namespace")).Update(context.TODO(), vm, k8smetav1.UpdateOptions{})
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

// vmiAnnotations generates a map of strings to be injected as annotations from a PVC name and an SSK Keyname
func vmiAnnotations(pvcName string, sshKeyName string) map[string]string {
	return map[string]string{
		"harvesterhci.io/diskNames": "[\"" + pvcName + "\"]",
		"harvesterhci.io/sshNames":  "[\"" + sshKeyName + "\"]",
	}
}

// setDefaultVMImage creates a default VM image based on Ubuntu if none has been provided at the command line.
func setDefaultVMImage(c *harvclient.Clientset, ctx *cli.Context) (result *v1beta1.VirtualMachineImage, err error) {

	result = &v1beta1.VirtualMachineImage{}
	vmImages, err1 := c.HarvesterhciV1beta1().VirtualMachineImages(ctx.String("namespace")).List(context.TODO(), k8smetav1.ListOptions{})

	if err1 != nil {
		err = fmt.Errorf("error during setting default VM Image: %w", err1)
		return
	}

	var vmImage *v1beta1.VirtualMachineImage

	if len(vmImages.Items) == 0 {
		vmImage, err1 = CreateVMImage(c, ctx.String("namespace"), "ubuntu-default-image", ubuntuDefaultImage)
		if err1 != nil {
			err = fmt.Errorf("impossible to create a default VM Image: %w", err1)
			return
		}
	} else {
		vmImage = &vmImages.Items[0]
	}

	imageID := vmImage.ObjectMeta.Name
	err1 = ctx.Set("vm-image-id", imageID)

	if err1 != nil {
		logrus.Warnf("error encountered during the storage of the imageID value: %s", imageID)
	}

	result = vmImage

	return
}

// setDefaultSSHKey assign a default SSH key to the VM if none was provided at the command line
func setDefaultSSHKey(c *harvclient.Clientset, ctx *cli.Context) (sshKey *v1beta1.KeyPair, err error) {
	sshKey = &v1beta1.KeyPair{}
	sshKeys, err1 := c.HarvesterhciV1beta1().KeyPairs(ctx.String("namespace")).List(context.TODO(), k8smetav1.ListOptions{})

	if err1 != nil {
		err = fmt.Errorf("error during listing Keypairs: %w", err1)
		return
	}

	if len(sshKeys.Items) == 0 {
		err = fmt.Errorf("no ssh keys exists in harvester, please add a new ssh key")
		return
	}

	sshKey = &sshKeys.Items[0]
	err = ctx.Set("ssh-keyname", sshKey.Name)

	if err != nil {
		logrus.Warnf("Error encountered during the storage of the SSH Keyname value: %s", sshKey.Name)
	}
	return
}

// getCloudInitNetworkData gives the ConfigMap object with name indicated in the command line,
// and will create a new one called "ubuntu-std-network" if none is provided or no ConfigMap was found with the same name
func getCloudInitData(ctx *cli.Context, scope string) (*v1.ConfigMap, error) {
	var cmName string
	c, err := GetKubeClient(ctx)

	if err != nil {
		return &v1.ConfigMap{}, err
	}
	if scope != "user" && scope != "network" {
		return nil, fmt.Errorf("wrong value for scope parameter")
	}

	flagName := "cloud-init-" + scope + "-data"

	if ctx.String(flagName) == "" {
		cmName = defaultCloudInitCmPrefix + scope + "-data"
	} else {
		cmName = ctx.String(flagName)
	}
	var ciData *v1.ConfigMap
	//var err error
	ciData, err = c.CoreV1().ConfigMaps(ctx.String("namespace")).Get(context.TODO(), cmName, k8smetav1.GetOptions{})

	if err != nil && cmName == ctx.String(flagName) {
		return nil, fmt.Errorf("%[1]v config map was not found, please specify another configmap or remove the %[1]v flag to use the default one for ubuntu", cmName)
	}

	var cloudInitContent string
	if scope == "user" {
		cloudInitContent = defaultCloudInitUserData
	} else if scope == "network" {
		cloudInitContent = defaultCloudInitNetworkData
	}

	if err != nil {
		var err1 error
		ciData, err1 = c.CoreV1().ConfigMaps(ctx.String("namespace")).Create(context.TODO(), &v1.ConfigMap{
			ObjectMeta: k8smetav1.ObjectMeta{
				Name: cmName,
			},
			Data: map[string]string{
				"cloudInit": cloudInitContent,
			},
		}, k8smetav1.CreateOptions{})

		if err1 != nil {
			fmt.Println("Error Creating CM: " + err1.Error())
			return nil, fmt.Errorf("error during creation of default cloud-init template")
		}
	}

	return ciData, nil
}

// CreateVMImage will create a VM Image on Harvester given an image name and an image URL
func CreateVMImage(c *harvclient.Clientset, namespace string, imageName string, url string) (*v1beta1.VirtualMachineImage, error) {

	vmImage, err := c.HarvesterhciV1beta1().VirtualMachineImages(namespace).Create(
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
