package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/viper"
	"github.com/variantdev/vals"
)

const fileRefVariable = "VALS_FILE"
const testVariable = "VALS_TEST"
const cacheSize = 256

// Usage
// VALS_NAME1=[secretref] VALS_NAME2=[secretref] VALS_FILE=[/path/to/file]:[secretref] vals-entrypoint [command]

func init() {
	viper.AutomaticEnv()
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {

	val, err := vals.New(vals.Options{CacheSize: cacheSize})
	if err != nil {
		return fmt.Errorf("failed to initialize vals: %s", err)
	}

	fileRefs, _ := getFileRefs(viper.GetStringSlice("vals_file"))
	renderedFileRefs, err := val.Eval(map[string]interface{}{
		"inline": fileRefs,
	})

	if err != nil {
		return fmt.Errorf("failed to render file refs: %s", err)
	}

	envRefs, err := getEnvRefs(findVariables())
	if err != nil {
		return fmt.Errorf("failed to locate env refs: %w", err)
	}

	renderedEnvRefs, err := val.Eval(map[string]interface{}{
		"inline": envRefs,
	})

	if err != nil {
		return fmt.Errorf("failed to render env refs: %w", err)
	}

	envs := mapInterfaceToMapString(renderedEnvRefs["inline"].(map[string]interface{}))
	files := mapInterfaceToMapString(renderedFileRefs["inline"].(map[string]interface{}))

	if viper.GetBool(testVariable) {
		fmt.Printf("vars: %v\n\n", envs)
		fmt.Printf("files: %v\n\n", files)
		return nil
	}

	err = writeFiles(files)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return runExec(envs)
}

func runExec(env map[string]string) error {
	cmd := exec.Command(os.Args[1], os.Args[2:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Env = buildEnvVars(env)
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	return cmd.Wait()
}

func buildEnvVars(in map[string]string) []string {
	env := os.Environ()
	for k, v := range in {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}

func mapInterfaceToMapString(mapInterface map[string]interface{}) map[string]string {
	mapString := make(map[string]string)
	for key, value := range mapInterface {
		strKey := fmt.Sprintf("%v", key)
		strValue := fmt.Sprintf("%v", value)

		mapString[strKey] = strValue
	}
	return mapString
}

func writeFiles(files map[string]string) error {
	for k, v := range files {
		bytes := []byte(v)
		err := ioutil.WriteFile(k, bytes, 0600)
		if err != nil {
			return err
		}
	}

	return nil
}

func findVariables() []string {
	var keys []string

	vars := os.Environ()
	for _, envVar := range vars {
		split := strings.SplitN(envVar, "=", 2)
		keys = append(keys, split[0])
	}

	return keys
}

func getFileRefs(in []string) (map[string]interface{}, error) {
	refs := make(map[string]interface{})

	for _, s := range in {
		split := strings.SplitN(s, ":", 2)
		if split[0] == "" || split[1] == "" {
			return nil, errors.New("invalid file ref format")
		}

		refs[split[0]] = split[1]
	}

	return refs, nil
}

func getEnvRefs(in []string) (map[string]interface{}, error) {
	refs := make(map[string]interface{})

	for _, v := range in {
		if v == fileRefVariable {
			continue
		}
		refs[v] = viper.GetString(v)
	}

	return refs, nil
}
