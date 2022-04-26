package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	harvclient "github.com/harvester/harvester/pkg/generated/clientset/versioned"
	rcmd "github.com/rancher/cli/cmd"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	templateIDLabel   string = "template.harvesterhci.io/templateID"
	templateClaimTemp string = "harvesterhci.io/volumeClaimTemplates"
	imageIDAnnot      string = "harvesterhci.io/imageId"
	sshkeyAnnotation  string = "harvesterhci.io/sshNames"
)

type TemplateData struct {
	Name       string
	Version    int
	Image      string
	Cpus       uint32
	Memory     string
	Interfaces []Interface
	Keypairs   []string
	Volumes    []Volume
}

type Volume struct {
	Name                  string
	Type                  string
	PersistentVolumeClaim PersistentVolumeClaimObject `yaml:"persistentVolumeClaim,omitempty"`
	CloudInitData         CloudInitObject             `yaml:"cloudInit,omitempty"`
}

type PersistentVolumeClaimObject struct {
	ClaimName string
	Size      string
}

type CloudInitObject struct {
	Name        string
	NetworkData string
	UserData    string
}

type Interface struct {
	Name        string
	Type        string
	NetworkType string
	NetworkName string
}

// TemplateCommand defines the CLI command that lists VM templates in Harvester
func TemplateCommand() cli.Command {
	return cli.Command{
		Name:    "template",
		Aliases: []string{"tpl"},
		Usage:   "Manipulate VM templates",
		Action:  templateList,
		Flags: []cli.Flag{
			nsFlag,
		},
		Subcommands: cli.Commands{
			cli.Command{
				Name:        "list",
				Aliases:     []string{"ls"},
				Usage:       "List templates",
				Description: "\nLists all the VM templates available in Harvester",
				ArgsUsage:   "None",
				Action:      templateList,
				Flags: []cli.Flag{
					nsFlag,
				},
			},
			cli.Command{
				Name:        "show",
				Aliases:     []string{"get"},
				Usage:       "show the content of a VM template",
				Description: "\nshows information about the VM template given as an argument",
				ArgsUsage:   "VM_TEMPLATE:VERSION",
				Action:      templateShow,
				Flags: []cli.Flag{
					nsFlag,
				},
			},
		},
	}
}

// templateList implements the subcommand `template list`
func templateList(ctx *cli.Context) (err error) {
	c, err := GetHarvesterClient(ctx)

	if err != nil {
		return
	}

	tplList, err := c.HarvesterhciV1beta1().VirtualMachineTemplates(ctx.String("namespace")).List(context.TODO(), k8smetav1.ListOptions{})

	if err != nil {
		return
	}

	writer := rcmd.NewTableWriter([][]string{
		{"NAME", "Name"},
		{"LATEST_VERSION", "LatestVersion"},
	},
		ctx)

	defer writer.Close()

	for _, tplItem := range tplList.Items {

		writer.Write(&TemplateData{
			Name:    tplItem.Name,
			Version: tplItem.Status.LatestVersion,
		})

	}

	return writer.Err()
}

