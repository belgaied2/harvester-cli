package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	harvclient "github.com/harvester/harvester/pkg/generated/clientset/versioned"
	"github.com/harvester/harvester/pkg/util"
	"github.com/minio/pkg/wildcard"
	rcmd "github.com/rancher/cli/cmd"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	VMv1 "kubevirt.io/api/core/v1"
)

const (
	vmAnnotationPVC              = "harvesterhci.io/volumeClaimTemplates"
	vmAnnotationNetworkIps       = "networks.harvesterhci.io/ips"
	defaultDiskSize              = "10Gi"
	defaultMemSize               = "1Gi"
	defaultNbCPUCores            = 1
	defaultNamespace             = "default"
	ubuntuDefaultImage           = "https://cloud-images.ubuntu.com/minimal/daily/focal/current/focal-minimal-cloudimg-amd64.img"
	defaultCloudInitUserData     = "#cloud-config\npassword: password\nchpasswd: { expire: False}\nssh_pwauth: True\npackages:\n  - qemu-guest-agent\nruncmd:\n  - [ systemctl, daemon-reload ]\n  - [ systemctl, enable, qemu-guest-agent.service ]\n  - [ systemctl, start, --no-block, qemu-guest-agent.service ]"
	defaultCloudInitNetworkData  = "version: 2\nrenderer: networkd\nethernets:\n  enp1s0:\n    dhcp4: true"
	defaultCloudInitCmPrefix     = "default-ubuntu-"
	defaultOverCommitSettingName = "overcommit-config"
)

