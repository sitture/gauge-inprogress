package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const (
	cgoEnabled        = "CGO_ENABLED"
	dotGauge          = ".gauge"
	plugins           = "plugins"
	GOARCH            = "GOARCH"
	GOOS              = "GOOS"
	X86               = "386"
	ARM64             = "arm64"
	x86_64            = "amd64"
	DARWIN            = "darwin"
	LINUX             = "linux"
	WINDOWS           = "windows"
	bin               = "bin"
	newDirPermissions = 0755
	gauge             = "gauge"
	inProgress        = "inprogress"
	deploy            = "deploy"
	pluginJSONFile    = "plugin.json"
)

var deployDir = filepath.Join(deploy, inProgress)

func isExecMode(mode os.FileMode) bool {
	return (mode & 0111) != 0
}

func mirrorFile(src, dst string) error {
	sfi, err := os.Stat(src)
	if err != nil {
		return err
	}
	if sfi.Mode()&os.ModeType != 0 {
		log.Fatalf("mirrorFile can't deal with non-regular file %s", src)
	}
	dfi, err := os.Stat(dst)
	if err == nil &&
		isExecMode(sfi.Mode()) == isExecMode(dfi.Mode()) &&
		(dfi.Mode()&os.ModeType == 0) &&
		dfi.Size() == sfi.Size() &&
		dfi.ModTime().Unix() == sfi.ModTime().Unix() {
		// Seems to not be modified.
		return nil
	}

	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, newDirPermissions); err != nil {
		return err
	}

	df, err := os.Create(dst)
	if err != nil {
		return err
	}
	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sf.Close()

	n, err := io.Copy(df, sf)
	if err == nil && n != sfi.Size() {
		err = fmt.Errorf("copied wrong size for %s -> %s: copied %d; want %d", src, dst, n, sfi.Size())
	}
	cerr := df.Close()
	if err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Chmod(dst, sfi.Mode())
	}
	if err == nil {
		err = os.Chtimes(dst, sfi.ModTime(), sfi.ModTime())
	}
	return err
}

func mirrorDir(src, dst string) error {
	err := filepath.Walk(src, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		suffix, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("failed to find Rel(%q, %q): %v", src, path, err)
		}
		return mirrorFile(path, filepath.Join(dst, suffix))
	})
	return err
}

