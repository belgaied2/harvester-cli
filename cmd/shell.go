package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"

	"github.com/urfave/cli"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	portforwardclgo "k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/kubectl/pkg/cmd/portforward"
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
			cli.BoolFlag{
				Name:   "pod-network",
				Usage:  "Options to connect to VM through pod network",
				EnvVar: "HARVESTER_VM_POD_NETWORK",
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

	restCl, restConf, err := GetRESTClientAndConfig(ctx)

	if err != nil {
		return fmt.Errorf("error when setting up Kubernetes API client: %w", err)
	}

	k, err := GetKubeClient(ctx)

	if err != nil {
		return fmt.Errorf("error when setting up Kubernetes API client: %w", err)
	}

	vmi, err := c.KubevirtV1().VirtualMachineInstances(ctx.String("namespace")).Get(context.TODO(), vmName, v1.GetOptions{})

	if err != nil {
		return fmt.Errorf("no virtual machine instance with this name exists in harvester, please check that the it is created and started")
	}

	vmPodList, err := k.CoreV1().Pods(ctx.String("namespace")).List(context.TODO(), v1.ListOptions{
		LabelSelector: "harvesterhci.io/vmNamePrefix=" + vmName,
	})

	if len(vmPodList.Items) == 0 {
		vmPodList, err = k.CoreV1().Pods(ctx.String("namespace")).List(context.TODO(), v1.ListOptions{
			LabelSelector: "harvesterhci.io/vmName=" + vmName,
		})
	}

	var ipAddress string
	var sshPort string

	if !ctx.Bool("pod-network") {
		ipAddress = vmi.Status.Interfaces[0].IP
		sshPort = "22"
	} else {
		o := &portforward.PortForwardOptions{
			Namespace:  ctx.String("namespace"),
			RESTClient: restCl,
			Config:     restConf,
			PodName:    vmPodList.Items[0].Name,
			Address:    []string{"localhost"},
			Ports:      []string{"2222", "22"},
			PortForwarder: &defaultPortForwarder{
				IOStreams: genericclioptions.IOStreams{
					In:     os.Stdin,
					Out:    os.Stdout,
					ErrOut: os.Stderr,
				},
			},
			PodClient:    k.CoreV1(),
			StopChannel:  make(chan struct{}, 1),
			ReadyChannel: make(chan struct{}),
		}

		fmt.Println("pod name:" + vmPodList.Items[0].Name)

		err = o.RunPortForward()

		if err != nil {
			fmt.Println(errors.Unwrap(err))
			return fmt.Errorf("error during setting port-forwarding for VM in pod networking %w", err)

		}

		ipAddress = "localhost"
		sshPort = "2222"
	}

	// sshKey, err := getSSHKeyFromFile(ctx.String("ssh-key"))

	if err != nil {
		return err
	}

	// tcpAddr := ipAddress + ":" + sshPort
	sshConnString := ctx.String("ssh-user") + "@" + ipAddress

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

type defaultPortForwarder struct {
	genericclioptions.IOStreams
}

func (f *defaultPortForwarder) ForwardPorts(method string, url *url.URL, opts portforward.PortForwardOptions) error {
	transport, upgrader, err := spdy.RoundTripperFor(opts.Config)
	if err != nil {
		return err
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, method, url)
	fw, err := portforwardclgo.NewOnAddresses(dialer, opts.Address, opts.Ports, opts.StopChannel, opts.ReadyChannel, f.Out, f.ErrOut)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
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
