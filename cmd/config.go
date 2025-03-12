package cmd

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/rancher/cli/cliclient"
	rcmd "github.com/rancher/cli/cmd"
	client "github.com/rancher/types/client/management/v3"
	"github.com/sirupsen/logrus"
	cliv1 "github.com/urfave/cli"
	"github.com/urfave/cli/v2"
)

const (
	kubeConfigFilename = "config"
)

// Conf is an Object that contains the configuration path and the configuration's file content as a string
type Conf struct {
	Path    string
	Content string
}

// ConfigCommand defines a CLI command to set up the Harvester Configuration files
func ConfigCommand() *cli.Command {
	return &cli.Command{
		Name:    "get-config",
		Aliases: []string{"c"},
		Usage:   "Get KUBECONFIG of Harvester cluster from Rancher",
		Action:  GetConfig,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path",
				Usage: "Set the path where to store the KUBE config file",
			},
			&cli.StringFlag{
				Name:     "cluster",
				Usage:    "Name of the cluster in Rancher for which the KUBECONFIG will be generated",
				EnvVars:  []string{"HARVESTER_CLUSTER_NAME"},
				Required: true,
				Value:    "local",
			},
		},
	}
}

func GetConfig(ctx *cli.Context) error {

	userHome, err := os.UserHomeDir()

	if err != nil {
		return err
	}
	p := ctx.String("path")
	if p == "" {
		p = path.Join(userHome, ".harvester")
	}

	cf := Conf{
		Path:    path.Join(p, kubeConfigFilename),
		Content: "",
	}

	flags := flag.NewFlagSet("get-config", flag.ContinueOnError)
	flags.String("config", "", "config content")
	flags.String("path", "", "path to the file")

	ctxv1 := cliv1.NewContext(&cliv1.App{Name: "harvester"}, flags, nil)
	err = ctxv1.Set("config", ctx.String("config"))
	if err != nil {
		return fmt.Errorf("error setting config flag: %w", err)
	}

	c, err := rcmd.GetClient(ctxv1)
	if err != nil {
		return err
	}

	resource, err := rcmd.Lookup(c, ctx.String("cluster"), "cluster")
	if err != nil {
		return err
	}

	cluster, err := getClusterByID(c, resource.ID)
	if err != nil {
		return err
	}

	config, err := c.ManagementClient.Cluster.ActionGenerateKubeconfig(cluster)
	if err != nil {
		return err
	}

	cf.Content = config.Config
	return createKubeconfigFile(cf)
}

func getClusterByID(
	c *cliclient.MasterClient,
	clusterID string,
) (*client.Cluster, error) {
	cluster, err := c.ManagementClient.Cluster.ByID(clusterID)
	if err != nil {
		return nil, fmt.Errorf("no cluster found with the ID [%s], run "+
			"`rancher clusters` to see available clusters: %s", clusterID, err)
	}
	return cluster, nil
}

func createKubeconfigFile(config Conf) error {
	err := os.MkdirAll(path.Dir(config.Path), 0700)

	if err != nil {
		return err
	}

	logrus.Infof("Saving config to %s", config.Path)
	p := config.Path

	output, err := os.Create(p)
	if err != nil {
		return err
	}

	l, err := output.WriteString(config.Content)
	logrus.Infof("Successfully written %d bytes to %s", l, config.Path)
	return err
}
