package main

import (
	"saas_autopack/util"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var globalJsLines []string

const GLOBALJSPATH = "./src/global.js"

const CONFIGPATH = "./config-overrides.js"

func main() {
	//读取gloabl.js
	globalJsByte, err := ioutil.ReadFile(GLOBALJSPATH)
	globalJsLines = strings.Split(string(globalJsByte), "\n")
	checkErr(err)

	files := make(map[string]bool)
	folders := make(map[string]bool)
	//读取文件夹

	fileAndFolders, err := ioutil.ReadDir(".")
	checkErr(err)

	for _, v := range fileAndFolders {
		if v.IsDir() {
			folders[v.Name()] = true
		} else {
			files[v.Name()] = true
		}
	}
	//判断运行根目录是否正确
	if _, ok := folders["src"]; !ok {
		fmt.Println("文件目录错误")
		os.Exit(0)
	}
	if _, ok := files["package.json"]; !ok {
		fmt.Println("文件目录错误")
		os.Exit(0)
	}
	//读取src/template
	fmt.Println("正在读取src/template")
	templates, err := ioutil.ReadDir("./src/template")
	checkErr(err)

	templateNames := make(map[string]string)
	for _, v := range templates {
		if v.IsDir() {
			templateNames[v.Name()] = "./src/template/" + v.Name()
		}
	}
	//build
	for k, v := range templateNames {
		fmt.Println("building", k)
		err = buildTemplate(k, v)
		checkErr(err)
		_ = os.RemoveAll("./build")
	}
}

func checkErr(err error) {
	if err != nil {
		fmt.Println("运行错误", err)
	}
}

func buildTemplate(name, path string) error {
	//修改global.js
	content := fmt.Sprintf("import Config from './template/%s';", name)
	err := editGlobal(content)
	if err != nil {
		return err
	}
	//修改config-overrides
	appNo, err := editConfig(path)
	if err != nil {
		return err
	}
	//yarn build
	yarn := exec.Command("yarn", "build")
	err = yarn.Run()
	if err != nil {
		return err
	}

	//复制index.html
	buildedIndexPath := "./build/index.html"
	buildedAppPath := "./build/" + appNo + "/index.html"
	err = util.CopyFile(buildedIndexPath, buildedAppPath)
	if err != nil {
		return err
	}

	//复制文件夹到根目录autopack
	if _, err := os.Stat("./autopack"); os.IsNotExist(err) {
		// path/to/whatever exists
		err = os.Mkdir("./autopack", 0755)
		if err != nil {
			return err
		}
	}

	err = util.CopyFolder("./build/"+appNo, "./autopack/"+appNo)
	if err != nil {
		return err
	}
	return nil
}

//修改global.js
func editGlobal(content string) error {
	globalJsLines[0] = content
	output := strings.Join(globalJsLines, "\n")
	err := ioutil.WriteFile(GLOBALJSPATH, []byte(output), 0755)
	return err
}

//修改config-overrides.js,返回appNo
func editConfig(path string) (string, error) {
	indexJS, err := ioutil.ReadFile(path + "/index.js")
	checkErr(err)

	reg, err := regexp.Compile(`formal = {[^s]+sysFlag: '(\d+)',`)

	checkErr(err)
	result := reg.FindAllSubmatch(indexJS, -1)
	if len(result) != 0 && len(result[0]) == 2 {
		appNo := string(result[0][1])
		appNoString := fmt.Sprintf("const dir = '%s/';", appNo)
		//修改config-overrides.js
		config, err := ioutil.ReadFile(CONFIGPATH)
		if err != nil {
			return "", err
		}
		//替换appNo
		configStr := string(config)
		configAppNoReg := regexp.MustCompile(`const dir = '\d+/';`)
		replaced := configAppNoReg.FindString(configStr)
		configStr = strings.Replace(configStr, replaced, appNoString, -1)
		err = ioutil.WriteFile(CONFIGPATH, []byte(configStr), 0755)
		if err != nil {
			return "", err
		}
		return appNo, nil
	}
	err = errors.New("读取AppNO失败")
	return "", err
}
