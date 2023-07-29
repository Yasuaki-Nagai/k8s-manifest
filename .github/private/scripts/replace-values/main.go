package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

type Secret struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

func main() {
	var (
		manifestsDir = flag.String("manifestsDir", "../../../manifests", "specify helm app contained directory")
		secretsDir   = flag.String("secretsDir", "../../../secrets", "specify secrets.yaml contained directory")
	)

	flag.Parse()
	helmApps, _ := os.ReadDir(*manifestsDir)

	for _, app := range helmApps {
		valuesFilePath := *manifestsDir + "/" + app.Name() + "/values.yaml"
		valuesData, valuesDataErr := os.ReadFile(valuesFilePath)
		if valuesDataErr != nil {
			fmt.Println("[ERROR] Can not read values.yaml, PATH: "+valuesFilePath, valuesDataErr)
			return
		}

		secretsFilePath := *secretsDir + "/" + app.Name() + "/secrets.yaml"
		secretsData, secretsDataErr := os.ReadFile(secretsFilePath)
		if secretsDataErr != nil {
			fmt.Println("[ERROR] MESSAGE: Can not read secrets.yaml, PATH: "+secretsFilePath, secretsDataErr)
			return
		}

		var secrets []Secret
		parseErr := yaml.Unmarshal(secretsData, &secrets)
		if parseErr != nil {
			fmt.Println("[ERROR] Failed to unmarshal secrets.yaml", parseErr)
			return
		}

		for _, secret := range secrets {
			valuesData = []byte(strings.ReplaceAll(string(valuesData), secret.Key, secret.Value))
		}

		fileWriteErr := os.WriteFile(valuesFilePath, valuesData, 0644)
		if fileWriteErr != nil {
			fmt.Println("[ERROR] Failed to merge secrets.yaml files", fileWriteErr)
			return
		}

		fmt.Println("[INFO] Successfully merged secrets.yaml files")
	}
}
