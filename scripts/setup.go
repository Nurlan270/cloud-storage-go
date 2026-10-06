package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

func main() {
	envFlag := flag.String("env", "local", "environment: local or prod")
	flag.Parse()

	env := *envFlag

	if env != "local" && env != "prod" {
		panic(fmt.Sprintf("unsupported environment: %s", env))
	}

	if err := createEnvFile(env); err != nil {
		panic(err)
	}

	if err := createConfigFile(env); err != nil {
		panic(err)
	}
}

func createEnvFile(env string) error {
	const (
		source      = ".env.example"
		destination = ".env"
	)

	//	Don't overwrite existing .env.
	if _, err := os.Stat(destination); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("check %s: %w", destination, err)
	}

	data, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("read %s: %w", source, err)
	}

	re := regexp.MustCompile(`(?m)^APP_ENV=.*$`)
	content := re.ReplaceAllString(string(data), "APP_ENV="+env)

	if err = os.WriteFile(destination, []byte(content), 0644); err != nil {
		return fmt.Errorf("write %s: %w", destination, err)
	}

	return nil
}

func createConfigFile(env string) error {
	source := filepath.Join("config", "local.yml")
	destination := filepath.Join("config", env+".yml")

	//	Don't overwrite existing config.
	if _, err := os.Stat(destination); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("check %s: %w", destination, err)
	}

	data, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("read %s: %w", source, err)
	}

	if err = os.WriteFile(destination, data, 0644); err != nil {
		return fmt.Errorf("write %s: %w", destination, err)
	}

	return nil
}
