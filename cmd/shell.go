package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/urfave/cli"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ShellCommand defines the CLI command that makes it possible to ssh into a VM
func ShellCommand() cli.Command {
	userHome, _ := os.UserHomeDir()
	return cli.Command{
		Name:      "shell",
		Aliases:   []string{"sh"},
		Usage:     "Access a VM using SSH",
		Action:    getShell,
		ArgsUsage: "VM_NAME",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "namespace, n",
				Usage:  "Namespace for the VM",
				EnvVar: "HARVESTER_VM_NAMESPACE",
				Value:  "default",
			},
			cli.StringFlag{
				Name:   "ssh-user, user",
				Usage:  "SSH user to be used for connecting to VM",
				EnvVar: "HARVESTER_VM_SSH_USER",
				Value:  "ubuntu",
			},
			cli.StringFlag{
				Name:   "ssh-key, i",
				Usage:  "Path to SSH Private Key to be used for connecting to VM",
				EnvVar: "HARVESTER_VM_SSH_KEY",
				Value:  userHome + "/.ssh/id_rsa",
			},
			cli.IntFlag{
				Name:   "ssh-port",
				Usage:  "TCP port to be used to connect to the VM using SSH, default is 22",
				EnvVar: "HARVESTER_VM_SSH_PORT",
				Value:  22,
			},
		},
	}
}

func getShell(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return fmt.Errorf("one and only one argument is accepted for this command, and that is the vm name")
	}

	vmName := ctx.Args().First()
	c, err := GetHarvesterClient(ctx)

	if err != nil {
		return err
	}

	vmi, err := c.KubevirtV1().VirtualMachineInstances(ctx.String("namespace")).Get(context.TODO(), vmName, v1.GetOptions{})

	if err != nil {
		return fmt.Errorf("no virtual machine instance with this name exists in harvester, please check that the it is created and started")
	}

	ipAddress := vmi.Status.Interfaces[0].IP
	sshPort := "22"
	// sshKey, err := getSSHKeyFromFile(ctx.String("ssh-key"))

	if err != nil {
		return err
	}

	// tcpAddr := ipAddress + ":" + sshPort
	sshConnString := "ubuntu@" + ipAddress

	cmd := exec.Command("ssh", "-i", ctx.String("ssh-key"), "-p", sshPort, sshConnString)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err = cmd.Run()

	if err != nil {
		return fmt.Errorf("error during execution of ssh command: %w", err)
	}

	return nil

}

// func getSSHKeyFromFile(file string) (ssh.AuthMethod, error) {
// 	buffer, err := ioutil.ReadFile(file)
// 	if err != nil {
// 		return nil, err
// 	}

// 	key, err := ssh.ParsePrivateKey(buffer)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return ssh.PublicKeys(key), nil
// }
