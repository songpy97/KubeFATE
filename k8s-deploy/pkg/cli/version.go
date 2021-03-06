package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/FederatedAI/KubeFATE/k8s-deploy/pkg/api"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"net/http"
)

func VersionCommand() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Flags: []cli.Flag{},
		Usage: "Show kubefate command line and service version",
		Action: func(c *cli.Context) error {
			serviceVersion, err := GetServiceVersion()
			if err != nil {
				fmt.Printf("* kubefate service connection error, %s\r\n", err.Error())
			} else {
				fmt.Printf("* kubefate service version=%s\r\n", serviceVersion)
			}
			fmt.Printf("* kubefate commandLine version=%s\r\n", api.ServiceVersion)
			return nil
		},
	}
}

func GetServiceVersion() (string, error) {
	r := &Request{
		Type: "GET",
		Path: "version",
		Body: nil,
	}

	serviceUrl := viper.GetString("serviceurl")
	apiVersion := api.ApiVersion + "/"
	if serviceUrl == "" {
		serviceUrl = "localhost:8080/"
	}
	Url := "http://" + serviceUrl + "/" + apiVersion + r.Path
	body := bytes.NewReader(r.Body)
	log.Debug().Str("Type", r.Type).Str("url", Url).Msg("Request")
	request, err := http.NewRequest(r.Type, Url, body)
	if err != nil {
		return "", err
	}

	token, err := getToken()
	if err != nil {
		return "", err
	}
	Authorization := fmt.Sprintf("Bearer %s", token)

	request.Header.Add("Authorization", Authorization)

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	type VersionResultMsg struct {
		Msg     string
		Version string
	}

	VersionResult := new(VersionResultMsg)

	err = json.Unmarshal(respBody, &VersionResult)
	if err != nil {
		return "", err
	}

	log.Debug().Int("Code", resp.StatusCode).Bytes("Body", respBody).Msg("ok")
	return VersionResult.Version, err
}
