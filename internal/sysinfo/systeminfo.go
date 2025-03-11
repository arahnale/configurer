package sysinfo

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"path/filepath"
	"time"

	"configurer/internal/ssh"
)

type SysInfoT struct {
	OsVersion  string `yaml:"os_version" json:"os_version"`
	MemorySize string `yaml:"memory_size" json:"memory_size"`
}

// Функция для генерации случайного имени файла
func generateRandomFileName() string {
	src := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(src)
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 10)
	for i := range b {
		b[i] = letters[rng.Intn(len(letters))]
	}
	return string(b)
}

func SysInfo(host string, port string, user string, password string, privateKeyPath string) (SysInfoT, error) {
	fmt.Println("start SysInfo")
	var sysInfo SysInfoT

	// Конфигурация SSH-клиента
	client, err := ssh.ConnectSSH(host, port, user, password, privateKeyPath)
	if err != nil {
		fmt.Println(err)
		return sysInfo, err
	}
	defer client.Close()

	// Копирование файла на сервер
	localFilePath := "./remote" // Локальный путь к файлу
	remoteDir := "/tmp"         // Директория на сервере

	randomFileName := generateRandomFileName()
	remoteFilePath := filepath.Join(remoteDir, randomFileName)

	err = ssh.CopyFileToServer(client, localFilePath, remoteFilePath)
	if err != nil {
		log.Fatalf("Ошибка при копировании файла: %v", err)
		return sysInfo, err
	}
	fmt.Printf("Файл скопирован на сервер: %s\n", remoteFilePath)

	output, err := ssh.RemoteCommand(client, remoteFilePath+" --getinfo=yes")
	if err != nil {
		log.Fatalf("Ошибка запуска скрипта для получения информации о сервере: %v", err)
		return sysInfo, err
	}

	// удаление
	_, err = ssh.RemoteCommand(client, "rm "+remoteFilePath)
	if err != nil {
		log.Fatalf("Ошибка удаления файла получения конфигурации сервера: %v", err)
		return sysInfo, err
	}

	// Преобразование output в структуру SysInfoT
	err = json.Unmarshal(output, &sysInfo)
	if err != nil {
		log.Fatalf("Ошибка декодирования JSON: %v", err)
		return sysInfo, err
	}

	// Вывод данных
	fmt.Println("Данные с сервера:")
	fmt.Printf("OS Version: %s\n", sysInfo.OsVersion)
	fmt.Printf("Memory Size: %s\n", sysInfo.MemorySize)

	return sysInfo, nil
}