func runProcess(command string, arg ...string) {
	cmd := exec.Command(command, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("Execute %v\n", cmd.Args)
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func compileGoPackage() {
	runProcess("go", "build", "-o", getGaugeExecutablePath(inProgress))
}

func getGaugeExecutablePath(file string) string {
	return filepath.Join(getBinDir(), getExecutableName(file))
}

func getExecutableName(file string) string {
	if getGOOS() == "windows" {
		return file + ".exe"
	}
	return file
}

// key will be the source file and value will be the target
func copyFiles(files map[string]string, installDir string) {
	for src, dst := range files {
		base := filepath.Base(src)
		installDst := filepath.Join(installDir, dst)
		log.Printf("Copying %s -> %s\n", src, installDst)
		stat, err := os.Stat(src)
		if err != nil {
			panic(err)
		}
		if stat.IsDir() {
			err = mirrorDir(src, installDst)
		} else {
			err = mirrorFile(src, filepath.Join(installDst, base))
		}
		if err != nil {
			panic(err)
		}
	}
}

func copyPluginFiles(destDir string) {
	files := make(map[string]string)
	if getGOOS() == "windows" {
		files[filepath.Join(getBinDir(), inProgress+".exe")] = bin
	} else {
		files[filepath.Join(getBinDir(), inProgress)] = bin
	}
	files[pluginJSONFile] = ""
	copyFiles(files, destDir)
}

func getPluginVersion() string {
	pluginProperties, err := getPluginProperties(pluginJSONFile)
	if err != nil {
		panic(fmt.Sprintf("Failed to get properties file. %s", err.Error()))
	}
	return pluginProperties["version"].(string)
}

func getBinDir() string {
	if *binDir == "" {
		return filepath.Join(bin, fmt.Sprintf("%s_%s", getGOOS(), getGOARCH()))
	}
	return filepath.Join(bin, *binDir)
}

func setEnv(envVariables map[string]string) {
	for k, v := range envVariables {
		os.Setenv(k, v)
	}
}

var install = flag.Bool("install", false, "Install to the specified prefix")
var pluginInstallPrefix = flag.String("plugin-prefix", "", "Specifies the prefix where the plugin will be installed")
var distro = flag.Bool("distro", false, "Creates distributables for the plugin")
var allPlatforms = flag.Bool("all-platforms", false, "Compiles or creates distributables for all platforms windows, linux, darwin both x86 and x86_64")
var binDir = flag.String("bin-dir", "", "Specifies OS_PLATFORM specific binaries to install when cross compiling")

var (
	platformEnvs = []map[string]string{
		{GOARCH: ARM64, GOOS: DARWIN, cgoEnabled: "0"},
		{GOARCH: x86_64, GOOS: DARWIN, cgoEnabled: "0"},
		{GOARCH: X86, GOOS: LINUX, cgoEnabled: "0"},
		{GOARCH: x86_64, GOOS: LINUX, cgoEnabled: "0"},
		{GOARCH: ARM64, GOOS: LINUX, cgoEnabled: "0"},
		{GOARCH: X86, GOOS: WINDOWS, cgoEnabled: "0"},
		{GOARCH: x86_64, GOOS: WINDOWS, cgoEnabled: "0"},
	}
)

func getPluginProperties(jsonPropertiesFile string) (map[string]interface{}, error) {
	pluginPropertiesJson, err := ioutil.ReadFile(jsonPropertiesFile)
	if err != nil {
		fmt.Printf("Could not read %s: %s\n", filepath.Base(jsonPropertiesFile), err)
		return nil, err
	}
	var pluginJson interface{}
	if err = json.Unmarshal(pluginPropertiesJson, &pluginJson); err != nil {
		fmt.Printf("Could not read %s: %s\n", filepath.Base(jsonPropertiesFile), err)
		return nil, err
	}
	return pluginJson.(map[string]interface{}), nil
}

func main() {
	flag.Parse()
	if *install {
		updatePluginInstallPrefix()
		installPlugin(*pluginInstallPrefix)
	} else if *distro {
		createPluginDistro(*allPlatforms)
	} else {
		compilePlugin()
	}
}

func compilePlugin() {
	if *allPlatforms {
		compileAcrossPlatforms()
	} else {
		compileGoPackage()
	}
}

func createPluginDistro(forAllPlatforms bool) {
	os.RemoveAll(deploy)
	if forAllPlatforms {
		for _, platformEnv := range platformEnvs {
			setEnv(platformEnv)
			fmt.Printf("Creating distro for platform => OS:%s ARCH:%s \n", platformEnv[GOOS], platformEnv[GOARCH])
			createDistro()
		}
	} else {
		createDistro()
	}
}

func createDistro() {
	packageName := fmt.Sprintf("%s-%s-%s.%s", inProgress, getPluginVersion(), getGOOS(), getArch())
	distroDir := filepath.Join(deploy, packageName)
	copyPluginFiles(distroDir)
	createZipFromUtil(deploy, packageName)
	os.RemoveAll(distroDir)
}

func runCommand(command string, arg ...string) {
	cmd := exec.Command(command, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Printf("Execute %v\n", cmd.Args)
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func createZipFromUtil(dir, name string) {
	wd, _ := os.Getwd()
	os.Chdir(filepath.Join(dir, name))
	runCommand("zip", "-r", filepath.Join("..", name+".zip"), ".")
	os.Chdir(wd)
}

func compileAcrossPlatforms() {
	for _, platformEnv := range platformEnvs {
		setEnv(platformEnv)
		fmt.Printf("Compiling for platform => OS:%s ARCH:%s \n", platformEnv[GOOS], platformEnv[GOARCH])
		compileGoPackage()
	}
}

func installPlugin(installPrefix string) {
	os.RemoveAll(deployDir)
	copyPluginFiles(deployDir)
	pluginInstallPath := filepath.Join(installPrefix, inProgress, getPluginVersion())
	err := mirrorDir(deployDir, pluginInstallPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to mirror directory  '%s' to '%s': %s", deployDir, pluginInstallPath, err.Error()))
	}
}

func updatePluginInstallPrefix() {
	if *pluginInstallPrefix == "" {
		if runtime.GOOS == "windows" {
			*pluginInstallPrefix = os.Getenv("APPDATA")
			if *pluginInstallPrefix == "" {
				panic(fmt.Errorf("failed to find AppData directory"))
			}
			*pluginInstallPrefix = filepath.Join(*pluginInstallPrefix, gauge, plugins)
		} else {
			userHome := getUserHome()
			if userHome == "" {
				panic(fmt.Errorf("failed to find User Home directory"))
			}
			*pluginInstallPrefix = filepath.Join(userHome, dotGauge, plugins)
		}
	}
}

func getUserHome() string {
	return os.Getenv("HOME")
}

func getArch() string {
	arch := getGOARCH()
	if arch == x86_64 {
		return "x86_64"
	}
	return arch
}

func getGOARCH() string {
	goArch := os.Getenv(GOARCH)
	if goArch == "" {
		return runtime.GOARCH

	}
	return goArch
}

func getGOOS() string {
	goOS := os.Getenv(GOOS)
	if goOS == "" {
		return runtime.GOOS
	}
	return goOS
}
