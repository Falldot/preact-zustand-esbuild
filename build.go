package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/Falldot/pze/pkg/dev"
	"github.com/evanw/esbuild/pkg/api"
)

const PUBLIC_DIR = "public/"
const DEV_BUILD_DIR = "build/"
const RELEASE_BUILD_DIR = "dist/"
const INDEX_HTML_NAME = "index.html"
const PATH_TO_STYLES = PUBLIC_DIR + "styles"
const PATH_TO_INDEX_HTML = PUBLIC_DIR + INDEX_HTML_NAME

func main() {
	release := flag.Bool("release", false, "Release build")
	flag.Parse()

	if *release {
		SassRelease()
		JsRelease()
		HTML(RELEASE_BUILD_DIR)
	} else {
		go SassDev()
		JsDev()
		HTML(DEV_BUILD_DIR)

		<-make(chan bool)
	}

	fmt.Println("Build done.")
}

func HTML(to string) error {
	bytesRead, err := ioutil.ReadFile(PATH_TO_INDEX_HTML)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(to+"index.html", bytesRead, 0644)
	if err != nil {
		return err
	}
	return nil
}

func JsDev() {
	result := api.Build(api.BuildOptions{
		EntryPoints: []string{"src/app.tsx"},
		Outdir:      DEV_BUILD_DIR + "js",
		EntryNames:  "[dir]/[name]",
		Inject:      []string{"./node_modules/react/index.js"},
		Sourcemap:   api.SourceMapLinked,
		Bundle:      true,
		Write:       true,
		Metafile:    true,
		Incremental: true,
		Define: map[string]string{
			"process.env.NODE_ENV": "development",
		},
		Platform: api.PlatformBrowser,
		Target:   api.ES2021,
	})

	if len(result.Errors) > 0 {
		errorHandler(result.Errors, nil)
		os.Exit(1)
	}

	var server dev.DevServer
	server = dev.DevServer{
		Port:      "8080",
		WatchDir:  "src",
		Index:     "build/index.html",
		StaticDir: "build",
		OnReload: func() {
			result := result.Rebuild()
			if errorHandler(result.Errors, server.SendError) {
				server.SendReload()
			}
		},
	}
	go server.Start()
}

func JsRelease() {
	result := api.Build(api.BuildOptions{
		EntryPoints: []string{"src/app.tsx"},
		Outdir:      RELEASE_BUILD_DIR + "js",
		EntryNames:  "[dir]/[name]",
		Inject:      []string{"./node_modules/react/index.js"},
		Define: map[string]string{
			"process.env.NODE_ENV": "production",
		},
		Platform:          api.PlatformBrowser,
		Target:            api.ES2021,
		Bundle:            true,
		Write:             true,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Metafile:          true,
	})

	if len(result.Errors) > 0 {
		errorHandler(result.Errors, nil)
		os.Exit(1)
	}
}

func SassDev() error {
	out := exec.Command("sass", "--watch", PATH_TO_STYLES+":"+DEV_BUILD_DIR+"css")
	stderr, err := out.StderrPipe()
	if err != nil {
		return err
	}
	stdout, err := out.StdoutPipe()
	if err != nil {
		return err
	}
	go func() {
		merged := io.MultiReader(stderr, stdout)
		scanner := bufio.NewScanner(merged)
		for scanner.Scan() {
			msg := scanner.Text()
			fmt.Println(msg)
		}
	}()
	if err := out.Run(); err != nil {
		return err
	}
	return nil
}

func SassRelease() error {
	cmd := exec.Command("sass", PATH_TO_STYLES+":"+RELEASE_BUILD_DIR+"css")
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func errorHandler(errors []api.Message, callback func(string)) bool {
	if len(errors) > 0 {
		str := api.FormatMessages(errors, api.FormatMessagesOptions{
			Kind:  api.ErrorMessage,
			Color: true,
		})
		for _, err := range str {
			fmt.Println(err)
			if callback != nil {
				callback(err)
			}
		}
		return false
	}
	return true
}