var (
	nsFlag = cli.StringFlag{
		Name:    "namespace, n",
		Usage:   "Namespace of the VM",
		EnvVars: []string{"HARVESTER_VM_NAMESPACE"},
		Value:   defaultNamespace,
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
func VMCommand() *cli.Command {
	return &cli.Command{
		Name:    "virtualmachine",
		Aliases: []string{"vm"},
		Usage:   "Manage Virtual Machines on Harvester",
		Action:  defaultAction(vmLs),
		Subcommands: []*cli.Command{
			{
				Name:        "list",
				Usage:       "List VMs",
				Aliases:     []string{"ls"},
				Description: "\nList all VMs in the current Harvester Cluster",
				ArgsUsage:   "None",
				Action:      vmLs,
				Flags: []cli.Flag{
					&nsFlag,
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
					&nsFlag,
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
					&nsFlag,
					&cli.StringFlag{
						Name:    "vm-description",
						Usage:   "Optional description of your VM",
						EnvVars: []string{"HARVESTER_VM_DESCRIPTION"},
						Value:   "",
					},
					&cli.StringFlag{
						Name:    "vm-image-id",
						Usage:   "Harvester Image ID of the VM to create",
						EnvVars: []string{"HARVESTER_VM_IMAGE_ID"},
						Value:   "",
					},
					&cli.StringFlag{
						Name:    "disk-size",
						Aliases: []string{"disk", "d"},
						Usage:   "Size of the primary VM disk",
						EnvVars: []string{"HARVESTER_VM_DISKSIZE"},
						Value:   defaultDiskSize,
					},
					&cli.StringFlag{
						Name:    "ssh-keyname",
						Aliases: []string{"i"},
						Usage:   "KeyName of the SSH Key to use with this VM",
						EnvVars: []string{"HARVESTER_VM_KEY"},
						Value:   "",
					},
					&cli.IntFlag{
						Name:    "cpus",
						Aliases: []string{"c"},
						Usage:   "Number of CPUs to dedicate to the VM",
						EnvVars: []string{"HARVESTER_VM_CPUS"},
						Value:   defaultNbCPUCores,
					},
					&cli.StringFlag{
						Name:    "memory",
						Aliases: []string{"m"},
						Usage:   "Amount of memory in the format XXGi",
						EnvVars: []string{"HARVESTER_VM_MEMORY"},
						Value:   defaultMemSize,
					},
					&cli.StringFlag{
						Name:    "user-data-cm-ref",
						Aliases: []string{"user-data-cm"},
						Usage:   "Name of the Cloud Init User Data Template to be used (already in Harvester)",
						EnvVars: []string{"HARVESTER_USER_DATA_CM_REF"},
						Value:   "",
					},
					&cli.StringFlag{
						Name:    "network-data-cm-ref",
						Aliases: []string{"network-data-cm"},
						Usage:   "Name of the Cloud Init Network Data Template to be used (already in Harvester)",
						EnvVars: []string{"HARVESTER_NETWORK_DATA_CM_REF"},
						Value:   "",
					},
					&cli.StringFlag{
						Name:    "user-data-filepath",
						Aliases: []string{"user-data-file"},
						Usage:   "Path to a valid cloud-init YAML file to be used with VM creation",
						EnvVars: []string{"HARVESTER_USER_DATA_FILEPATH"},
						Value:   "",
					},
					&cli.StringFlag{
						Name:    "network-data-filepath",
						Aliases: []string{"network-data-file"},
						Usage:   "Path to a valid cloud-init YAML file to be used with VM creation",
						EnvVars: []string{"HARVESTER_NETWORK_DATA_FILEPATH"},
						Value:   "",
					},
					&cli.StringFlag{
						Name:    "template",
						Aliases: []string{"from-template"},
						Usage:   "Harvester VM Template to use for creating the VM in the format <template_name>:<version> or <template> in which case the latest version will be used",
						EnvVars: []string{"HARVESTER_VM_TEMPLATE"},
						Value:   "",
					},
					&cli.IntFlag{
						Name:    "count",
						Aliases: []string{"number"},
						Usage:   "Number of identical VMs to create",
						EnvVars: []string{"HARVESTER_VM_COUNT"},
						Value:   1,
					},
					&cli.StringFlag{
						Name:    "network",
						Aliases: []string{"net"},
						Usage:   "Network to which the VM should be belong",
						EnvVars: []string{"HARVESTER_VM_NETWORK"},
						Value:   "vlan1",
					},
				},
			},
			{
				Name:      "stop",
				Usage:     "Stop a VM",
				Action:    vmStop,
				ArgsUsage: "[VM_NAME]",
				Flags: []cli.Flag{
					&nsFlag,
				},
			},
			{
				Name:      "start",
				Usage:     "Start a VM",
				Action:    vmStart,
				ArgsUsage: "[VM_NAME]",
				Flags: []cli.Flag{
					&nsFlag,
				},
			},
			{
				Name:      "restart",
				Usage:     "Restart a VM",
				Action:    vmRestart,
				ArgsUsage: "[VM_NAME]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "vm-name, name",
						Usage: "Name of the VM to restart",
					},
					&nsFlag,
				},
			},
		},
	}
}

// vmLs lists the VMs available in Harvester
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
		ctxv1)

	defer writer.Close()

	for _, vm := range vmList.Items {

		state := string(vm.Status.PrintableStatus)

		var IP string
		if vmiMap[vm.Name].Status.Interfaces == nil {
			IP = ""
		} else {
			IP = vmiMap[vm.Name].Status.Interfaces[0].IP
		}

		var memory string
		if vm.Spec.Template.Spec.Domain.Resources.Limits.Memory().CmpInt64(int64(0)) == 0 {
			memory = vm.Spec.Template.Spec.Domain.Resources.Requests.Memory().String()
		} else {
			memory = vm.Spec.Template.Spec.Domain.Resources.Limits.Memory().String()
		}

		writer.Write(&VirtualMachineData{
			State:          state,
			VirtualMachine: vm,
			Name:           vm.Name,
			Node:           vmiMap[vm.Name].Status.NodeName,
			CPU:            vm.Spec.Template.Spec.Domain.CPU.Cores,
			Memory:         memory,
			IPAddress:      IP,
		})

	}

	return writer.Err()
}

