package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	termbox "github.com/nsf/termbox-go"
)

const gradleFilePath = "app/build.gradle"
const versionCodeMark = "versionCode"
const versionNameMark = "versionName"

func init() {
	if err := termbox.Init(); err != nil {
		panic(err)
	}
	termbox.SetCursor(0, 0)
	termbox.HideCursor()
}

func main() {
	var absolutePath string
	flag.StringVar(&absolutePath, "path", "/", "Provide project path as an absolute path")
	flag.Parse()

	if strings.Compare(absolutePath, "/") == 0 {
		absolutePath = getCurrentDirectory() + "/" + gradleFilePath
	}
	fmt.Printf("provided path was %s\n", absolutePath)
	exists := Exists(absolutePath)
	if exists {
		updateVersion(absolutePath)
	} else {
		fmt.Println("gradle配置文件不存在")
	}
}

func updateVersion(path string) {
	// ext{
	// 	supportVer = '28.0.0-alpha3'
	// 	minSdkVersion = 16
	// 	targetSdkVersion = 24
	// 	compileSdkVersion = 28
	// 	versionConfigs = [
	// 			versionCode: 13,
	// 			versionName: "35.1.2"
	// 	]
	// }
	gradleConfigBytes, _ := ioutil.ReadFile(path)
	gradleConfig := string(gradleConfigBytes)

	//versionCode:13
	versionCodeRe, _ := regexp.Compile(versionCodeMark + ".?([1-9]\\d*)")
	currentVersionCodeConfig := versionCodeRe.FindString(gradleConfig)
	//13
	codeRe, _ := regexp.Compile("[1-9]\\d*")
	currentVersionCode := codeRe.FindString(currentVersionCodeConfig)
	//13+1
	newVersionCode, _ := strconv.Atoi(currentVersionCode)
	newVersionCode++
	//versionCode: 14
	newVersionCodeConfig := codeRe.ReplaceAllString(currentVersionCodeConfig, strconv.Itoa(newVersionCode))
	// ext{
	// 	supportVer = '28.0.0-alpha3'
	// 	minSdkVersion = 16
	// 	targetSdkVersion = 24
	// 	compileSdkVersion = 28
	// 	versionConfigs = [
	// 			versionCode: 14,
	// 			versionName: "35.1.2"
	// 	]
	// }
	gradleConfig = versionCodeRe.ReplaceAllString(gradleConfig, newVersionCodeConfig)
	//versionName: "35.1.2"
	versionNameRe, _ := regexp.Compile(versionNameMark + ".\"?(.*?)\"")
	currentVersionNameConfig := versionNameRe.FindString(gradleConfig)
	//35.1.2
	nameRe, _ := regexp.Compile("\\d+(\\.\\d+)*")
	currentVersionName := nameRe.FindString(currentVersionNameConfig)
	//2
	lastNameRe, _ := regexp.Compile("\\d+")
	lastNameArray := lastNameRe.FindAllString(currentVersionName, -1)
	lastName := lastNameArray[len(lastNameArray)-1]
	//2+1
	newLastName, _ := strconv.Atoi(lastName)
	newLastName++
	//35.1.3
	newVersionNamePart2 := strconv.Itoa(newLastName)
	newVersionNamePart1 := currentVersionName[0 : len(currentVersionName)-len(newVersionNamePart2)]
	newVersionName := newVersionNamePart1 + newVersionNamePart2
	//versionName: "35.1.3"
	newVersionNameConfig := nameRe.ReplaceAllString(currentVersionNameConfig, newVersionName)
	// ext{
	// 	supportVer = '28.0.0-alpha3'
	// 	minSdkVersion = 16
	// 	targetSdkVersion = 24
	// 	compileSdkVersion = 28

	// 	versionConfigs = [
	// 			versionCode: 14,
	// 			versionName: "35.1.3"
	// 	]
	// }
	gradleConfig = versionNameRe.ReplaceAllString(gradleConfig, newVersionNameConfig)
	//func WriteFile(filename string, data []byte, perm os.FileMode) error
	//WriteFile writes data to a file named by filename. If the file does not exist,
	//WriteFile creates it with permissions perm; otherwise WriteFile truncates it before writing.
	//https://golang.org/pkg/io/ioutil/
	//所以我还是不懂第三个参数具体干嘛的
	ioutil.WriteFile(path, []byte(gradleConfig), 0644)
	fmt.Println("new version code:" + strconv.Itoa(newVersionCode))
	fmt.Println("new version name:" + newVersionName)
	fmt.Println("update version file sccuess")
	executeAndPrint("gradle", "apkRelease")
	applicationIDRe, _ := regexp.Compile("applicationId" + ".\"(.*)\"")
	applicationID := applicationIDRe.FindString(gradleConfig)
	pkNameRe, _ := regexp.Compile("\"(.*)\"")
	pkName := pkNameRe.FindString(applicationID)
	pkName = strings.Replace(pkName, "\"", "", -1)
	outputDirName := pkName + "_code" + strconv.Itoa(newVersionCode) + "_name" + newVersionName
	executeAndPrint("7z", "a", "pk/"+outputDirName+".7z", "pk/"+outputDirName+"/*.apk")
	executeAndPrint("qshell", "rput", "adesk", outputDirName+".7z", "pk/"+outputDirName+".7z")
	executeAndPrint("git", "add", gradleFilePath)
	executeAndPrint("git", "commit", "-m", "来自自动打包程序，已自动更新到版本v"+newVersionName)
	executeAndPrint("git", "tag", newVersionName)
	executeAndPrint("git", "push")
	fmt.Println("upload success!")
	fmt.Println("qshell rput " + outputDirName + ".7z" + " " + "pk/" + outputDirName + ".7z")
	fmt.Println("visit link: http://files.valorachen.top/index.php?bucket=adesk&name=%E6%97%A5%E5%B8%B8%E6%9B%B4%E6%96%B0%E5%8C%85")
}

func executeAndPrint(name string, arg ...string) {
	// 执行系统命令
	// 第一个参数是命令名称
	// 后面参数可以有多个，命令参数
	cmd := exec.Command(name, arg...)
	// 获取输出对象，可以从该对象中读取输出结果
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	// 保证关闭输出流
	defer stdout.Close()
	// 运行命令
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	// 读取输出结果
	opBytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(opBytes))
}

func pause() {
	fmt.Println("请按任意键继续...")
Loop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			break Loop
		}
	}
}

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

// 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// 判断所给路径是否为文件
func IsFile(path string) bool {
	return !IsDir(path)
}
