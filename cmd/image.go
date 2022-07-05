package cmd

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	rcmd "github.com/rancher/cli/cmd"
	"github.com/rancher/cli/config"
	"github.com/sirupsen/logrus"
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

type CatalogEntry struct {
	Id        int64  `json:"id,omitempty"`
	ShortName string `json:"shortName"`
	Version   string `json:"version"`
	Url       string `json:"url"`
	Build     string `json:"build"`
}

type Catalog struct {
	HarvesterImageCatalog map[string][]CatalogEntry `json:"HarvesterImageCatalog"`
}

type Os struct {
	Id             int64
	Name           string
	NumberOfImages string
}

const (
	defaultCatalogSource = "https://raw.githubusercontent.com/belgaied2/harvester-cli/feature-image-upload/image-metadata.json"
)

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
						Usage:    "Location from which the image will be put into Harvester, this should be either an HTTP(S) link or a path to a file that harvester will use to get the image",
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
			cli.Command{
				Name:        "catalog",
				Aliases:     []string{"cat"},
				Usage:       "lists an image catalog",
				Description: "\nShows a list of freely available linux images to download from URLs",
				ArgsUsage:   "",
				Action:      imageCatalog,
				Flags: []cli.Flag{
					nsFlag,
					cli.StringFlag{
						Name:     "metadata-url",
						Usage:    "Location from which to get the metadata JSON file",
						EnvVar:   "HARVESTER_CATALOG_METADATA",
						Required: false,
						Value:    defaultCatalogSource,
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
	if !strings.HasPrefix(source, "http") {
		var fileInf fs.FileInfo
		if fileInf, err = os.Stat(source); err == nil {
			logrus.Debug("Source is a valid file!")
			filesize := fileInf.Size()
			sourceType = "upload"

			var rancherServerConfig *config.ServerConfig
			var harvesterURL string
			rancherServerConfig, harvesterURL, err = getHarvesterAPIFromConfig(ctx)

			if err != nil {
				return
			}
			logrus.Info("Successfully computed URL and credentials to Harvester!")

			var fileReader io.Reader
			fileReader, err = os.Open(source)
			if err != nil {
				return
			}

			var req *http.Request
			multipartBody := &bytes.Buffer{}
			writer := multipart.NewWriter(multipartBody)
			var part io.Writer
			part, err = writer.CreateFormFile("chunk", filepath.Base(source))

			if err != nil {
				return
			}

			_, err = io.Copy(part, fileReader)
			logrus.Info("Successfully preparated file for upload!")
			if err != nil {
				return
			}

			err = writer.Close()
			if err != nil {
				return
			}

			var vmImageCreateName string
			vmImageCreateName, err = createImageObjectInAPI(ctx, vmImageDisplayName, sourceType, source)
			if err != nil {
				return
			}
			logrus.Info("Image Object successfully created in Kubernetes API!")
			urlToSendFile := harvesterURL + "/v1/harvester/harvesterhci.io.virtualmachineimages/" + ctx.String("namespace") + "/" + vmImageCreateName + "?action=upload&size=" + strconv.FormatInt(filesize, 10)

			req, err = http.NewRequest("POST", urlToSendFile, multipartBody)
			if err != nil {
				return
			}
			var rootCAs *x509.CertPool
			rootCAs, err = x509.SystemCertPool()
			if err != nil {
				return err
			}
			pemBlock, _ := pem.Decode([]byte(rancherServerConfig.CACerts))
			var ownCert *x509.Certificate
			ownCert, err = x509.ParseCertificate(pemBlock.Bytes)
			if err != nil {
				return fmt.Errorf("invalid CA certification in Rancher configuration, %w", err)
			}

			rootCAs.AddCert(ownCert)

			req.Header.Add("Authorization", "Bearer "+rancherServerConfig.TokenKey)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: rootCAs,
				},
			}
			httpClient := &http.Client{Transport: tr}

			logrus.Info("Uploading image file ...")
			var resp *http.Response
			resp, err = httpClient.Do(req)

			if err != nil {
				return err
			}

			if resp.Status == "200 OK" {
				logrus.Info("Successfully uploaded the image file! DONE!")
				return nil
			} else {
				return fmt.Errorf("uploading image file to harvester was not successful: %s", resp.Body)
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

func imageCatalog(ctx *cli.Context) (err error) {

	metadataUrl := ctx.String("metadata-url")

	logrus.Debug("current metadata url: " + metadataUrl)

	var resp *http.Response
	resp, err = http.Get(metadataUrl)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	var body []byte
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var catalog Catalog
	err = json.Unmarshal(body, &catalog)

	if err != nil {
		return
	}

	writer := rcmd.NewTableWriter([][]string{
		{"NUMBER", "Id"},
		{"NAME", "Name"},
		{"NUMBER OF IMAGES", "NumberOfImages"},
	},
		ctx)

	osChoiceMap := make(map[int64]string)
	var i int64 = 0

	for os, imageList := range catalog.HarvesterImageCatalog {
		i++
		number := int64(len(imageList))
		writer.Write(&Os{
			Id:             i,
			Name:           os,
			NumberOfImages: strconv.FormatInt(number, 10),
		})
		osChoiceMap[i] = os

	}

	writer.Close()

	fmt.Println("Insert a number to select the image OS: ")
	reader := bufio.NewReader(os.Stdin)
	selection, err := GetSelectionFromInput(reader, len(osChoiceMap))
	if err != nil {
		return err
	}

	osSelection := osChoiceMap[int64(selection)]

	fmt.Printf("\nHere are the images available for %s\n\n", osSelection)

	writer = rcmd.NewTableWriter([][]string{
		{"NUMBER", "Id"},
		{"NAME", "ShortName"},
		{"VERSION", "Version"},
		{"BUILD", "Build"},
		{"URL", "Url"},
	}, ctx)

	imageChoiceMap := make(map[int64]string)

	for i, catalogItem := range catalog.HarvesterImageCatalog[osSelection] {
		catalogItem.Id = int64(i) + 1
		writer.Write(catalogItem)
		imageChoiceMap[catalogItem.Id] = catalogItem.Url
	}

	writer.Close()

	fmt.Printf("\nInsert a number to select an image to download: \n")
	selection, err = GetSelectionFromInput(reader, len(imageChoiceMap))
	if err != nil {
		return err
	}

	imageUrl := imageChoiceMap[int64(selection)]
	fmt.Printf("\nYour image URL is : %s\n", imageUrl)
	imageUrlObject, err := url.Parse(imageUrl)

	if err != nil {
		return fmt.Errorf("the url parsed from the metadata file is invalid, %w", err)
	}

	urlPathComponents := strings.Split(imageUrlObject.EscapedPath(), "/")
	imageFilename := urlPathComponents[len(urlPathComponents)-1]

	imageCreatedName, err := createImageObjectInAPI(ctx, imageFilename, "download", imageUrl)
	if err != nil {
		return fmt.Errorf("error during creation of image in Harvester %w", err)
	}

	logrus.Infof("Image was created in Harvester with display name %s and id %s", imageFilename, imageCreatedName)

	return nil
}
