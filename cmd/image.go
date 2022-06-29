package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	rcmd "github.com/rancher/cli/cmd"
	"github.com/urfave/cli"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ImageData struct {
	Name       string
	Id         string
	SourceType string
	Url        string
}

// TemplateCommand defines the CLI command that lists VM templates in Harvester
func ImageCommand() cli.Command {
	return cli.Command{
		Name:    "image",
		Aliases: []string{"img"},
		Usage:   "Manipulate VM images",
		Action:  imageList,
		Flags: []cli.Flag{
			nsFlag,
		},
		Subcommands: cli.Commands{
			cli.Command{
				Name:        "list",
				Aliases:     []string{"ls"},
				Usage:       "List VM images",
				Description: "\nLists all the VM images available in Harvester",
				ArgsUsage:   "",
				Action:      imageList,
				Flags: []cli.Flag{
					nsFlag,
				},
			},
			cli.Command{
				Name:        "create",
				Aliases:     []string{"add"},
				Usage:       "Creates a VM image",
				Description: "\nCreates a VM image from a source location",
				ArgsUsage:   "VM_IMAGE_DISPLAYNAME",
				Action:      imageCreate,
				Flags: []cli.Flag{
					nsFlag,
					cli.StringFlag{
						Name:     "source",
						Usage:    "Location from which the image will be put into Harvester, this should be an HTTP(S) link that harvester will use to download the image",
						EnvVar:   "HARVESTER_VM_IMAGE_LINK",
						Required: true,
					},
					cli.StringFlag{
						Name:     "description",
						Usage:    "Description of the VM Image",
						EnvVar:   "HARVESTER_VM_IMAGE_DESCRIPTION",
						Required: false,
					},
				},
			},
		},
	}
}

func imageList(ctx *cli.Context) (err error) {
	c, err := GetHarvesterClient(ctx)

	if err != nil {
		return
	}

	imgList, err := c.HarvesterhciV1beta1().VirtualMachineImages(ctx.String("namespace")).List(context.TODO(), k8smetav1.ListOptions{})

	if err != nil {
		return
	}

	writer := rcmd.NewTableWriter([][]string{
		{"NAME", "Name"},
		{"ID", "Id"},
		{"SOURCE TYPE", "SourceType"},
		{"URL", "Url"},
	},
		ctx)

	defer writer.Close()

	for _, imgItem := range imgList.Items {

		writer.Write(&ImageData{
			Name:       imgItem.Spec.DisplayName,
			Id:         imgItem.Name,
			SourceType: imgItem.Spec.SourceType,
			Url:        imgItem.Spec.URL,
		})

	}

	return writer.Err()
}

// imageCreate create a VM Image in Harvester based on a URL and a display name as well as an optional description
func imageCreate(ctx *cli.Context) (err error) {
	if len(ctx.Args()) != 1 {
		err = fmt.Errorf("wrong number of arguments")
	}

	if err != nil {
		return
	}

	vmImageDisplayName := ctx.Args()[0]
	source := ctx.String("source")
	if !strings.HasPrefix(source, "http") { //If the upload option is implemented, this will need to change!
		err = fmt.Errorf("source flag is not a valid http link")
	}

	if err != nil {
		return
	}

	vmImage := &v1beta1.VirtualMachineImage{
		ObjectMeta: k8smetav1.ObjectMeta{
			GenerateName: "image-",
		},
		Spec: v1beta1.VirtualMachineImageSpec{
			Description: ctx.String("description"),
			DisplayName: vmImageDisplayName,
			SourceType:  "download", //If the upload option is implemented, this will need to change!
			URL:         source,
		},
	}

	c, err := GetHarvesterClient(ctx)

	if err != nil {
		return
	}

	_, err = c.HarvesterhciV1beta1().VirtualMachineImages(ctx.String("namespace")).Create(context.TODO(), vmImage, k8smetav1.CreateOptions{})

	return
}
