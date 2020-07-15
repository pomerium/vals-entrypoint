package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/variantdev/vals"
)

const cacheSize = 256
const fileRefVariable = "VALS_FILES"

var rootCmd = &cobra.Command{
	Use:  "vals-entrypoint",
	Long: "Bootstrap environment variables and config files from secrets engines supported by `val`",
}

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "execute the following command with variables interpollated and files generated",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return run(args)
	},
}

func init() {
	rootCmd.PersistentFlags().String("vals-files", "", "specify comma separated list of file(s) to write a secret into.  format: [/path/to/file.yaml:ref+gcpsecrets://my-project-id/test]")
	rootCmd.PersistentFlags().Bool("test", false, "run in test mode and output results of interpollation")

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		panic(err)
	}

	if err := viper.BindEnv("vals-files", fileRefVariable); err != nil {
		panic(err)
	}

	if err := viper.BindPFlag("test", rootCmd.PersistentFlags().Lookup("test")); err != nil {
		panic(err)
	}

	viper.AutomaticEnv()
	rootCmd.AddCommand(execCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(command []string) error {
	valsfiles := viper.GetStringSlice("vals-files")

	val, err := vals.New(vals.Options{CacheSize: cacheSize})
	if err != nil {
		return fmt.Errorf("failed to initialize vals: %s", err)
	}

	files, err := renderFiles(val, valsfiles)
	if err != nil {
		return fmt.Errorf("failed to render files: %w", err)
	}

	envs, err := renderVars(val)
	if err != nil {
		return fmt.Errorf("failed to render vars: %w", err)
	}

	if viper.GetBool("test") {
		fmt.Printf("vars: %v\n\n", envs)
		fmt.Printf("files: %v\n\n", files)
		return nil
	}

	if len(valsfiles) > 0 {
		err = writeFiles(files)
		if err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	}

	return runExec(command, envs)
}

func renderFiles(val *vals.Runtime, in []string) (map[string]string, error) {
	fileRefs, _ := getFileRefs(in)
	renderedFileRefs, err := val.Eval(map[string]interface{}{
		"inline": fileRefs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to render file refs: %s", err)
	}

	return mapInterfaceToMapString(renderedFileRefs["inline"].(map[string]interface{})), nil
}

func renderVars(val *vals.Runtime) (map[string]string, error) {
	envRefs, err := getEnvRefs(findVariables())
	if err != nil {
		return nil, fmt.Errorf("failed to locate env refs: %w", err)
	}

	renderedEnvRefs, err := val.Eval(map[string]interface{}{
		"inline": envRefs,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to render env refs: %w", err)
	}

	return mapInterfaceToMapString(renderedEnvRefs["inline"].(map[string]interface{})), nil
}

func runExec(command []string, env map[string]string) error {
	cmd := exec.Command(command[0], command[1:]...)
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
		refs[v] = os.Getenv(v)
	}

	return refs, nil
}
