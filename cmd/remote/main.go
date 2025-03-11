package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"flag"

	"configurer/internal/sysinfo"
)

func getOSVersion() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "Ошибка при чтении /etc/os-release"
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			return strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), `"`)
		}
	}

	return "Не удалось определить версию ОС"
}

func getMemorySize() string {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return "Ошибка при чтении /proc/meminfo"
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return fmt.Sprintf("%s %s", parts[1], parts[2])
			}
		}
	}

	return "Не удалось определить размер оперативной памяти"
}

func main() {

	flagGetInfo := flag.String("getinfo", "", "Получение данных системы")

	flag.Parse()

	if *flagGetInfo != "" {
		sysInfo := sysinfo.SysInfoT{
			OsVersion:  getOSVersion(),
			MemorySize: getMemorySize(),
		}

		jsonData, err := json.Marshal(sysInfo)
		if err != nil {
			fmt.Print("{'error': '" + err.Error() + "'}")
			return
		}

		// Вывод данных в stdout
		fmt.Print(string(jsonData))
	}

}