//templateShow prints the content of the VM template in argument given the CLI context
// It checks that the number of arguments provided is equal to one then queries the VirtualMachineTemplateVersion to print its content in YAML format.
func templateShow(ctx *cli.Context) error {

	if len(ctx.Args()) != 1 {
		return fmt.Errorf("wrong number of arguments, one and only one argument is accepted by this method")
	}

	vmTemplateArg := ctx.Args().First()

	pattern, err := regexp.Compile(`[a-zA-Z0-9\-]*:[0-9]?`)

	if err != nil {
		return err
	}

	if !pattern.MatchString(vmTemplateArg) {
		return fmt.Errorf("the argument provide does not have the right format, please give a VM template with a version in the format <VM_TEMPLATE_NAME>:<VERSION>")
	}

	c, err := GetHarvesterClient(ctx)

	if err != nil {
		return err
	}

	// Getting the template name and the version from the argument
	expArray := SplitOnColon(vmTemplateArg)
	vmTemplateName := expArray[0]
	vmTemplateVersion, err := strconv.Atoi(expArray[1])

	if err != nil {
		return fmt.Errorf("failed to convert version to integer, %w", err)
	}

	vmTemplateNameWithNS := vmTemplateName
	matchingVMTemplate, err := fetchTemplateVersionFromInt(ctx.String("namespace"), c, vmTemplateVersion, vmTemplateNameWithNS)

	if err != nil {
		return fmt.Errorf("error during querying VM Template, %w", err)
	}

	imageName, err := getImageName(matchingVMTemplate, c)

	if err != nil {
		return err
	}

	var toShowTemplate TemplateData
	toShowTemplate.Name = matchingVMTemplate.Labels[templateIDLabel]
	toShowTemplate.Version = matchingVMTemplate.Status.Version
	toShowTemplate.Cpus = matchingVMTemplate.Spec.VM.Spec.Template.Spec.Domain.CPU.Cores
	toShowTemplate.Memory = matchingVMTemplate.Spec.VM.Spec.Template.Spec.Domain.Resources.Limits.Memory().String()
	toShowTemplate.Volumes, err = mapVolumeData(ctx, matchingVMTemplate)

	if err != nil {
		return err
	}
	var keypairsFromTemplate []string
	err = json.Unmarshal(([]byte)(matchingVMTemplate.Spec.VM.Spec.Template.ObjectMeta.Annotations[sshkeyAnnotation]), &keypairsFromTemplate)

	toShowTemplate.Keypairs = keypairsFromTemplate

	if err != nil {
		return err
	}

	toShowTemplate.Interfaces = mapInterfaceData(matchingVMTemplate)
	toShowTemplate.Image = imageName

	templateYAMLbytes, err := yaml.Marshal(&toShowTemplate)

	if err != nil {
		return fmt.Errorf("failed during encoding an object to YAML: %w", err)
	}

	var templateYAMLstring string = string(templateYAMLbytes)

	fmt.Println(templateYAMLstring)

	return nil

}

// mapVolumeData returns an array of Volume objects that need to be added to the VirtualMachineInstanceTemplate when creating the VM object
func mapVolumeData(ctx *cli.Context, matchingVMTemplate *v1beta1.VirtualMachineTemplateVersion) (volumes []Volume, err error) {

	for _, origVolume := range matchingVMTemplate.Spec.VM.Spec.Template.Spec.Volumes {
		if origVolume.VolumeSource.PersistentVolumeClaim != nil {

			size, err := getPvcSizeFromMatchingAnnotation(origVolume.PersistentVolumeClaim.ClaimName, matchingVMTemplate)

			if err != nil {
				return []Volume{}, err
			}

			volumes = append(volumes, Volume{
				Name: origVolume.Name,
				Type: "persistentVolumeClaim",
				PersistentVolumeClaim: PersistentVolumeClaimObject{
					ClaimName: origVolume.PersistentVolumeClaim.ClaimName,
					Size:      size,
				},
			})
		}
		if origVolume.VolumeSource.CloudInitNoCloud != nil {

			networkData, err := getCloudInitDataFromSecret(ctx, origVolume.CloudInitNoCloud.UserDataSecretRef.Name, matchingVMTemplate.Namespace, "networkdata")

			if err != nil {
				return []Volume{}, err
			}

			userData, err := getCloudInitDataFromSecret(ctx, origVolume.CloudInitNoCloud.UserDataSecretRef.Name, matchingVMTemplate.Namespace, "userdata")

			if err != nil {
				return []Volume{}, err
			}

			volumes = append(volumes, Volume{
				Name: origVolume.Name,
				Type: "cloudInit",
				CloudInitData: CloudInitObject{
					Name:        origVolume.CloudInitNoCloud.UserDataSecretRef.Name,
					NetworkData: networkData,
					UserData:    userData,
				},
			})
		}
	}

	return
}

