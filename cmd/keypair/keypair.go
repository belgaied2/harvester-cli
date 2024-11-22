package cmd_keypair

import (
	"context"
	"time"

	cmd_image "github.com/belgaied2/harvester-cli/cmd/image"
	cmd_vm "github.com/belgaied2/harvester-cli/cmd/vm"
	"github.com/belgaied2/harvester-cli/common"
	rcmd "github.com/rancher/cli/cmd"
	"github.com/urfave/cli/v2"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KeypairData struct {
	Name              string
	Fingerprint       string
	CreationTimestamp string
}

// TemplateCommand defines the CLI command that lists VM templates in Harvester
func KeypairCommand() *cli.Command {
	return &cli.Command{
		Name:    "keypair",
		Aliases: []string{"key", "ssh-key"},
		Usage:   "Manipulate SSH Keypairs",
		Action:  keypairList,
		Flags: []cli.Flag{
			&cmd_vm.NamespaceFlag,
		},
		Subcommands: cli.Commands{
			&cli.Command{
				Name:        "list",
				Aliases:     []string{"ls"},
				Usage:       "List SSH Keypairs",
				Description: "\nLists all the SSH Keypairs available in Harvester",
				ArgsUsage:   "None",
				Action:      keypairList,
				Flags: []cli.Flag{
					&cmd_vm.NamespaceFlag,
				},
			},
		},
	}
}

func keypairList(ctx *cli.Context) (err error) {
	c, err := common.GetHarvesterClient(ctx)

	if err != nil {
		return
	}

	keyList, err := c.HarvesterhciV1beta1().KeyPairs(ctx.String("namespace")).List(context.TODO(), k8smetav1.ListOptions{})

	if err != nil {
		return
	}

	writer := rcmd.NewTableWriter([][]string{
		{"NAME", "Name"},
		{"FINGERPRINT", "Fingerprint"},
		{"CREATION TIMESTAMP", "CreationTimestamp"},
	},
		cmd_image.Ctxv1)

	defer writer.Close()

	for _, keyItem := range keyList.Items {

		writer.Write(&KeypairData{
			Name:              keyItem.Name,
			Fingerprint:       keyItem.Status.FingerPrint,
			CreationTimestamp: keyItem.CreationTimestamp.Format(time.RFC822),
		})

	}

	return writer.Err()
}
