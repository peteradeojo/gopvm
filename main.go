package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/peteradeojo/gopvm/types"
	"github.com/peteradeojo/gopvm/util"
)

// Flags
var (
	versionFlag bool
	installFlag string
	useFlag     string
	extFlag     string
)

var AppConfig *types.Config

func main() {
	flag.BoolVar(&versionFlag, "versions", false, "Get PHP Versions")
	flag.StringVar(&useFlag, "use", "", "Switch PHP Versions")
	flag.StringVar(&installFlag, "install", "", "Install PHP version")
	flag.StringVar(&extFlag, "ext", "", "Install PHP extensions")

	flag.Parse()

	runtime.GOMAXPROCS(8)

	// Setup config file
	AppConfig = &types.Config{}

	configFilename := ".pvm/config"
	configFile := loadConfigFile(&configFilename)

	AppConfig.SetFile(configFile)
	defer configFile.Close()

	AppConfig.ReadConfig(&configFilename)

	AppConfig.MakeDirs()

	defer AppConfig.Save()

	if versionFlag {
		fetchVersions()
		return
	}

	if useFlag != "" {
		v := prepareVersion(useFlag)
		useVersion(v)

		installPear(v)
		return
	}

	// Install Flow
	if installFlag != "" {
		v := prepareVersion(installFlag)

		dist, release, err := fetchReleaseDist(v)
		if err != nil {
			panic(err)
		}

		if dist == "" {
			panic("Unable to retrieve distribution")
		}

		loc, err := extractDist(release)
		if err != nil {
			panic(err)
		}

		_, err = configureDist(loc)
		if err != nil {
			log.Println("Unable to configure PHP installation. Please check the guide at https://github.com/peteradeojo/gopvm")
			panic(err)
		}

		err = linkVersion(loc)
		if err != nil {
			panic(err)
		}

		AppConfig.InstalledVersions = append(AppConfig.InstalledVersions, v)
		return
	}

	// extensions flag
	if extFlag != "" {
		FindPackage(extFlag)
	}
}

func BootstrapAppDir(dir string) {
	fs.ReadDir(os.DirFS(dir), ".pvm")
}

func loadConfigFile(fileName *string) *os.File {
	if fileName == nil {
		defaultConfig := "./.pvm/config.json"
		fileName = &defaultConfig

		if exists, _ := util.CheckDirExists("./.pvm"); !exists {
			os.Mkdir(".pvm", 0777)
		}
	}

	file, err := os.OpenFile(*fileName, os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		log.Fatalln(err)
	}

	return file
}

func loadCachedVersions() (types.ReleaseData, error) {
	var releaseData types.ReleaseData

	data, err := os.ReadFile(AppConfig.CacheDir + "versions.json")
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &releaseData)
	if err != nil {
		return nil, err
	}

	return releaseData, nil
}

func fetchVersionsFromRemote() (types.ReleaseData, error) {
	var releaseData types.ReleaseData
	response, err := http.Get("http://php.net/releases/index.php?json")

	if err != nil {
		return nil, err
	}

	r, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(r, &releaseData)

	// Cache versions
	file, err := os.OpenFile(AppConfig.CacheDir+"versions.json", os.O_CREATE|os.O_WRONLY, 0766)
	if err != nil {
		return releaseData, err
	} else {
		file.Write(r)
		file.Close()
	}

	return releaseData, nil
}

func fetchVersions() {
	releaseData, err := loadCachedVersions()
	if err != nil || releaseData == nil {
		log.Println(err)
		releaseData, err = fetchVersionsFromRemote()
	}

	if err != nil {
		log.Println(err)
	}

	if releaseData != nil {
		displayVersions(releaseData)
	}
}

func displayVersions(release types.ReleaseData) {
	for _, r := range release {
		fmt.Printf("Version: %s - Date: %s\n", r.Version, r.Date)
		if len(r.SupportedVersions) > 0 {
			fmt.Println("Supported Versions:")
			for _, s := range r.SupportedVersions {
				fmt.Println(s)
			}
		}
		fmt.Println()
	}
}

func prepareVersion(version string) string {
	if len(version) == 3 {
		return version + ".0"
	}

	return version
}

// Idempotent?
func resolveVersionToRelease(version string) string {
	if len(version) == 3 {
		version += ".0"
	}

	return fmt.Sprintf("php-%v.tar.gz", version)
}

func fetchReleaseDist(version string) (string, string, error) {
	// release := resolveVersionToRelease(version)
	dist, err := fetchDistFromCache(version)

	if dist == "" || err != nil {
		return getDistFromSource(version)
	}

	return dist, version, err
}

func resolveReleaseToCacheDest(version string) string {
	cache := fmt.Sprintf("%v%v", AppConfig.CacheDir, resolveVersionToRelease(version))
	return cache
}

func resolveReleaseToInstallDir(release string) string {
	return fmt.Sprintf("%v%v", AppConfig.InstallDir, "php-"+release)
}

func fetchDistFromCache(release string) (string, error) {
	cache := resolveReleaseToCacheDest(release)

	if e, err := util.CheckDirExists(cache); e {
		if err != nil {
			return "", err
		}

		if e {
			return cache, nil
		}
	}

	return "", nil
}

