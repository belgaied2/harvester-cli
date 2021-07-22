package cmd

import (
	"fmt"
	"os"
	"path"

	harvclient "github.com/harvester/harvester/pkg/generated/clientset/versioned"
	regen "github.com/zach-klippenstein/goregen"
	"kubevirt.io/client-go/kubecli"
)

func NewTrue() *bool {
	b := true
	return &b
}

// randomID returns a random string used as an ID internally in Harvester.
func RandomID() string {
	res, err := regen.Generate("[a-z]{3}[0-9][a-z]")
	if err != nil {
		fmt.Println("Random function was not successful!")
		return ""
	}
	return res
}

// GetHarvesterClient creates a Client for Harvester from Config input
func GetHarvesterClient() (Client, error) {

	p := path.Join(os.ExpandEnv("${HOME}/.harvester"), "config")
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
