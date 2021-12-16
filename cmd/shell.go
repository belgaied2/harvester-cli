package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"

	"github.com/shiena/ansicolor"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
	c, err := GetHarvesterClient()

	if err != nil {
		return err
	}

	vmi, err := (*c.KubevirtClient).VirtualMachineInstance(ctx.String("namespace")).Get(vmName, &v1.GetOptions{})

	if err != nil {
		return fmt.Errorf("no virtual machine instance with this name exists in harvester, please check that the it is created and started")
	}

	ipAddress := vmi.Status.Interfaces[0].IP
	sshKey, err := getSSHKeyFromFile(ctx.String("ssh-key"))

	if err != nil {
		return err
	}

	config := ssh.ClientConfig{
		User: ctx.String("ssh-user"),
		Auth: []ssh.AuthMethod{
			sshKey,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshServer := ipAddress + ":" + fmt.Sprintf("%d", ctx.Int("ssh-port"))
	conn, err := ssh.Dial("tcp", sshServer, &config)
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}
	defer conn.Close()

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := conn.NewSession()
	if err != nil {
		panic("Failed to create session: " + err.Error())
	}
	defer session.Close()

	// Set IO
	session.Stdout = ansicolor.NewAnsiColorWriter(os.Stdout)
	session.Stderr = ansicolor.NewAnsiColorWriter(os.Stderr)
	in, _ := session.StdinPipe()

	// Set up terminal modes
	// https://net-ssh.github.io/net-ssh/classes/Net/SSH/Connection/Term.html
	// https://www.ietf.org/rfc/rfc4254.txt
	// https://godoc.org/golang.org/x/crypto/ssh
	// THIS IS THE TITLE
	// https://pythonhosted.org/ANSIColors-balises/ANSIColors.html
	modes := ssh.TerminalModes{
		ssh.ECHO:  0, // Disable echoing
		ssh.IGNCR: 1, // Ignore CR on input.
	}

	// Request pseudo terminal
	//if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
	if err := session.RequestPty("xterm-256color", 80, 40, modes); err != nil {
		//if err := session.RequestPty("vt100", 80, 40, modes); err != nil {
		//if err := session.RequestPty("vt220", 80, 40, modes); err != nil {
		log.Fatalf("request for pseudo terminal failed: %s", err)
	}

	// Start remote shell
	if err := session.Shell(); err != nil {
		log.Fatalf("failed to start shell: %s", err)
	}

	// Handle control + C
	cha := make(chan os.Signal, 1)

	sshCtx, close := context.WithCancel(context.Background())

	signal.Notify(cha, os.Interrupt)

	// Go Routine waiting for OS Interrupts, and cancels the context if it gets one
	go func() {
		<-cha
		fmt.Printf("\nreceived interrupt, exiting")
		close()
	}()

	// Iterates on commands until context is cancelled from
	for {

		select {
		case <-sshCtx.Done():
			return nil
		default:
			reader := bufio.NewReader(os.Stdin)
			str, _ := reader.ReadString('\n')
			fmt.Fprint(in, str)

		}

	}

}

func getSSHKeyFromFile(file string) (ssh.AuthMethod, error) {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(key), nil
}
