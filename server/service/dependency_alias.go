package service

import "strings"

var pythonModulePackageAliases = map[string]string{
	"crypto": "pycryptodome",
}

func ResolvePythonAutoInstallPackage(moduleName string) string {
	moduleName = strings.TrimSpace(moduleName)
	if moduleName == "" {
		return ""
	}

	if mapped, exists := pythonModulePackageAliases[strings.ToLower(moduleName)]; exists {
		return mapped
	}

	return moduleName
}