func getDistFromSource(version string) (string, string, error) {
	fmt.Printf("Fetching version from remote: %s\n", version)

	release := resolveVersionToRelease(version)
	url := fmt.Sprintf("https://www.php.net/distributions/%s", release)

	dest := resolveReleaseToCacheDest(version)
	fmt.Printf("Fetching version to destination: %s\n", dest)
	os.Remove(dest)

	cmd := prepareCommand("wget", url, "-O", dest)
	err := cmd.Run()

	if err != nil {
		log.Fatalf(err.Error())
	}

	return dest, version, err
}

func extractDist(version string) (string, error) {
	i := resolveReleaseToInstallDir(version)

	if e, _ := util.CheckDirExists(i); e {
		return resolveReleaseToInstallDir(version), nil
	}

	dist := resolveReleaseToCacheDest(version)
	cmd := prepareCommand("tar", "-xvf", dist, "--cd", AppConfig.InstallDir)
	err := cmd.Run()

	return resolveReleaseToInstallDir(version), err
}

func prepareCommand(cmd string, args ...string) *exec.Cmd {
	command := exec.Command(cmd, args...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command
}

func configureDist(location string) (string, error) {
	os.Chdir(location)

	if c, _ := util.CheckFileExists("./Makefile"); !c { // Makefile doesn't exist, ./configure hasn't been run
		iconvDir := os.Getenv("ICONV_DIR")

		// Setup install args to install basic useful extensions
		var installArgs []string = []string{"--enable-calendar", "--enable-bcmath", "--enable-exif", "--enable-ftp", "--enable-mbstring", "--enable-soap", "--enable-sockets", "--enable-shmop", "--enable-sysvsem", "--enable-sysvshm", "--enable-pcntl", "--enable-phar", "--enable-opcache", "--with-curl", "--with-openssl", "--with-zlib", "--with-mysqli", "--with-pdo-mysql", "--with-pdo-sqlite", "--with-zip", "--with-jpeg", "--with-freetype", "--with-gettext"}

		if iconvDir != "" {
			installArgs = append(installArgs, fmt.Sprintf("--with-iconv=%v", iconvDir))
		}

		fmt.Printf("%+v\n", installArgs)

		cmd := prepareCommand("./configure", installArgs...)
		err := cmd.Run()

		if err != nil {
			return "", err
		}
	}

	if c, _ := util.CheckFileExists("./sapi/cli/php"); !c { // php binary doesn't exist, make hasn' been run
		procs := runtime.NumCPU()

		fmt.Printf("Running make: make -j%d\n", procs)

		cmd := prepareCommand("make", fmt.Sprintf("-j%d", procs))
		err := cmd.Run()

		return location, err
	}

	return location, nil
}

func useVersion(version string) {
	dest := resolveReleaseToInstallDir(version)
	if c, _ := util.CheckDirExists(dest); !c {
		panic("Version distribution not found. Run pvm -install %version%")
	}

	err := linkVersion(dest)
	if err != nil {
		panic(err)
	}

	AppConfig.CurrentActiveVersion = version

	fmt.Printf("PHP current version has been set to %s\n", version)
}

func linkVersion(location string) error {
	cmd := prepareCommand("ln", "-s", "-F", fmt.Sprintf("%+s/sapi/cli/php", location), "/usr/local/bin/php")
	err := cmd.Run()

	if err != nil {
		return err
	}

	// cmd = prepareCommand("ln", "-s", "-F", fmt.Sprintf("%+s/php.ini-development", location), "/usr/local/lib/php.ini")

	// err = cmd.Run()

	cmd = prepareCommand("ln", "-s", "-F", fmt.Sprintf("%+s/php.ini-development", location), "/usr/local/lib/php.ini")
	err = cmd.Run()
	return err
}

func installPear(version string) error {
	location := resolveReleaseToInstallDir(version)

	fmt.Println(os.Getwd())

	prepareCommand("rm", "/usr/local/bin/pear").Run()
	prepareCommand("rm", "/usr/local/etc/pear.conf").Run()
	prepareCommand("rm", "-rf", "/usr/local/lib/php/pear").Run()

	cmd := prepareCommand(
		"php", fmt.Sprintf("%s/pear/install-pear-nozlib.phar", location),
		"-d", "include_path=/tmp/pear/temp",
		"-d", fmt.Sprintf("php_dir=$(pwd)/install/php-%+s/lib/php", version),
		"-d", fmt.Sprintf("doc_dir=$(pwd)/install/php-%+s/lib/php/doc", version),
		"-d", fmt.Sprintf("bin_dir=$(pwd)/install/php-%+s/bin", version),
		"-d", fmt.Sprintf("data_dir=$(pwd)/install/php-%+s/lib/php/data", version),
		"-d", fmt.Sprintf("cfg_dir=$(pwd)/install/php-%+s/lib/php/cfg", version),
		"-d", fmt.Sprintf("www_dir=$(pwd)/install/php-%+s/lib/php/htdocs", version),
		"-d", fmt.Sprintf("test_dir=$(pwd)/install/php-%+s/lib/php/test", version),
		"-d", "temp_dir=/tmp/pear/temp",
	)
	err := cmd.Run()

	if err != nil {
		fmt.Println(err)
	}

	// pharLocation := fmt.Sprintf("%s/pear/go-pear.phar", location)

	// cmd := prepareCommand("wget", "http://pear.php.net/go-pear.phar", "-O", pharLocation)
	// err := cmd.Run()
	return err
}
