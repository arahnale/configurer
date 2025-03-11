package config

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

func GetHostStates(host *Host) error {
	if host.Env == "" {
		return fmt.Errorf("host env is not set")
	}

	topPath := filepath.Join(host.Env, "states", "top.yaml")
	topContent, err := os.ReadFile(topPath)
	if err != nil {
		return fmt.Errorf("failed to read top.yaml: %v", err)
	}

	var topData []map[string][]string
	if err := yaml.Unmarshal(topContent, &topData); err != nil {
		return fmt.Errorf("failed to parse top.yaml: %v", err)
	}

	for _, entry := range topData {
		for stateName, subStates := range entry {
			if isStateEnabled(host, stateName) {
				for _, subState := range subStates {
					steps := processState(host, subState)
					if len(steps) > 0 {
						host.States = append(host.States, map[string]interface{}{
							subState: steps,
						})
					}
				}
			}
		}
	}

	return nil
}

func isStateEnabled(host *Host, stateName string) bool {
	vars, ok := host.Variables[stateName]
	if !ok {
		return false
	}

	varsMap, ok := vars.(map[string]interface{})
	if !ok {
		return false
	}

	stateVal, ok := varsMap["__state__"]
	if !ok {
		return false
	}

	enabled, ok := stateVal.(bool)
	return ok && enabled
}

func processState(host *Host, statePath string) []interface{} {
	filePath, err := resolveStatePath(host.Env, statePath)
	if err != nil {
		log.Printf("Warning: %v", err)
		return nil
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Warning: failed to read state file: %v", err)
		return nil
	}

	rendered, err := renderTemplate(string(content), host)
	if err != nil {
		log.Printf("Template error: %v", err)
		return nil
	}

	var steps []interface{}
	if err := yaml.Unmarshal([]byte(rendered), &steps); err != nil {
		log.Printf("YAML parse error: %v", err)
		return nil
	}

	return processSteps(host, steps)
}

func resolveStatePath(env, state string) (string, error) {
	parts := strings.Split(state, ".")
	basePath := filepath.Join(append([]string{env, "states"}, parts...)...)

	// Try .yaml file first
	filePath := basePath + ".yaml"
	if _, err := os.Stat(filePath); err == nil {
		return filePath, nil
	}

	// Try init.yaml in directory
	dirPath := filepath.Join(append([]string{env, "states"}, parts...)...)
	initPath := filepath.Join(dirPath, "init.yaml")
	if _, err := os.Stat(initPath); err == nil {
		return initPath, nil
	}

	return "", fmt.Errorf("state path not found: %s", state)
}

func renderTemplate(content string, host *Host) (string, error) {
	tmpl, err := template.New("template").Parse(content)
	if err != nil {
		return "", err
	}

	data := struct {
		Host *Host
		Tmpl string
	}{
		Host: host,
		Tmpl: host.Env,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func processSteps(host *Host, steps []interface{}) []interface{} {
	var processed []interface{}
	for _, step := range steps {
		if stepMap, ok := step.(map[string]interface{}); ok {
			if includes, ok := stepMap["include"].([]interface{}); ok {
				for _, inc := range includes {
					if incStr, ok := inc.(string); ok {
						subSteps := processState(host, incStr)
						if len(subSteps) > 0 {
							processed = append(processed, map[string]interface{}{
								incStr: subSteps,
							})
						}
					}
				}
			} else {
				processed = append(processed, step)
			}
		} else {
			processed = append(processed, step)
		}
	}
	return processed
}
