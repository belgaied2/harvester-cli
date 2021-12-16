package cmd

import (
	"fmt"
	"os"

	harvclient "github.com/harvester/harvester/pkg/generated/clientset/versioned"
	"github.com/urfave/cli"
	regen "github.com/zach-klippenstein/goregen"
	"kubevirt.io/client-go/kubecli"
)

//NewTrue returns a pointer to true
func NewTrue() *bool {
	b := true
	return &b
}

// RandomID returns a random string used as an ID internally in Harvester.
func RandomID() string {
	res, err := regen.Generate("[a-z]{3}[0-9][a-z]")
	if err != nil {
		fmt.Println("Random function was not successful!")
		return ""
	}
	return res
}

// GetHarvesterClient creates a Client for Harvester from Config input
func GetHarvesterClient(ctx *cli.Context) (Client, error) {
	p := os.ExpandEnv(ctx.GlobalString("config"))
	kubevirtClient, err := kubecli.GetKubevirtClientFromFlags("", p)

	if err != nil {
		return Client{}, err
	}

	harvesterClient, err := harvclient.NewForConfig(kubevirtClient.Config())
	if err != nil {
		return Client{}, err
	}

	return Client{
		KubevirtClient:  &kubevirtClient,
		HarvesterClient: harvesterClient,
	}, nil

}
