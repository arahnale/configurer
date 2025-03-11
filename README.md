# Структура:
ROSTER - список серверов которыми требуется управлять

VARIABLES - переменные, типа pillar в salt

STATES - описание действий которые требуется применить к серверу на базе VARIABLES


---

# ROSTER

у меня есть файл ./roster/hosts.yaml
```
default: # группа по умолчанию, параметры отсюда должны наследоваться на все группы
  hosts: # список серверов
    server1: # имя сервера
      host: 192.168.0.11
      password: 123 # пароль от сервера
    server2:
      password: 312 # пароль от сервера
      host: 192.168.0.12
      user: ubuntu
  password: pass
  ssh_key: ./ssh/id_rsa # путь к ключу ssh
  env: master
  user: root # если у сервера не определен пользователь, то по умолчанию будет использоваться пользователь root
  vars: # свободная структура данных, тут могут быть любые json структуры 
    timezone: UTC+3
group1:
  hosts:
    server1:
  vars:
    web: ["php-fpm","nginx"]
group2:
  hosts:
    server2:
  vars:
    web: ["apache2"]
group3:
  hosts:
    server3:
      host: 192.168.0.13
      ssh_key: ./id_rsa3
```
напиши программу на golang для того, чтобы получить список хостов.
все параметры из группы default должны наследоваться на все хосты
все параметры группы в которой находится хост, должны наследоваться на этот хост
если в хосте есть свои параметры, то они не должны стираться
поле host - обязательное, если оно не определено, то требуется выдать ошибку
ssh_key - если не определен, то по умолчанию имеет значение ./ssh/id_rsa
env - если не определен, то по умолчанию имеет значение master 
user - если не определен, то по умолчанию имеет значение root
имя основной функции должно быть GetRosterHosts которая возвращает структуру, в ней же должно происходить чтение файла
в итоге я должен получить yaml
```
server1: # имя сервера
  user: root
  password: 123
  ssh_key: id_rsa
  vars:
    timezone: UTC+3
    web: ["php-fpm","nginx"]
  groups: [default, group1]
  variables:
  sysinfo:
  env: master
server2: # имя сервера
  user: ubuntu
  password: 321
  ssh_key: id_rsa
  vars:
    timezone: UTC+3
    web: ["apache2"]
  groups: [default, group2]
  variables:
  sysinfo:
  env: master
server3: # имя сервера
  user: root
  password: 321
  ssh_key: ./id_rsa3
  vars:
    timezone: UTC+3
  groups: [default, group3]
  variables:
  sysinfo:
  env: master
```


---

# VARIABLES

переменные для хоста

Структура
```
type Host struct {
	Name      string                 `yaml:"name"`
	Address   string                 `yaml:"address"`
	User      string                 `yaml:"user,omitempty"`
	Port      string                 `yaml:"port,omitempty"`
	Password  string                 `yaml:"password,omitempty"`
	SSHKey    string                 `yaml:"ssh_key,omitempty"`
	Vars      map[string]interface{} `yaml:"vars,omitempty"`
	Groups    []string               `yaml:"groups,omitempty"`
	SysInfo   map[string]interface{} `yaml:"sysinfo,omitempty"`
	Variables map[string]interface{} `yaml:"variables,omitempty"`
	States    []map[string]interface{} `yaml:"states,omitempty"`
	Env       string                 `yaml:"env,omitempty"`
}
```

у меня есть структура файлов, требуется прочитать файл {Host.Env}/variables/top.yaml
```
- default: # должно наследоваться на все хосты
  - service.ssh
  - config.bash
  - repos.aptly

- name:2204: # host означает что требуется применять только к серверу с именем 2204
  - service.files # требуется прочитать файл в директории ${Host.Env}/variables/service/files/init.yaml, если файл ${Host.Env}/variables/service/files.yaml не найден
  - config.docker # требуется прочитать файл в директории ${Host.Env}/variables/config/docker.yaml, если файл docker.yaml найден, то искать файл ${Host.Env}/variables/config/docker/init.yaml не требуется

- group: # group означает что требуется производить поиск по группе сервера
  {% for group in Host.groups %}
  - groups.{{ group }}
  {% endfor %}

- vars:timezone: # если в Host присутствует vars['timezone'] , то требуется читать файл
  - sevice.timezone

- vars:profile:
  - profiles.{{ vars.profile }}

- host:
  - hosts.{{ name }}
```

`- default:` применяется по умолчанию

