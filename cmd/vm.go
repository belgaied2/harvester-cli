package cmd

import (
	"os"
	"path"

	rcmd "github.com/rancher/cli/cmd"
	"github.com/urfave/cli"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	VMv1 "kubevirt.io/client-go/api/v1"
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

type VirtualMachineData struct {
	State          string
	VirtualMachine VMv1.VirtualMachine
	Name           string
	Node           string
	CPU            uint32
	Memory         string
	IPAddress      string
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

// GetHarvesterClient creates a Client for Harvester from Config input
func getHarvesterClient() (kubecli.KubevirtClient, error) {

	p := path.Join(os.ExpandEnv("${HOME}/.harvester"), "config")

	return kubecli.GetKubevirtClientFromFlags("", p)

}

func vmLs(ctx *cli.Context) error {

	c, err := getHarvesterClient()

	if err != nil {
		return err
	}

	vmList, err := c.VirtualMachine("default").List(&v1.ListOptions{})

	if err != nil {
		return err
	}

	vmiList, err := c.VirtualMachineInstance("default").List(&v1.ListOptions{})

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

func vmDelete(ctx *cli.Context) {

}

func vmCreate(ctx *cli.Context) {

}

// Start issues a power on for the virtual machine instance.
func vmStart(ctx *cli.Context) error {

	c, err := getHarvesterClient()
	if err != nil {
		return err
	}

	vm, err := c.VirtualMachine("default").Get(ctx.Args().First(), &v1.GetOptions{})

	*vm.Spec.Running = true

	if err != nil {
		return err
	}

	_, err = c.VirtualMachine("default").Update(vm)
	return err
}

// Stop issues a power off for the virtual machine instance.
func vmStop(ctx *cli.Context) error {

	c, err := getHarvesterClient()
	if err != nil {
		return err
	}

	vm, err := c.VirtualMachine("default").Get(ctx.Args().First(), &v1.GetOptions{})
	*vm.Spec.Running = false

	if err != nil {
		return err
	}

	_, err = c.VirtualMachine("default").Update(vm)
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