// getCloudInitDataFromSecret will query a secret to read Cloud Init Data to include in the VM Spec
func getCloudInitDataFromSecret(ctx *cli.Context, secretName, namespace, dataType string) (data string, err error) {
	c, err := GetKubeClient(ctx)

	if err != nil {
		return

	}

	cloudInitSecretContent, err1 := c.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, k8smetav1.GetOptions{})

	if err1 != nil {
		err = fmt.Errorf("error during getting cloud-init secret: %w", err1)
		return
	}
	data = string(cloudInitSecretContent.Data[dataType])

	return

}

// getPvcSizeFromMatchingAnnotation finds out the size of PVC in the template annotations, this is necessary to create a VM with the right volume size
func getPvcSizeFromMatchingAnnotation(claimName string, matchingVMTemplate *v1beta1.VirtualMachineTemplateVersion) (size string, err error) {
	claims, err := getPvcFromAnnotation(matchingVMTemplate)

	if err != nil {
		return
	}

	var matchingClaimFromAnnotation v1.PersistentVolumeClaim

	for _, claim := range claims {
		if claim.Name == claimName {
			matchingClaimFromAnnotation = claim
		}
	}

	size = matchingClaimFromAnnotation.Spec.Resources.Requests.Storage().String()

	return
}

// getImageName extracts the VM image name from the template
func getImageName(matchingVMTemplate *v1beta1.VirtualMachineTemplateVersion, c *harvclient.Clientset) (image string, err error) {
	claimObjectList, err1 := getPvcFromAnnotation(matchingVMTemplate)
	image = ""

	if err1 != nil {
		err = fmt.Errorf("error during unmarshalling an annotation, %w", err1)
		return
	}

	var imageIDFull string
	for _, claimObject := range claimObjectList {
		if claimObject.Annotations[imageIDAnnot] != "" {
			imageIDFull = claimObject.ObjectMeta.Annotations[imageIDAnnot]
		}
	}
	if imageIDFull == "" {
		err = fmt.Errorf("no ImageID found in template")
		return
	}

	imageNS := strings.Split(imageIDFull, "/")[0]
	imageID := strings.Split(imageIDFull, "/")[1]

	imageObject, err1 := c.HarvesterhciV1beta1().VirtualMachineImages(imageNS).Get(context.TODO(), imageID, k8smetav1.GetOptions{})

	if err1 != nil {
		err = fmt.Errorf("error during getting image object, %w", err1)
		return
	}

	image = imageObject.Spec.DisplayName
	return
}

// getPvcFromAnnotation finds out the PVC data in the template that will be used within the VM Spec.
func getPvcFromAnnotation(matchingVMTemplate *v1beta1.VirtualMachineTemplateVersion) ([]v1.PersistentVolumeClaim, error) {
	claimAnnot := matchingVMTemplate.Spec.VM.ObjectMeta.Annotations[templateClaimTemp]

	// fmt.Printf("Annotation:\n%s", claimAnnot)
	claimObjectList := []v1.PersistentVolumeClaim{}

	err1 := json.Unmarshal([]byte(claimAnnot), &claimObjectList)
	return claimObjectList, err1
}

// mapInterfaceData extracts an array of Network Interfaces from the template
func mapInterfaceData(vmTemplateVersion *v1beta1.VirtualMachineTemplateVersion) []Interface {
	result := []Interface{}
	origInterfaces := vmTemplateVersion.Spec.VM.Spec.Template.Spec.Domain.Devices.Interfaces

	for _, origInterface := range origInterfaces {
		networkName := ""
		networkType := ""
		for _, origNetwork := range vmTemplateVersion.Spec.VM.Spec.Template.Spec.Networks {
			if origNetwork.Name == origInterface.Name {

				if origNetwork.NetworkSource.Multus != nil {
					networkName = origNetwork.NetworkSource.Multus.NetworkName
					networkType = "multus"
				} else {
					networkType = "pod"

				}

			}
		}

		targetInterface := Interface{
			Name:        origInterface.Name,
			Type:        origInterface.Model,
			NetworkName: networkName,
			NetworkType: networkType,
		}
		result = append(result, targetInterface)
	}

	return result
}
