package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type resource struct {
	identifier string
	envName    string
	filePath   string
}

func getEnvironmentResources(directory string, envName string, ext string) ([]resource, error) {
	res, err := getAllResources(directory, ext)

	if err != nil {
		return nil, fmt.Errorf("couldn't get %v resources from %v: %w", ext, directory, err)
	}

	return resolveResourcesForEnvironment(res, envName), nil
}

func getAllResources(directory string, ext string) ([]resource, error) {
	resources := make([]resource, 0)

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !hasExtension(info.Name(), ext) {
			return nil
		}

		identifier, env, err := parseResourceFileName(info.Name(), ext)

		if err != nil {
			println(fmt.Sprintf("unexpected file %v: %v", info.Name(), err))
			return nil
		}

		resources = append(resources, resource{identifier, env, path})

		return nil
	})

	if err != nil {
		if fileError := err.(*os.PathError); fileError.Err == syscall.ENOENT {
			return resources, nil
		}
	}

	return resources, err
}

func hasExtension(name string, ext string) bool {
	return strings.HasSuffix(strings.ToLower(name), fmt.Sprintf(".%v", ext))
}

func parseResourceFileName(name string, ext string) (identifier string, env string, err error) {
	if hasExtension(name, ext) {
		name = name[0 : len(name)-len(ext)-1]

		lastDashIdx := strings.LastIndex(name, "-")

		if lastDashIdx != -1 {
			identifier = name[0:lastDashIdx]
			env = name[lastDashIdx+1:]
			return
		}
	}

	err = errors.New(fmt.Sprintf("expected filename {identifier}-{environment}.%v", ext))
	return
}

func resolveResourcesForEnvironment(resources []resource, envName string) []resource {
	result := make([]resource, 0)

	byIdentifier := make(map[string][]resource)
	for _, m := range resources {
		mSlice := byIdentifier[m.identifier]
		if mSlice == nil {
			mSlice = make([]resource, 0)
		}
		mSlice = append(mSlice, m)
		byIdentifier[m.identifier] = mSlice
	}

	for _, mSlice := range byIdentifier {
		specific := -1
		dflt := -1

		for i, m := range mSlice {
			if m.envName == envName {
				specific = i
			} else if m.envName == "default" {
				dflt = i
			}
		}

		if specific != -1 {
			result = append(result, mSlice[specific])
		} else if dflt != -1 {
			result = append(result, mSlice[dflt])
		}
	}

	return result
}
