package main

import (
	"configurer/internal/config"
	"configurer/internal/sysinfo"
	"fmt"

	"gopkg.in/yaml.v2"
)

func main() {
	// Получаем хосты из roster/hosts.yaml
	hosts, err := config.GetRosterHosts("./roster/hosts.yaml")
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	// Получаем grains/sysinfo из серверов
	for hostname, host := range hosts {
		if hostname != "2404" {
			continue
		}

		// Подключаемся к хосту и получаем grains/sysinfo
		sysInfo, err := sysinfo.SysInfo(host.Address, host.Port, host.User, host.Password, host.SSHKey)
		if err != nil {
			fmt.Println(err)
			return
		}

		host.SysInfo = sysInfo
		hosts[hostname] = host

		if err := config.GetHostVariables(&host); err != nil {
			fmt.Println("Error GetHostVariables:", err)
			return
		}

		if err := config.GetHostStates(&host); err != nil {
			fmt.Printf("Error GetHostStates: %v\n", err)
			return
		}

		printHost(&host)
	}

	// for _, host := range hosts {
	// 	err := config.PopulateHostVariables(&host)
	// 	if err != nil {
	// 		fmt.Println("Error:", err)
	// 		return
	// 	}
	// 	printHost(&host)
	// }

	// printHosts(&hosts)

	return

	// // Получаем хосты из top.yaml
	// pillarHosts, err := config.GetTopHosts("./pillar/top.yaml", "base")
	// if err != nil {
	// 	fmt.Printf("Ошибка при получении хостов из top.yaml: %v\n", err)
	// 	return
	// }

	// // Объединяем хосты из roster/hosts.yaml и top.yaml
	// for hostname, host := range hosts {
	// 	if hostname != "2404" {
	// 		continue
	// 	}
	// 	// Проверяем, есть ли хост в pillarHosts
	// 	if pillarHost, exists := pillarHosts[hostname]; exists {
	// 		// Если хост найден, добавляем его Pillars в текущий хост
	// 		if host.Pillars == nil {
	// 			host.Pillars = make(map[string]interface{})
	// 		}
	// 		for key, value := range pillarHost.Pillars {
	// 			host.Pillars[key] = value
	// 		}
	// 		host.Env = "base"
	// 		err := config.LoadStates(&host, "./")
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 		hosts[hostname] = host
	// 	}
	// }

	// теперь надо выполнить states
	// требуется, прочитать файл top со стейтами
	// и составить список стейтов которые требуется применить
}

func printHosts(hosts *map[string]config.Host) {
	// Выводим результат в YAML
	output, err := yaml.Marshal(hosts)
	if err != nil {
		fmt.Printf("Ошибка при маршалинге YAML: %v\n", err)
	}
	fmt.Println(string(output))
}

func printHost(host *config.Host) {
	// Выводим результат в YAML
	output, err := yaml.Marshal(host)
	if err != nil {
		fmt.Printf("Ошибка при маршалинге YAML: %v\n", err)
	}
	fmt.Println(string(output))
}
