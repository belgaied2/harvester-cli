package cmd_shell

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/belgaied2/harvester-cli/common"
	"github.com/harvester/harvester/pkg/generated/clientset/versioned"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	portforwardclgo "k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/kubectl/pkg/cmd/portforward"
)

// ShellCommand defines the CLI command that makes it possible to ssh into a VM
func ShellCommand() *cli.Command {
	userHome, _ := os.UserHomeDir()
	return &cli.Command{
		Name:      "shell",
		Aliases:   []string{"sh"},
		Usage:     "Access a VM using SSH",
		Action:    getShell,
		ArgsUsage: "VM_NAME",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "namespace, n",
				Usage:   "Namespace for the VM",
				EnvVars: []string{"HARVESTER_VM_NAMESPACE"},
				Value:   "default",
			},
			&cli.StringFlag{
				Name:    "ssh-user, user",
				Usage:   "SSH user to be used for connecting to VM",
				EnvVars: []string{"HARVESTER_VM_SSH_USER"},
				Value:   "ubuntu",
			},
			&cli.StringFlag{
				Name:    "ssh-key, i",
				Usage:   "Path to SSH Private Key to be used for connecting to VM",
				EnvVars: []string{"HARVESTER_VM_SSH_KEY"},
				Value:   userHome + "/.ssh/id_rsa",
			},
			&cli.IntFlag{
				Name:    "ssh-port",
				Usage:   "TCP port to be used to connect to the VM using SSH, default is 22",
				EnvVars: []string{"HARVESTER_VM_SSH_PORT"},
				Value:   22,
			},
			&cli.BoolFlag{
				Name:    "pod-network",
				Usage:   "Options to connect to VM through pod network",
				EnvVars: []string{"HARVESTER_VM_POD_NETWORK"},
			},
		},
	}
}

// getShell implements the command `shell`
// It accepts only one argument, that should be the VM name
func getShell(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return fmt.Errorf("one and only one argument is accepted for this command, and that is the vm name")
	}

	vmName := ctx.Args().First()
	c, err := common.GetHarvesterClient(ctx)

	if err != nil {
		return err
	}

	restConf, err := common.GetRESTClientAndConfig(ctx)

	if err != nil {
		return fmt.Errorf("error when setting up Kubernetes API client: %w", err)
	}

	k, err := common.GetKubeClient(ctx)

	if err != nil {
		return fmt.Errorf("error when setting up Kubernetes API client: %w", err)
	}

	vmi, err := c.KubevirtV1().VirtualMachineInstances(ctx.String("namespace")).Get(context.TODO(), vmName, v1.GetOptions{})

	if err != nil {
		return fmt.Errorf("no virtual machine instance with this name exists in harvester, please check that the it is created and started")
	}

	var ipAddress string
	var sshPort string

	netType, networkNum, err := networkType(vmName, c, ctx)

	if err != nil {
		return fmt.Errorf("error determining VM's network type: %w", err)
	}

	if netType == "pod" || ctx.Bool("pod-network") {

		sshPort, err = getFreeLocalPort()
		if err != nil {
			return fmt.Errorf("unable to find free local port: %w", err)
		}

		err = sshOverPortForward(k, ctx, vmName, sshPort, restConf)
		if err != nil {
			return fmt.Errorf("ssh over Port Forwarding failed: %w", err)
		}

	} else {
		ipAddress = vmi.Status.Interfaces[networkNum].IP
		sshPort = "22"

		if ipAddress == "" {
			return fmt.Errorf("the designated VM does not have a valid IP Address")
		}

		err = doSSH(ctx, ipAddress, sshPort)
		if err != nil {
			return err
		}

	}

	return nil

}

// networkType finds out to which network interface to connect to the VM
// if a bridge network interface exists, it will be returned
// if no bridge network interface exists, but a Pod Network interface does, it will use the last one it encounters
// if no interface could be defined, it throws an error
func networkType(vmName string, c *versioned.Clientset, ctx *cli.Context) (string, int, error) {

	vm, err := c.KubevirtV1().VirtualMachines(ctx.String("namespace")).Get(context.TODO(), vmName, v1.GetOptions{})
	if err != nil {
		return "", 0, fmt.Errorf("error querying VM object: %w", err)
	}
	onlyPodNetwork := false
	podNetworkNumber := 0

	for i, network := range vm.Spec.Template.Spec.Networks {
		if network.Multus != nil {
			return "bridge", i, nil
		} else if network.Pod != nil {
			onlyPodNetwork = true
			podNetworkNumber = i
		}
	}
	if onlyPodNetwork {
		return "pod", podNetworkNumber, nil
	}

	return "", 0, fmt.Errorf("no valid network type found for VM: %s", vmName)
}

