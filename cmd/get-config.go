package cmd

import (
	"fmt"

	"github.com/rancher/cli/cliclient"
	rcmd "github.com/rancher/cli/cmd"
	client "github.com/rancher/types/client/management/v3"
	"github.com/urfave/cli"
)

func GetConfigCommand() cli.Command {
	return cli.Command{
		Name:    "get-config",
		Aliases: []string{"c"},
		Usage:   "Get KUBECONFIG of LOCAL cluster from Rancher",
		Action:  getConfig,
	}
}

func getConfig(ctx *cli.Context) error {

	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := rcmd.GetClient(ctx)
	if err != nil {
		return err
	}

	resource, err := rcmd.Lookup(c, ctx.Args().First(), "cluster")
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
	fmt.Println(config.Config)
	return nil
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
