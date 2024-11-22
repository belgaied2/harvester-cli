package cmd_login

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	cmd_config "github.com/belgaied2/harvester-cli/cmd/config"
	cmd_image "github.com/belgaied2/harvester-cli/cmd/image"
	"github.com/belgaied2/harvester-cli/common"
	"github.com/sirupsen/logrus"

	"github.com/grantae/certinfo"
	"github.com/rancher/cli/cliclient"
	rcmd "github.com/rancher/cli/cmd"
	"github.com/rancher/cli/config"
	managementClient "github.com/rancher/types/client/management/v3"
	"github.com/urfave/cli/v2"
)

// LoginData is a data structure to store login context information: project and ClusterName
type LoginData struct {
	Project     managementClient.Project
	Index       int
	ClusterName string
}

type CACertResponse struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func LoginCommand() *cli.Command {
	return &cli.Command{
		Name:      "login",
		Aliases:   []string{"l"},
		Usage:     "Login to a Rancher server",
		Action:    loginSetup,
		ArgsUsage: "[SERVERURL]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "context",
				Usage: "Set the context during login",
			},
			&cli.StringFlag{
				Name:  "token,t",
				Usage: "Token from the Rancher UI",
			},
			&cli.StringFlag{
				Name:  "cacert",
				Usage: "Location of the CACerts to use",
			},
			&cli.StringFlag{
				Name:  "name",
				Usage: "Name of the Server",
			},
			&cli.BoolFlag{
				Name:  "skip-verify",
				Usage: "Skip verification of the CACerts presented by the Server",
			},
			&cli.StringFlag{
				Name:  "harvester-config-path, path",
				Usage: "Defines the folder in which the Harvester Kubeconfig will be created",
			},
			&cli.StringFlag{
				Name:  "cluster",
				Usage: "Harvester cluster for which a Kubeconfig should be downloaded",
			},
		},
	}
}

