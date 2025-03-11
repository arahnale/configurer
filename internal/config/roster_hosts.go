package config

// TODO: требуется добавить проверку, если в yaml отсутствует host - адрес сервера, то требуется использовать имя хоста как адрес

import (
	"os"

	"gopkg.in/yaml.v3"

	"configurer/internal/sysinfo"
)

type Host struct {
	Name      string                   `yaml:"name"`
	Address   string                   `yaml:"address"`
	User      string                   `yaml:"user,omitempty"`
	Port      string                   `yaml:"port,omitempty"`
	Password  string                   `yaml:"password,omitempty"`
	SSHKey    string                   `yaml:"ssh_key,omitempty"`
	Vars      map[string]interface{}   `yaml:"vars,omitempty"`
	Groups    []string                 `yaml:"groups,omitempty"`
	SysInfo   sysinfo.SysInfoT         `yaml:"sysinfo,omitempty"`
	Variables map[string]interface{}   `yaml:"variables,omitempty"`
	States    []map[string]interface{} `yaml:"states,omitempty"`
	Env       string                   `yaml:"env,omitempty"`
}

type Group struct {
	Hosts    map[string]Host        `yaml:"hosts,omitempty"`
	Vars     map[string]interface{} `yaml:"vars,omitempty"`
	User     string                 `yaml:"user,omitempty"`
	Password string                 `yaml:"password,omitempty"`
	Port     string                 `yaml:"port,omitempty"`
	SSHKey   string                 `yaml:"ssh_key,omitempty"`
	Env      string                 `yaml:"env,omitempty"`
}

type Roster struct {
	Default Group            `yaml:"default"`
	Groups  map[string]Group `yaml:",inline"`
}

func GetRosterHosts(filePath string) (map[string]Host, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var roster Roster
	err = yaml.Unmarshal(data, &roster)
	if err != nil {
		return nil, err
	}

	hosts := make(map[string]Host)

	// Обрабатываем хосты из группы default
	for hostName, host := range roster.Default.Hosts {
		finalHost := Host{
			Name:     hostName,
			Address:  host.Address,
			User:     roster.Default.User,
			Password: roster.Default.Password,
			Port:     roster.Default.Port,
			SSHKey:   roster.Default.SSHKey,
			Env:      roster.Default.Env,
			Vars:     make(map[string]interface{}),
			Groups:   []string{"default"},
		}

		// Применяем переменные из группы default
		for k, v := range roster.Default.Vars {
			finalHost.Vars[k] = v
		}

		// Применяем настройки хоста из группы default
		if host.Address != "" {
			finalHost.Address = host.Address
		}
		if host.User != "" {
			finalHost.User = host.User
		}
		if host.Password != "" {
			finalHost.Password = host.Password
		}
		if host.SSHKey != "" {
			finalHost.SSHKey = host.SSHKey
		}
		if host.Env != "" {
			finalHost.Env = host.Env
		}
		for k, v := range host.Vars {
			finalHost.Vars[k] = v
		}

		hosts[hostName] = finalHost
	}

	// Обрабатываем хосты из других групп
	for groupName, group := range roster.Groups {
		for hostName, host := range group.Hosts {
			finalHost, exists := hosts[hostName]
			if !exists {
				finalHost = Host{
					Name:     hostName,
					Address:  host.Address,
					User:     roster.Default.User,
					Password: roster.Default.Password,
					Port:     roster.Default.Port,
					SSHKey:   roster.Default.SSHKey,
					Env:      roster.Default.Env,
					Vars:     make(map[string]interface{}),
					Groups:   []string{"default", groupName},
				}

				// Применяем переменные из группы default
				for k, v := range roster.Default.Vars {
					finalHost.Vars[k] = v
				}
			} else {
				finalHost.Groups = append(finalHost.Groups, groupName)
			}

			// Применяем настройки группы
			if group.User != "" {
				finalHost.User = group.User
			}
			if group.Password != "" {
				finalHost.Password = group.Password
			}
			if group.Port != "" {
				finalHost.Port = group.Port
			}
			if group.SSHKey != "" {
				finalHost.SSHKey = group.SSHKey
			}
			if group.Env != "" {
				finalHost.Env = group.Env
			}
			for k, v := range group.Vars {
				finalHost.Vars[k] = v
			}

			// Применяем настройки хоста (переопределяют групповые и default)
			if host.Address != "" {
				finalHost.Address = host.Address
			}
			if host.User != "" {
				finalHost.User = host.User
			}
			if host.Password != "" {
				finalHost.Password = host.Password
			}
			if host.SSHKey != "" {
				finalHost.SSHKey = host.SSHKey
			}
			if host.Env != "" {
				finalHost.Env = host.Env
			}
			for k, v := range host.Vars {
				finalHost.Vars[k] = v
			}

			hosts[hostName] = finalHost
		}
	}

	return hosts, nil
}
