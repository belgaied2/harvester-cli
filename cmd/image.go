package cmd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strconv"
	"strings"

	"net/http"
	"net/url"

	"github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	rcmd "github.com/rancher/cli/cmd"
	"github.com/rancher/cli/config"
	"github.com/urfave/cli"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

type ImageData struct {
	Name       string
	Id         string
	SourceType string
	Url        string
}

// TemplateCommand defines the CLI command that lists VM templates in Harvester
func ImageCommand() cli.Command {
	return cli.Command{
		Name:    "image",
		Aliases: []string{"img"},
		Usage:   "Manipulate VM images",
		Action:  imageList,
		Flags: []cli.Flag{
			nsFlag,
		},
		Subcommands: cli.Commands{
			cli.Command{
				Name:        "list",
				Aliases:     []string{"ls"},
				Usage:       "List VM images",
				Description: "\nLists all the VM images available in Harvester",
				ArgsUsage:   "",
				Action:      imageList,
				Flags: []cli.Flag{
					nsFlag,
				},
			},
			cli.Command{
				Name:        "create",
				Aliases:     []string{"add"},
				Usage:       "Creates a VM image",
				Description: "\nCreates a VM image from a source location",
				ArgsUsage:   "VM_IMAGE_DISPLAYNAME",
				Action:      imageCreate,
				Flags: []cli.Flag{
					nsFlag,
					cli.StringFlag{
						Name:     "source",
						Usage:    "Location from which the image will be put into Harvester, this should be an HTTP(S) link that harvester will use to download the image",
						EnvVar:   "HARVESTER_VM_IMAGE_LINK",
						Required: true,
					},
					cli.StringFlag{
						Name:     "description",
						Usage:    "Description of the VM Image",
						EnvVar:   "HARVESTER_VM_IMAGE_DESCRIPTION",
						Required: false,
					},
				},
			},
		},
	}
}

func imageList(ctx *cli.Context) (err error) {
	c, err := GetHarvesterClient(ctx)

	if err != nil {
		return
	}

	imgList, err := c.HarvesterhciV1beta1().VirtualMachineImages(ctx.String("namespace")).List(context.TODO(), k8smetav1.ListOptions{})

	if err != nil {
		return
	}

	writer := rcmd.NewTableWriter([][]string{
		{"NAME", "Name"},
		{"ID", "Id"},
		{"SOURCE TYPE", "SourceType"},
		{"URL", "Url"},
	},
		ctx)

	defer writer.Close()

	for _, imgItem := range imgList.Items {

		writer.Write(&ImageData{
			Name:       imgItem.Spec.DisplayName,
			Id:         imgItem.Name,
			SourceType: imgItem.Spec.SourceType,
			Url:        imgItem.Spec.URL,
		})

	}

	return writer.Err()
}

// imageCreate create a VM Image in Harvester based on a URL and a display name as well as an optional description
func imageCreate(ctx *cli.Context) (err error) {
	if len(ctx.Args()) != 1 {
		err = fmt.Errorf("wrong number of arguments")
	}

	if err != nil {
		return
	}

	vmImageDisplayName := ctx.Args()[0]
	source := ctx.String("source")
	sourceType := "download"
	if !strings.HasPrefix(source, "http") { //If the upload option is implemented, this will need to change!
		var fileInf fs.FileInfo
		if fileInf, err = os.Stat(source); err == nil {
			filesize := fileInf.Size()
			sourceType = "upload"
			var vmImageCreateName string
			vmImageCreateName, err = createImageObjectInAPI(ctx, vmImageDisplayName, sourceType, source)
			if err != nil {
				return
			}
			var rancherServerConfig *config.ServerConfig
			var harvesterURL string
			rancherServerConfig, harvesterURL, err = getHarvesterAPIFromConfig(ctx)

			if err != nil {
				return
			}

			urlToSendFile := harvesterURL + "/v1/harvester/harvesterhci.io.virtualmachineimages/" + ctx.String("namespace") + "/" + vmImageCreateName + "/action=upload&size=" + strconv.FormatInt(filesize, 10)
			var fileReader io.Reader
			fileReader, err = os.Open(source)
			if err != nil {
				return
			}

			var req *http.Request

			req, err = http.NewRequest("POST", urlToSendFile, fileReader)
			if err != nil {
				return
			}

			rootCAs, err := x509.SystemCertPool()
			if err != nil {
				return err
			}
			pemBlock, _ := pem.Decode([]byte(rancherServerConfig.CACerts))
			ownCert, err := x509.ParseCertificate(pemBlock.Bytes)
			if err != nil {
				return fmt.Errorf("invalid CA certification in Rancher configuration, %w", err)
			}

			rootCAs.AddCert(ownCert)

			req.Header.Add("Authorization", "Bearer "+rancherServerConfig.TokenKey)
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: rootCAs,
				},
			}
			httpClient := &http.Client{Transport: tr}

			var resp *http.Response
			resp, err = httpClient.Do(req)

			if err != nil {
				return err
			}

			if resp.Status == "200 OK" {
				return nil
			} else {
				return fmt.Errorf("uploading image file to harvester was not successful")
			}

		} else {

			err = fmt.Errorf("source flag is neither a valid http link and nor a valid filepath")
			return

		}

	}
	_, err = createImageObjectInAPI(ctx, vmImageDisplayName, sourceType, source)

	return

}

func createImageObjectInAPI(ctx *cli.Context, vmImageDisplayName string, sourceType string, source string) (vmImageCreateName string, err error) {

	if sourceType == "upload" {
		source = ""
	}

	vmImage := &v1beta1.VirtualMachineImage{
		ObjectMeta: k8smetav1.ObjectMeta{
			GenerateName: "image-",
		},
		Spec: v1beta1.VirtualMachineImageSpec{
			Description: ctx.String("description"),
			DisplayName: vmImageDisplayName,
			SourceType:  sourceType,
			URL:         source,
		},
	}

	c, err := GetHarvesterClient(ctx)

	if err != nil {
		return
	}

	vmImageCreated, err := c.HarvesterhciV1beta1().VirtualMachineImages(ctx.String("namespace")).Create(context.TODO(), vmImage, k8smetav1.CreateOptions{})

	if err != nil {
		return
	}

	vmImageCreateName = vmImageCreated.Name
	return
}

func getHarvesterAPIFromConfig(ctx *cli.Context) (serverConfig *config.ServerConfig, harvesterKubeAPIServerURL string, err error) {

	p := os.ExpandEnv(ctx.GlobalString("harvester-config"))
	restConfig, err := clientcmd.BuildConfigFromFlags("", p)

	if err != nil {
		return
	}

	harvesterKubeAPIServerURL = restConfig.Host
	u, err := url.Parse(harvesterKubeAPIServerURL)

	if err != nil {
		return
	}

	harvesterKubeAPIServerHost := u.Host

	tokenMap, configMap, err := GetRancherTokenMap(ctx)

	if err != nil {
		return
	}

	var ok bool = false

	if _, ok = tokenMap[harvesterKubeAPIServerHost]; ok {
		serverConfig = configMap[harvesterKubeAPIServerHost]
		return
	} else {
		return nil, "", fmt.Errorf("not able to determine harvester API URL")
	}

}
