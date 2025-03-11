package config

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"dario.cat/mergo"
	"gopkg.in/yaml.v3"
)

func GetHostVariables(host *Host) error {
	log.Printf("[INFO] Starting variable processing for host '%s'", host.Name)
	log.Printf("[DEBUG] Host details: %+v", host)

	renderedTop, err := renderTopTemplate(host)
	if err != nil {
		return fmt.Errorf("error rendering top template: %v", err)
	}
	log.Printf("[DEBUG] Rendered top.yaml content:\n%s", string(renderedTop))

	var rawSections []map[interface{}]interface{}
	if err := yaml.Unmarshal(renderedTop, &rawSections); err != nil {
		return fmt.Errorf("error parsing top.yaml: %v", err)
	}

	var paths []string
	for i, sec := range rawSections {
		log.Printf("[DEBUG] Processing section %d: %+v", i+1, sec)
		for key, value := range sec {
			condition := fmt.Sprintf("%v", key)
			log.Printf("[INFO] Evaluating condition: %s", condition)

			items, ok := value.([]interface{})
			if !ok {
				log.Printf("[WARN] Invalid format for condition %s, skipping", condition)
				continue
			}

			applicable := false
			parts := strings.SplitN(condition, ":", 2)
			conditionType := parts[0]
			conditionValue := ""
			if len(parts) > 1 {
				conditionValue = parts[1]
			}

			switch conditionType {
			case "default":
				log.Printf("[INFO] Applying default section")
				applicable = true
			case "name":
				applicable = host.Name == conditionValue
				log.Printf("[INFO] Name condition '%s' matched: %t", conditionValue, applicable)
			case "group":
				applicable = contains(host.Groups, conditionValue)
				log.Printf("[INFO] Group condition '%s' matched: %t (host groups: %v)",
					conditionValue, applicable, host.Groups)
			case "vars":
				_, applicable = host.Vars[conditionValue]
				log.Printf("[INFO] Var condition '%s' exists: %t (host vars: %v)",
					conditionValue, applicable, host.Vars)
			default:
				log.Printf("[WARN] Unknown condition type: %s", conditionType)
			}

			if applicable {
				log.Printf("[INFO] Condition '%s' is applicable, adding %d items",
					condition, len(items))
				for _, item := range items {
					path := fmt.Sprintf("%v", item)
					paths = append(paths, path)
				}
			}
		}
	}

	log.Printf("[INFO] Collected %d variable paths to process: %v", len(paths), paths)
	variables := make(map[string]interface{})

	for _, path := range paths {
		log.Printf("[INFO] Processing variable path: %s", path)
		filePath := strings.ReplaceAll(path, ".", "/")
		data, err := loadFile(host.Env, filePath)
		if err != nil {
			log.Printf("[WARN] Failed to load %s: %v", filePath, err)
			continue
		}

		log.Printf("[DEBUG] Merging variables from %s: %+v", path, data)
		if err := mergo.Merge(&variables, data, mergo.WithOverride); err != nil {
			log.Printf("[ERROR] Failed to merge variables from %s: %v", path, err)
			continue
		}
		log.Printf("[INFO] Successfully merged variables from %s", path)
	}

	host.Variables = variables
	log.Printf("[INFO] Completed variable processing for host '%s'", host.Name)
	log.Printf("[DEBUG] Final variables: %+v", variables)
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func renderTopTemplate(host *Host) ([]byte, error) {
	topPath := filepath.Join(host.Env, "variables", "top.yaml")
	content, err := os.ReadFile(topPath)
	if err != nil {
		return nil, fmt.Errorf("error reading top.yaml: %v", err)
	}

	tmpl, err := template.New("top").Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("error parsing top template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, host); err != nil {
		return nil, fmt.Errorf("error executing top template: %v", err)
	}

	return buf.Bytes(), nil
}

func loadFile(baseDir string, path string) (map[string]interface{}, error) {
	log.Printf("[DEBUG] Attempting to load variables for path: %s", path)
	yamlPath := filepath.Join(baseDir, "variables", path+".yaml")

	if _, err := os.Stat(yamlPath); err == nil {
		log.Printf("[INFO] Found direct YAML file: %s", yamlPath)
		return loadAndProcessFile(baseDir, yamlPath)
	}

	initPath := filepath.Join(baseDir, "variables", path, "init.yaml")
	if _, err := os.Stat(initPath); err == nil {
		log.Printf("[INFO] Found init YAML file: %s", initPath)
		return loadAndProcessFile(baseDir, initPath)
	}

	log.Printf("[WARN] No files found for path: %s (tried: %s and %s)",
		path, yamlPath, initPath)
	return nil, fmt.Errorf("file not found for path %s", path)
}

func loadAndProcessFile(baseDir, filePath string) (map[string]interface{}, error) {
	log.Printf("[INFO] Loading file: %s", filePath)
	data := make(map[string]interface{})

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	if err := yaml.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("error unmarshaling YAML: %w", err)
	}

	if includes, ok := data["include"].([]interface{}); ok {
		log.Printf("[INFO] Processing includes in %s: %v", filePath, includes)
		delete(data, "include")

		for _, inc := range includes {
			incPath, ok := inc.(string)
			if !ok {
				log.Printf("[WARN] Invalid include format in %s: %v", filePath, inc)
				continue
			}

			incPath = strings.ReplaceAll(incPath, ".", "/")
			log.Printf("[INFO] Processing include: %s", incPath)

			currentDir := filepath.Dir(filePath)
			variablesDir := filepath.Join(baseDir, "variables")
			relPath, err := filepath.Rel(variablesDir, currentDir)
			if err != nil {
				return nil, fmt.Errorf("error calculating relative path: %w", err)
			}

			fullIncPath := filepath.Join(relPath, incPath)
			log.Printf("[DEBUG] Resolved include path: %s", fullIncPath)

			includedData, err := loadFile(baseDir, fullIncPath)
			if err != nil {
				log.Printf("[WARN] Failed to load include %s: %v", fullIncPath, err)
				continue
			}

			if err := mergo.Merge(&data, includedData, mergo.WithOverride); err != nil {
				log.Printf("[ERROR] Failed to merge include %s: %v", fullIncPath, err)
				continue
			}
			log.Printf("[INFO] Successfully merged include: %s", fullIncPath)
		}
	}

	return data, nil
}