// loginSetup implements the login cli command, it creates the necessary local configuration in order for Harvester CLI to be able to query Rancher and Harvester.
func loginSetup(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowCommandHelp(ctx, "login")
	}

	cf, err := common.LoadConfig(ctx)
	if err != nil {
		return err
	}

	serverName := ctx.String("name")
	if serverName == "" {
		serverName = "rancherDefault"
	}

	serverConfig := &config.ServerConfig{}

	// Validate the url and drop the path
	u, err := url.ParseRequestURI(ctx.Args().First())
	if err != nil {
		return fmt.Errorf("failed to parse SERVERURL (%s), make sure it is a valid HTTPS URL (e.g. https://rancher.yourdomain.com or https://1.1.1.1). Error: %s", ctx.Args().First(), err)
	}

	u.Path = ""
	serverConfig.URL = u.String()

	if ctx.String("token") != "" {
		auth := common.SplitOnColon(ctx.String("token"))
		if len(auth) != 2 {
			return errors.New("invalid token")
		}
		serverConfig.AccessKey = auth[0]
		serverConfig.SecretKey = auth[1]
		serverConfig.TokenKey = ctx.String("token")
	} else {
		// This can be removed once username and password is accepted
		return errors.New("token flag is required")
	}

	if ctx.String("cacert") != "" {
		cert, err := common.LoadAndVerifyCert(ctx.String("cacert"))
		if err != nil {
			return err
		}
		serverConfig.CACerts = cert

	}

	c, err := cliclient.NewManagementClient(serverConfig)
	if err != nil {
		if _, ok := err.(*url.Error); ok && strings.Contains(err.Error(), "certificate signed by unknown authority") {
			// no cert was provided and it's most likely a self signed cert if
			// we get here so grab the cacert and see if the user accepts the server
			c, err = getCertFromServer(ctx, serverConfig)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	proj, err := getProjectContext(ctx, c)
	if err != nil {
		return err
	}

	// Set the default server and proj for the user
	serverConfig.Project = proj
	cf.CurrentServer = serverName
	cf.Servers[serverName] = serverConfig

	_, err = os.Stat(cf.Path)
	if os.IsNotExist(err) {
		logrus.Info("configuration folder for Rancher does not exist, creating it")
		err = os.MkdirAll(filepath.Dir(cf.Path), 0700)
		if err != nil {
			return err
		}
	}
	err = cf.Write()
	if err != nil {
		return err
	}

	err = ctx.Set("cluster", common.SplitOnColon(proj)[0])

	if err != nil {
		return err
	}

	err = ctx.Set("path", ctx.String("harvester-config-path"))

	if err != nil {
		return err
	}

	err = cmd_config.GetConfig(ctx)
	if err != nil {
		return err
	}

	return nil
}

func getProjectContext(ctx *cli.Context, c *cliclient.MasterClient) (string, error) {
	// If context is given
	if ctx.String("context") != "" {
		context := ctx.String("context")
		// Check if given context is in valid format
		_, _, err := common.ParseClusterAndProjectID(context)
		if err != nil {
			return "", fmt.Errorf("unable to parse context (%s). Please provide context as local:p-xxxxx, c-xxxxx:p-xxxxx, or c-xxxxx:project-xxxxx", context)
		}
		// Check if context exists
		_, err = common.Lookup(c, context, "project")
		if err != nil {
			return "", fmt.Errorf("unable to find context (%s). Make sure the context exists and you have permissions to use it. Error: %s", context, err)
		}
		return context, nil
	}

	projectCollection, err := c.ManagementClient.Project.List(common.DefaultListOpts(ctx))
	if err != nil {
		return "", err
	}

	projDataLen := len(projectCollection.Data)

	if projDataLen == 0 {
		logrus.Warn("No projects found, context could not be set. Please create a project and run `rancher login` again.")
		return "", nil
	}

	if projDataLen == 1 {
		logrus.Infof("Only 1 project available: %s", projectCollection.Data[0].Name)
		return projectCollection.Data[0].ID, nil
	}

	if projDataLen == 2 {
		var sysProj bool
		var defaultID string
		for _, proj := range projectCollection.Data {
			if proj.Name == "Default" {
				defaultID = proj.ID
			}
			if proj.Name == "System" {
				sysProj = true
			}
			if sysProj && defaultID != "" {
				return defaultID, nil
			}
		}
	}

	clusterNames, err := common.GetClusterNames(ctx, c)
	if err != nil {
		return "", err
	}

	writer := rcmd.NewTableWriter([][]string{
		{"NUMBER", "Index"},
		{"CLUSTER NAME", "ClusterName"},
		{"PROJECT ID", "Project.ID"},
		{"PROJECT NAME", "Project.Name"},
		{"PROJECT DESCRIPTION", "Project.Description"},
	}, cmd_image.Ctxv1)

	for i, item := range projectCollection.Data {
		writer.Write(&LoginData{
			Project:     item,
			Index:       i + 1,
			ClusterName: clusterNames[item.ClusterID],
		})
	}

	writer.Close()
	if nil != writer.Err() {
		return "", writer.Err()
	}

	fmt.Print("Select a Project:")

	reader := bufio.NewReader(os.Stdin)

	var selection int
	selection, err = common.GetSelectionFromInput(reader, len(projectCollection.Data))

	if err != nil {
		return "", err
	}

	return projectCollection.Data[selection-1].ID, nil
}

func getCertFromServer(ctx *cli.Context, cf *config.ServerConfig) (*cliclient.MasterClient, error) {
	req, err := http.NewRequest("GET", cf.URL+"/v3/settings/cacerts", nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(cf.AccessKey, cf.SecretKey)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var certReponse *CACertResponse
	err = json.Unmarshal(content, &certReponse)
	if err != nil {
		return nil, fmt.Errorf("unable to parse response from %s/v3/settings/cacerts\nError: %s\nResponse:\n%s", cf.URL, err, content)
	}

	cert, err := common.VerifyCert([]byte(certReponse.Value))
	if err != nil {
		return nil, err
	}

	// Get the server cert chain in a printable form
	serverCerts, err := processServerChain(res)
	if err != nil {
		return nil, err
	}

	if !ctx.Bool("skip-verify") {
		if ok := verifyUserAcceptsCert(serverCerts, cf.URL); !ok {
			return nil, errors.New("CACert of server was not accepted, unable to login")
		}
	}

	cf.CACerts = cert

	return cliclient.NewManagementClient(cf)
}

func verifyUserAcceptsCert(certs []string, url string) bool {
	fmt.Printf("The authenticity of server '%s' can't be established.\n", url)
	fmt.Printf("Cert chain is : %v \n", certs)
	fmt.Print("Do you want to continue connecting (yes/no)? ")

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		input := scanner.Text()
		input = strings.ToLower(strings.TrimSpace(input))

		if input == "yes" || input == "y" {
			return true
		} else if input == "no" || input == "n" {
			return false
		}
		fmt.Printf("Please type 'yes' or 'no': ")
	}
	return false
}

func processServerChain(res *http.Response) ([]string, error) {
	var allCerts []string

	for _, cert := range res.TLS.PeerCertificates {
		result, err := certinfo.CertificateText(cert)
		if err != nil {
			return allCerts, err
		}
		allCerts = append(allCerts, result)
	}
	return allCerts, nil
}
