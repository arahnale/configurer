package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type hostsT struct {
	Host            string `yaml:"host"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	pillars         []string
	pillarVariables interface{}
}

// Структура для хранения данных из YAML файла
type topPillarT struct {
	Pillars map[string][]string `yaml:",inline"`
}

type PillarT struct {
	Pillar []map[string]interface{} `json:"mappings,omitempty"`
}

func ReadHosts() (map[string]hostsT, error) {
	filename := "hosts.yaml"

	// Чтение содержимого YAML-файла
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Не удалось прочитать файл: %v", err)
	}

	// Декодирование YAML-данных в структуру
	var servers map[string]hostsT

	err = yaml.Unmarshal(yamlFile, &servers)
	if err != nil {
		log.Fatalf("Ошибка разбора YAML: %v", err)
	}

	// мердж с default
	for serverName, _ := range servers {
		if serverName == "default" {
			continue
		}

		dataServer := servers[serverName]

		if dataServer.User == "" {
			dataServer.User = servers["default"].User
		}
		if dataServer.Password == "" {
			dataServer.Password = servers["default"].Password
		}

		servers[serverName] = dataServer
	}

	return servers, nil
}

func PrintHosts(servers *map[string]hostsT) {
	// Вывод результатов после мерджа с default
	for serverName, server := range *servers {
		fmt.Printf("%s:\n", serverName)
		fmt.Printf("\thost: '%s'\n", server.Host)
		fmt.Printf("\tuser: '%s'\n", server.User)
		fmt.Printf("\tpassword: '%s'\n", server.Password)
	}
}

func ReadTop() (topPillarT, error) {
	filename := "pillar/top.yaml"
	var servers topPillarT

	// Чтение содержимого YAML-файла
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Не удалось прочитать файл: %v", err)
		return servers, err
	}

	err = yaml.Unmarshal(yamlFile, &servers)
	if err != nil {
		log.Fatalf("Ошибка разбора YAML: %v", err)
		return servers, err
	}

	// var Pillars PillarT

	// по уму надо прочитать теперь ссылки что указыны в pillar
	fmt.Println("\nServers:")
	for server, commands := range servers.Pillars {
		fmt.Println("server " + server + ":")

		for _, command := range commands {
			fmt.Println("-", command)

			// читаю файл с пиллар
			filename := "./pillar/" + command + "/" + "init.yaml"
			fmt.Println("pillar file " + filename)
			yamlFile, err := os.ReadFile(filename)
			if err != nil {
				continue
			}

			var data interface{}
			err = yaml.Unmarshal(yamlFile, &data)
			if err != nil {
				log.Fatalf("Ошибка разбора YAML: %v", err)
				return servers, nil
			}

			// Pillars[server] := data

			fmt.Println("printData")
			printData(data)
		}
	}

	return servers, nil
}

func printData(data interface{}) {
	switch v := data.(type) {
	case map[interface{}]interface{}:
		for key, value := range v {
			fmt.Printf("%s:\n", key)
			printData(value)
		}
	case []interface{}:
		for i, item := range v {
			fmt.Printf("[%d]: ", i+1)
			printData(item)
		}
	default:
		fmt.Println(v)
	}
}

func main() {
	servers, _ := ReadHosts()

	PrintHosts(&servers)

	pillars, _ := ReadTop()
	_ = pillars

}