// getFreeLocalPort finds a random free port on the local machine as a source to the port forwarding.
func getFreeLocalPort() (string, error) {

	//TODO: Change implementation
	return "32222", nil
}

// sshOverPortFoward contains the steps to make an SSH connection to a VM that is on the PodNetwork and not on the bridge network.
// It first finds out what is the Pod that is driving the VM
// Then, it populates the PortForwardOptions struct that is used as a container for all the parameters necessary to port forward using Kubernetes API
// Finally, it relies on Go Routines to open the Port forwarding tunnel and do the actual SSH connection
func sshOverPortForward(k *kubernetes.Clientset, ctx *cli.Context, vmName string, sshPort string, restConf *rest.Config) error {
	var err error
	vmPodList, _ := k.CoreV1().Pods(ctx.String("namespace")).List(context.TODO(), v1.ListOptions{
		LabelSelector: "harvesterhci.io/vmNamePrefix=" + vmName,
	})

	if len(vmPodList.Items) == 0 {
		vmPodList, err = k.CoreV1().Pods(ctx.String("namespace")).List(context.TODO(), v1.ListOptions{
			LabelSelector: "harvesterhci.io/vmName=" + vmName,
		})

		if err != nil {
			return fmt.Errorf("unable to find pods for the VM:%s, error: %w", vmName, err)
		}
	}

	ipAddress := "localhost"

	o := &portforward.PortForwardOptions{
		Namespace:    ctx.String("namespace"),
		Config:       restConf,
		PodName:      vmPodList.Items[0].Name,
		Address:      []string{ipAddress},
		Ports:        []string{sshPort + ":22"},
		PodClient:    k.CoreV1(),
		StopChannel:  make(chan struct{}, 1),
		ReadyChannel: make(chan struct{}),
	}
	var wg sync.WaitGroup
	wg.Add(1)

	fmt.Println("pod name:" + vmPodList.Items[0].Name)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("Bye...")
		close(o.StopChannel)
		wg.Done()
	}()

	go func() {

		err := doPortForward(o)
		if err != nil {
			panic(err)
		}
	}()

	<-o.ReadyChannel
	err = doSSH(ctx, ipAddress, sshPort)
	if err != nil {
		return err
	}

	wg.Done()
	return nil
}

// doSSH implements the actual SSHing into the VM. For simplicity's sake, it relies on the system's SSH command, usually present on all major OSes
// Linux, Windows, MacOS
func doSSH(ctx *cli.Context, ipAddress string, sshPort string) error {
	sshConnString := ctx.String("ssh-user") + "@" + ipAddress

	cmd := exec.Command("ssh", "-i", ctx.String("ssh-key"), "-p", sshPort, sshConnString)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()

	if err != nil {
		return fmt.Errorf("error during execution of ssh command: %w", err)
	}
	return nil
}

// doPortForward implements the actual Port forwarding before an SSH connection can be done
// it relies on the content of the PortForwardOptions struct defined in the upstream kubectl project
func doPortForward(o *portforward.PortForwardOptions) error {
	roundTripper, upgrader, err := spdy.RoundTripperFor(o.Config)
	if err != nil {
		panic(err)
	}

	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", o.Namespace, o.PodName)
	hostIP := strings.TrimLeft(o.Config.Host, "htps:/")
	serverURL := url.URL{Scheme: "https", Path: path, Host: hostIP}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, &serverURL)
	var berr, bout bytes.Buffer
	buffErr := bufio.NewWriter(&berr)
	buffOut := bufio.NewWriter(&bout)

	fw, err := portforwardclgo.New(dialer, o.Ports, o.StopChannel, o.ReadyChannel, buffOut, buffErr)

	if err != nil {
		return fmt.Errorf("error when creating portforwarder Object: %w", err)
	}

	err = fw.ForwardPorts()

	if err != nil {
		logrus.Error(buffErr)
		return fmt.Errorf("port forwarding failed: %w", err)
	}
	return nil
}