`- service.ssh` требуется заменить точки на слеши и прочитать файл ${Env}/variables/service/ssh.yaml , если этот файл не будет найден, то ${Env}/variables/service/ssh/init.yaml

`- name:2204:` - означает что надо применять для хоста с именем 2204

`- group:web:` - означает что требуется применить ко всем хостам, у которых есть группа `web` в Host.Groups

`- vars:timezone:` - означает что требуется применить к хостам, у которых есть Host.Vars['timezone']

шаблонизируй содержимое файлов через text/template


Содержимое master/variables/service/ssh/init.yaml
```
ssh:
  __state__: True
  users:
    - admin
    - root
```

Содержимое master/variables/hosts/2404/init.yaml
```
include:
  - service.nginx
```

Содержимое master/variables/config/bash/init.yaml
```
bash:
  __state__: True
```

Содержимое master/variables/repos/aptly/init.yaml
```
repo:
  __state__: True
  repos_list:
    aptly:
      url: http://qweqew.qweqwe stable deb
      key: http://qweqwe.qwqwe/key
```

Содержимое master/variables/service/apache/init.yaml
```
apache:
  __state__: True
  config:
    workers: 123
```

Содержимое master/variables/service/nginx/init.yaml
```
nginx:
  __state__: True
  config:
    workers: 312
logrotate:
  config:
    nginx.conf: |
      config log rotate
```

Содержимое master/variables/service/php/init.yaml
```
php:
  __state__: True
  config:
    pool: 123
```

Содержимое master/variables/service/files/init.yaml
```
files:
  __state__: True
```

Содержимое master/variables/service/dokcer/init.yaml
```
include: # поддержка вложенных файлов, требуется прочитать указанные файлы
  - ubuntu2404

dokcer:
  __state__: True
  config: |
    abracadbra
    sim salabim
```

Содержимое master/variables/service/dokcer/ubuntu1804.yaml
```
repo:
  __state__: True
  repos_list:
    docker: http://docker.com stable bionic main
```

Содержимое master/variables/service/dokcer/ubuntu2204.yaml
```
repo:
  __state__: True
  repos_list:
    docker: http://docker.com stable jammy main
```

Содержимое master/variables/service/dokcer/ubuntu2404.yaml
```
repo:
  __state__: True
  repos_list:
    docker: http://docker.com stable noble main
```

Содержимое master/variables/service/dokcer/default.yaml
```
repo:
  __state__: True
  repos_list:
    docker: http://docker.com stable ubuntu main
```

Напиши функцию на golang, которая принимаем в себя ссылку на структуру Host c именем хоста и наполняет ее Host.Variables. если файл отсутствует, то требуется выдать предупреждение и продолжить выполнение. передавая хост 2404 и ссылку на его структуру, на выходе я хочу получить структуру в виде yaml:
```
2404:
  variables:
    ssh:
      __state__: True
      users:
        - admin
        - root
    bash:
      __state__: True
    files:
      __state__: True
    repo:
      __state__: True
      repos_list:
        aptly:
          url: http://qweqew.qweqwe stable deb
          key: http://qweqwe.qwqwe/key
    nginx:
      __state__: True
      config:
        workers: 312
    logrotate:
      config:
        nginx.conf: |
          config log rotate
```


---

# STATES

Требуется написать функцию GetHostStates на golang, которая будет принимать в себя указатель на Host и изменять его. Структура Host в ней содержится инфомация о сервере. 
```
type Host struct {
	Name      string                 `yaml:"name"`
	Address   string                 `yaml:"address"`
	User      string                 `yaml:"user,omitempty"`
	Port      string                 `yaml:"port,omitempty"`
	Password  string                 `yaml:"password,omitempty"`
	SSHKey    string                 `yaml:"ssh_key,omitempty"`
	Vars      map[string]interface{} `yaml:"vars,omitempty"`
	Groups    []string               `yaml:"groups,omitempty"`
	SysInfo   sysinfo.SysInfoT       `yaml:"sysinfo,omitempty"`
	Variables map[string]interface{} `yaml:"variables,omitempty"`
	States    []map[string]interface{} `yaml:"states,omitempty"`
	Env       string                 `yaml:"env,omitempty"`
}
```

Ко всем файлам применять шаблонизатор

В Variables содержится информация о стейтах которые требуется применить. в файле {{ Env }}/states/top.yaml содержится информация о порядке применения о списке стейтов и порядке их применения
```
- timezone:
  - services.timezone
- bash:
  - config.bash
- syslog-ng:
  - services.syslog-ng
```