// vmDelete deletes VMs which name is given in argument
func vmDelete(ctx *cli.Context) error {
	c, err := GetHarvesterClient(ctx)

	if err != nil {
		return err
	}

	for _, vmName := range ctx.Args().Slice() {

		if strings.Contains(vmName, "*") || strings.Contains(vmName, "?") {
			matchingVMs := buildVMListMatchingWildcard(c, ctx, vmName)

			for _, vmExisting := range matchingVMs {

				err := vmDeleteWithPVC(&vmExisting, c, ctx)
				if err != nil {
					return err
				}
			}
		} else {
			vm, err := c.KubevirtV1().VirtualMachines(ctx.String("namespace")).Get(context.TODO(), vmName, k8smetav1.GetOptions{})

			if err != nil {
				return fmt.Errorf("no VM with the provided name found")
			}

			err = vmDeleteWithPVC(vm, c, ctx)
			if err != nil {
				return err
			}
		}
	}

	return nil

}

func vmDeleteWithPVC(vmExisting *VMv1.VirtualMachine, c *harvclient.Clientset, ctx *cli.Context) error {

	vmCopy := vmExisting.DeepCopy()
	var removedPVCs []string
	if vmCopy.Spec.Template != nil {
		for _, vol := range vmCopy.Spec.Template.Spec.Volumes {
			if vol.PersistentVolumeClaim == nil {
				continue
			}

			removedPVCs = append(removedPVCs, vol.PersistentVolumeClaim.ClaimName)

		}
	}

	vmCopy.Annotations[util.RemovedPVCsAnnotationKey] = strings.Join(removedPVCs, ",")
	_, err := c.KubevirtV1().VirtualMachines(ctx.String("namespace")).Update(context.TODO(), vmCopy, k8smetav1.UpdateOptions{})

	if err != nil {
		return fmt.Errorf("error during removal of PVCs in the VM reference, %w", err)
	}

	err = c.KubevirtV1().VirtualMachines(ctx.String("namespace")).Delete(context.TODO(), vmCopy.Name, k8smetav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("VM named %s could not be deleted successfully: %w", vmCopy.Name, err)
	} else {
		logrus.Infof("VM %s deleted successfully", vmCopy.Name)
	}
	return nil
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

// vmCreateFromTemplate creates a VM from a VM template provided in the CLI command
func vmCreateFromTemplate(ctx *cli.Context, c *harvclient.Clientset) error {
	template := ctx.String("template")

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
		templateVersion, err = fetchTemplateVersionFromInt(ctx.String("namespace"), c, version, templateName)
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
func fetchTemplateVersionFromInt(namespace string, c *harvclient.Clientset, version int, templateName string) (*v1beta1.VirtualMachineTemplateVersion, error) {

	templateSelector := "template.harvesterhci.io/templateID=" + templateName

	allTemplateVersions, err := c.HarvesterhciV1beta1().VirtualMachineTemplateVersions(namespace).List(context.TODO(), k8smetav1.ListOptions{
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

// vmCreateFromImage creates a VM from a VM Image using the CLI command context to get information
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
	vmNameBase := ctx.Args().First()

	vmLabels := map[string]string{
		"harvesterhci.io/creator": "harvester",
	}
	vmiLabels := vmLabels

	if ctx.Int("count") == 0 {
		return fmt.Errorf("VM count provided is 0, no VM will be created")
	}

	// Checking if provided Network exists in Harvester
	_, err = c.K8sCniCncfIoV1().NetworkAttachmentDefinitions(ctx.String("namespace")).Get(context.TODO(), ctx.String("network"), k8smetav1.GetOptions{})

	if err != nil {
		return fmt.Errorf("problem while verifying network existence; %w", err)
	}

	for i := 1; i <= ctx.Int("count"); i++ {
		var vmName string
		if ctx.Int("count") > 1 {
			vmName = vmNameBase + "-" + fmt.Sprint(i)
		} else {
			vmName = vmNameBase
		}

		vmiLabels["harvesterhci.io/vmName"] = vmName
		vmiLabels["harvesterhci.io/vmNamePrefix"] = vmNameBase
		diskRandomID := RandomID()
		pvcName := vmName + "-disk-0-" + diskRandomID
		pvcAnnotation := "[{\"metadata\":{\"name\":\"" + pvcName + "\",\"annotations\":{\"harvesterhci.io/imageId\":\"" + ctx.String("namespace") + "/" + ctx.String("vm-image-id") + "\"}},\"spec\":{\"accessModes\":[\"ReadWriteMany\"],\"resources\":{\"requests\":{\"storage\":\"" + ctx.String("disk-size") + "\"}},\"volumeMode\":\"Block\",\"storageClassName\":\"" + storageClassName + "\"}}]"

		if vmTemplate == nil {

			vmTemplate, err = buildVMTemplate(ctx, c, pvcName, vmiLabels, vmNameBase)
			if err != nil {
				return err
			}
		} else {
			vmTemplate.Spec.Volumes[0].PersistentVolumeClaim.ClaimName = pvcName

			if vmTemplate.ObjectMeta.Labels == nil {
				vmTemplate.ObjectMeta.Labels = make(map[string]string)
			}

			vmTemplate.ObjectMeta.Labels["harvesterhci.io/vmNamePrefix"] = vmNameBase
			vmTemplate.Spec.Affinity = &v1.Affinity{
				PodAntiAffinity: &v1.PodAntiAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
						{
							Weight: int32(1),
							PodAffinityTerm: v1.PodAffinityTerm{
								TopologyKey: "kubernetes.io/hostname",
								LabelSelector: &k8smetav1.LabelSelector{
									MatchLabels: map[string]string{
										"harvesterhci.io/vmNamePrefix": vmNameBase,
									},
								},
							},
						},
					},
				},
			}
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
	}

	return nil
}

// buildVMTemplate creates a *VMv1.VirtualMachineInstanceTemplateSpec from the CLI Flags and some computed values
func buildVMTemplate(ctx *cli.Context, c *harvclient.Clientset,
	pvcName string, vmiLabels map[string]string, vmName string) (vmTemplate *VMv1.VirtualMachineInstanceTemplateSpec, err error) {

	var err1 error
	cloudInitUserData, err1 := getCloudInitData(ctx, "user")
	vmTemplate = nil
	if err1 != nil {
		err = fmt.Errorf("error during getting cloud init user data from Harvester: %w", err1)
		return
	}

	var sshKey *v1beta1.KeyPair

	keyName := ctx.String("ssh-keyname")
	if keyName != "" {
		sshKey, err1 = c.HarvesterhciV1beta1().KeyPairs(ctx.String("namespace")).Get(context.TODO(), keyName, k8smetav1.GetOptions{})
		if err1 != nil {
			err = fmt.Errorf("error during getting keypair from Harvester: %w", err1)
			return
		}
		logrus.Debugf("SSH Key Name %s given does exist!", ctx.String("ssh-keyname"))

	} else {
		sshKey, err1 = setDefaultSSHKey(c, ctx)
		if err1 != nil {
			err = fmt.Errorf("error during setting default SSH key: %w", err1)
			return
		}
	}

	if sshKey == nil || sshKey == (&v1beta1.KeyPair{}) {
		err = fmt.Errorf("no keypair could be defined")
		return
	}

	cloudInitSSHSection := "\nssh_authorized_keys:\n  - " + sshKey.Spec.PublicKey + "\n"

	cloudInitNetworkData, err1 := getCloudInitData(ctx, "network")
	if err1 != nil {
		err = fmt.Errorf("error during getting cloud-init for networking: %w", err1)
		return
	}
	logrus.Debug("CloudInit: ")

	overCommitSetting, err := c.HarvesterhciV1beta1().Settings().Get(context.TODO(), defaultOverCommitSettingName, k8smetav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("encountered issue when querying Harvester for setting %s: %w", defaultOverCommitSettingName, err)
	}

	var overCommitSettingMap map[string]int
	err = json.Unmarshal([]byte(overCommitSetting.Default), &overCommitSettingMap)
	if err != nil {
		return nil, fmt.Errorf("encountered issue when unmarshalling setting value %s: %w", defaultOverCommitSettingName, err)
	}

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
							NetworkName: ctx.String("network"),
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
							UserData:    cloudInitUserData + cloudInitSSHSection,
							NetworkData: cloudInitNetworkData,
						},
					},
				},
			},
			Domain: VMv1.DomainSpec{
				CPU: &VMv1.CPU{
					Cores:   uint32(ctx.Int("cpus")),
					Sockets: 1,
					Threads: 1,
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
						"memory": HandleMemoryOverCommittment(overCommitSettingMap, ctx.String("memory")),
						"cpu":    HandleCPUOverCommittment(overCommitSettingMap, int64(ctx.Int("cpus"))),
					},
					Limits: v1.ResourceList{
						"memory": resource.MustParse(ctx.String("memory")),
						"cpu":    *resource.NewQuantity(int64(ctx.Int("cpus")), resource.DecimalSI),
					},
				},
			},
			Affinity: &v1.Affinity{
				PodAntiAffinity: &v1.PodAntiAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
						{
							Weight: int32(1),
							PodAffinityTerm: v1.PodAffinityTerm{
								TopologyKey: "kubernetes.io/hostname",
								LabelSelector: &k8smetav1.LabelSelector{
									MatchLabels: map[string]string{
										"harvesterhci.io/vmNamePrefix": vmName,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return
}

// vmStart issues a power on for the virtual machine instances which names are given as argument to the start command.
func vmStart(ctx *cli.Context) error {

	c, err := GetHarvesterClient(ctx)
	if err != nil {
		return err
	}

	for _, vmName := range ctx.Args().Slice() {

		if strings.Contains(vmName, "*") || strings.Contains(vmName, "?") {
			matchingVMs := buildVMListMatchingWildcard(c, ctx, vmName)

			for _, vmNameExisting := range matchingVMs {
				err = startVMbyRef(c, ctx, vmNameExisting)
				if err != nil {
					return err
				}
			}
		} else {
			return startVMbyName(c, ctx, vmName)

		}
	}

	return nil
}

// buildVMListMatchingWildcard creates an array of VM objects which names match the given wildcard pattern
func buildVMListMatchingWildcard(c *harvclient.Clientset, ctx *cli.Context, vmNameWildcard string) []VMv1.VirtualMachine {
	vms, err := c.KubevirtV1().VirtualMachines(ctx.String("namespace")).List(context.TODO(), k8smetav1.ListOptions{})

	if err != nil {
		logrus.Warnf("No VMs found with name %s", vmNameWildcard)
	}

	var matchingVMs []VMv1.VirtualMachine
	for _, vm := range vms.Items {
		// logrus.Warnf("current VM checked: %s", vm.Name)
		if wildcard.Match(vmNameWildcard, vm.Name) {
			matchingVMs = append(matchingVMs, vm)
			// logrus.Warnf("VM %s appended to list of matching VMs", vm.Name)
		}
	}
	logrus.Infof("number of matching VMs for pattern %s: %d", vmNameWildcard, len(matchingVMs))
	return matchingVMs
}

// startVMbyName starts a VM by first issuing a GET using the VM name, then updating the resulting VM object
func startVMbyName(c *harvclient.Clientset, ctx *cli.Context, vmName string) error {
	vm, err := c.KubevirtV1().VirtualMachines(ctx.String("namespace")).Get(context.TODO(), vmName, k8smetav1.GetOptions{})

	if err != nil {
		err1 := fmt.Errorf("vm with provided name not found: %w", err)
		logrus.Errorf("No VM named %s was not found (%s) the subsequent VMs will not be started!", vmName, err)
		return err1
	}

	return startVMbyRef(c, ctx, *vm)
}

// startVMbyRef updates a VM object to make it Running
func startVMbyRef(c *harvclient.Clientset, ctx *cli.Context, vm VMv1.VirtualMachine) (err error) {

	*vm.Spec.Running = true

	_, err = c.KubevirtV1().VirtualMachines(ctx.String("namespace")).Update(context.TODO(), &vm, k8smetav1.UpdateOptions{})

	if err != nil {
		logrus.Warnf("An error happened while starting VM %s: %s", vm.Name, err)
	} else {
		logrus.Infof("VM %s started successfully", vm.Name)
	}
	return nil
}

// vmStop issues a power off for the virtual machine instances which name is given as argument.
func vmStop(ctx *cli.Context) error {

	c, err := GetHarvesterClient(ctx)
	if err != nil {
		return err
	}
	for _, vmName := range ctx.Args().Slice() {

		if strings.Contains(vmName, "*") || strings.Contains(vmName, "?") {
			matchingVMs := buildVMListMatchingWildcard(c, ctx, vmName)

			for _, vmExisting := range matchingVMs {
				err = stopVMbyRef(c, ctx, &vmExisting)
				if err != nil {
					return err
				}
			}
		} else {
			return stopVMbyName(c, ctx, vmName)
		}
	}
	return err
}

// stopVMbyName will stop a VM by first finding it by its name and then call stopBMbyRef function
func stopVMbyName(c *harvclient.Clientset, ctx *cli.Context, vmName string) error {
	vm, err := c.KubevirtV1().VirtualMachines(ctx.String("namespace")).Get(context.TODO(), vmName, k8smetav1.GetOptions{})

	if err != nil {
		err1 := fmt.Errorf("vm with provided name not found: %s", err)
		logrus.Errorf("No VM named %s was not found (%s) the subsequent VMs will not be stopped!", vmName, err)
		return err1
	}

	return stopVMbyRef(c, ctx, vm)
}

// stopVMbyRef will stop a VM by updating Spec.Running field of the VM object
func stopVMbyRef(c *harvclient.Clientset, ctx *cli.Context, vm *VMv1.VirtualMachine) error {
	*vm.Spec.Running = false

	_, err := c.KubevirtV1().VirtualMachines(ctx.String("namespace")).Update(context.TODO(), vm, k8smetav1.UpdateOptions{})
	if err != nil {
		logrus.Warnf("An error happened while stopping VM %s: %s", vm.Name, err)
	} else {
		logrus.Infof("VM %s stopped successfully", vm.Name)
	}
	return nil
}

// Restart reboots virtual machine instances by calling successively vmStop and vmStart
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
			err = fmt.Errorf("impossible to create a default VM Image: %s", err1)
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
		err = fmt.Errorf("error during listing Keypairs: %s", err1)
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
func getCloudInitData(ctx *cli.Context, scope string) (string, error) {
	var cmName string
	c, err := GetKubeClient(ctx)

	if err != nil {
		return "", err
	}
	if scope != "user" && scope != "network" {
		return "", fmt.Errorf("wrong value for scope parameter")
	}

	flagName := scope + "-data"
	var cloudInitDataString string

	if ctx.String(flagName+"-filepath") == "" {
		flagName = flagName + "-cm-ref"

		cmName = ctx.String(flagName)

		var ciData *v1.ConfigMap
		if cmName != "" {
			ciData, err = c.CoreV1().ConfigMaps(ctx.String("namespace")).Get(context.TODO(), cmName, k8smetav1.GetOptions{})

			if err != nil {
				return "", fmt.Errorf("%[1]v config map was not found, please specify another configmap or remove the %[1]v flag to use the default one for ubuntu", cmName)
			}

			return ciData.Data["cloudInit"], nil
		}

		if scope == "user" {
			return defaultCloudInitUserData, nil
		} else if scope == "network" {
			return defaultCloudInitNetworkData, nil
		}

	}
	if ctx.String(flagName+"-cm-ref") != "" {
		return "", fmt.Errorf("you can't specify both a configmap reference and a file path for the cloud-init data")
	}

	var cloudInitDataBytes []byte
	if cloudInitDataBytes, err = ioutil.ReadFile(ctx.String(flagName + "-filepath")); err != nil {
		return "", fmt.Errorf("error during reading of cloud-init file: %s", err)
	}
	cloudInitDataString = string(cloudInitDataBytes)

	return cloudInitDataString, nil
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
