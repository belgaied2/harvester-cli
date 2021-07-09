package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/urfave/cli"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubevirt.io/client-go/kubecli"
)

const (
	vmAnnotationDescription = "field.cattle.io/description"
	vmAnnotationNetworkIps  = "networks.harvester.cattle.io/ips"
	dvAnnotationImageID     = "harvester.cattle.io/imageId"
	dvSourceHTTPURLPrefix   = "http://minio.harvester-system:9000/vm-images/"
	defaultSSHUser          = "ubuntu"
	// defaultVmLabels	= {
	// 	"harvester.cattle.io/creator": "harvester"}
	defaultVMName        = "test-vm"
	defaultVMDescription = "Test VM for Harvester"
	defaultDiskSize      = "10Gi"
	defaultMemSize       = "2Gi"
	defaultNbCPUCores    = 1
	defaultNamespace     = "default"
)

func VMCommand() cli.Command {
	return cli.Command{
		Name:    "vritualmachine",
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
				Flags:       []cli.Flag{},
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
			},
			{
				Name:      "create",
				Usage:     "Create a VM",
				Action:    vmCreate,
				ArgsUsage: "[VM_NAME]",
				Flags:     []cli.Flag{},
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

// GetHarvesterClient creates a Client for Harvester from Config input
func getHarvesterClient() (kubecli.KubevirtClient, error) {
	// kubecli.DefaultClientConfig() prepares config using kubeconfig.
	// typically, you need to set env variable, KUBECONFIG=<path-to-kubeconfig>/.kubeconfig
	p := path.Join(os.ExpandEnv("${HOME}/.harvester"), "config")

	return kubecli.GetKubevirtClientFromFlags("", p)
	// caCertBytes, errCA := base64.StdEncoding.DecodeString(d.CACertBase64)
	// certBytes, errCert := base64.StdEncoding.DecodeString(d.CertBase64)
	// keyBytes, errKey := base64.StdEncoding.DecodeString(d.KeyBase64)

	// if errCA != nil || errCert != nil || errKey != nil {
	// 	fmt.Println("An error happened during Base64 decoding of input certificate strings. The following error happened: %w", errCA)
	// }
	// clientConfig := restclient.Config{
	// 	Host: d.HarvesterHost,
	// 	TLSClientConfig: restclient.TLSClientConfig{
	// 		ServerName: "harvester",
	// 		CAData:     caCertBytes,
	// 		CertData:   certBytes,
	// 		KeyData:    keyBytes,
	// 	},
	// }

	// get the kubevirt client, using which kubevirt resources can be managed.
	// return kubecli.GetKubevirtClientFromRESTConfig(&clientConfig)

}

func vmLs(ctx *cli.Context) error {

	c, err := getHarvesterClient()

	if err != nil {
		return err
	}

	vmList, err := c.VirtualMachine("default").List(&v1.ListOptions{})

	for _, vm := range vmList.Items {

		fmt.Println(vm.Name)
	}

	return err
}

func vmDelete(ctx *cli.Context) {

}

func vmCreate(ctx *cli.Context) {

}