если в Host.Variables у сервера присутствует map ["timezone"]["__state__"] в значении True, то требуется обработать
```
- timezone:
  - services.timezone
```
в services.timezone заменить точки на слеши, то есть преобразовать services.timezone в путь {{ Env }}/states/services/timezone.yaml, если он отсутствует то требуется прочитать файл {{ Env }}/states/services/timezone/init.yaml
по аналогии надо обработать и другие элементы массива

если Host.Variables у сервера присутствует map ["timezone"]["__state__"] в значении False или вовсе отсутствует, то обрабатывать `- timezone:` и его содержимое не требуется

содержимое {{ Env }}/states/config/bash/init.yaml
```
- include:
  - config.bash.config

- file:
    path: /root/.bashrc
    sources: 
      - {{ tmpl }}/files/default/bashrc
      - {{ tmpl }}/files/{{ host.Name }}/bashrc
```

содержимое {{ Env }}/states/config/bash/config.yaml
```
- cmd:
    run: user set bash
```

если встречаем `- include:` , то требуется прочитать файл, который там указан, например config.bash.config заменяем точки на слеши, получаем путь {{ Env }}/states/config/bash/config.yaml, если он не найден, то {{ Env }}/states/config/bash/config/init.yaml , если и он не найден, то просто продолжаем работу.

теперь наш Host должен выглядеть так
```
Host:
  States:
    - services.timezone:
      - services.timezone.init:
        - cmd: 
            run: timezone set Moskow
    - config.bash:
      - config.bash.config:
        - cmd:
            run: user set bash
      - config.bash.init:
        - cmd:
            run: timezone set Moskow
```

Если файл конфигурации отсутствует, то требуется выдать предупреждение и продолжить работу.
используй os вместо ioutil. используй "gopkg.in/yaml.v3" для работы с yaml
шаблонизируй содержимое файлов через text/template
добавь логирование, чтобы понимать что происходит

---

# ФОРМАТ СТЕЙТОВ

выполнение команды

```
- cmd:
    run: echo $ENV # выполнение команды
    ontrue_set: # в случае успеха присвоить значение переменной
    ontrue_cmd: # выполнить команду
    ontrue_state: # выполнить стейт
    onfalse_set: # в случае провала присвоить значение переменной 
    onfalse_cmd: 
    onfalse_state: 
    onequals: # сравнить возврат, если он равен, то можно выполнить onequals_cmd или onequals_state
    onequals_cmd:
    onequals_state:
    unless: # 
    if: # 
    if_var: need_state=abracadabra # стейт будет выполняться только в случае успеха проверки переменной из variables
    user: 
    shell:
    env:
    failhard_local: # сломаться и прекратить выполнение стейта, в случае ошибки
    failhard_global: # сломаться и прекратить выполнение стейта, в случае ошибки
```

управление файлами

```
- file:
    path: /root/.bashrc
    sources:
      - /files/default/bashrc
      - /files/host.Name/bashrc
    pathes: # может существовать только path или pathes
      /root/.bashrc:
        - /files/default/bashrc
      /root/.ssh/config:
        - /files/default/ssh/config
    onchange_state: services.ssh.restart
    onchange_set: need_ssh_reload=True
    user: root
    group: root
    mode: 0644
    template:
      local: local_env # передача локальных переменных в шаблон (вопрос только стоит ли это делать)
    check: nginx -t # команда для проверки и отката файлов в случае провала
```

управление директориями

```
- directory:
    path: /etc/nginx
    sources:
      - ./files/default/
      - ./files/groups/
      - ./files/hosts/
    onchange_state: services.nginx.restart # если изменился конфигурационный файл, то перезапустить сервис
    onchange_set: need_nginx_reload=True # если изменился конфигурационный файл, то установить переменную
    exclude: # исключить файлы или директории из директории
      - /etc/nginx/ssl
    merge: # политика слияния файлов. mirror - неизвестные файлы удаляются, overlay - неизвестные файлы не удаляются просто добавляются, пусть это будет политика по умолчанию.
    user: root
    group: root
    mode: 0755
    template:
      local: local_env
    if: dpkg -l nginx
    unless: ls /etc/nginx
    failhard_local: True
    failhard_global: True
    check: nginx -t # команда для проверки и отката файлов в случае провала
```

управление переменными (возможно потребуется для определения например состояние слейва или мастера)

```
- variable:
    name: need_nginx_reload
    value: True
    if: ls /etc/nginx
    unless: dpkg -l nginx
```

---

применение стейтов на сервере

взять переменные из states и применить их на сервере

