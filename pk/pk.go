package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const gradleFilePath = "app/build.gradle"
const versionCodeMark = "versionCode"
const versionNameMark = "versionName"
const oldPKGradleConfig = "\napply from: \"../pk/pk_old.gradle\""
const newPKGradleConfig = "\napply from: \"../pk/pk.gradle\""
const downloadURL = "http://adesk.valorachen.top"
const siteURL = "http://files.valorachen.top/index.php?bucket=adesk&name=%E6%97%A5%E5%B8%B8%E6%9B%B4%E6%96%B0%E5%8C%85"

func main() {
	var absolutePath string
	var oldModeEnabled bool
	var versionUpdateEnabled bool
	var email string
	flag.StringVar(&absolutePath, "p", "/", "app module 下的 build.gradle路径，仅用于调试，默认情况下无需添加")
	flag.BoolVar(&oldModeEnabled, "o", false, "是否使用旧版打包方式，默认为使用新打包方式")
	flag.BoolVar(&versionUpdateEnabled, "v", false, "是否自增版本号并提交到 Git，默认关闭")
	flag.StringVar(&email, "e", "", "上传七牛并发邮件时需要的邮件地址，默认关闭此功能")
	flag.Parse()

	if strings.Compare(absolutePath, "/") == 0 {
		absolutePath = getCurrentDirectory() + "../app/build.gradle"
	}
	fmt.Printf("provided path was %s\n", absolutePath)
	exists := Exists(absolutePath)
	if exists {
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
		gradleConfigBytes, _ := ioutil.ReadFile(absolutePath)
		gradleConfig := string(gradleConfigBytes)

		//versionCode:13
		versionCodeRe, _ := regexp.Compile(versionCodeMark + ".?([1-9]\\d*)")
		currentVersionCodeConfig := versionCodeRe.FindString(gradleConfig)
		//13
		codeRe, _ := regexp.Compile("[1-9]\\d*")
		currentVersionCode := codeRe.FindString(currentVersionCodeConfig)
		newVersionCode, _ := strconv.Atoi(currentVersionCode)
		//13+1

		if versionUpdateEnabled {
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
		}

		//versionName: "35.1.2"
		versionNameRe, _ := regexp.Compile(versionNameMark + ".\"?(.*?)\"")
		currentVersionNameConfig := versionNameRe.FindString(gradleConfig)
		//35.1.2
		nameRe, _ := regexp.Compile("\\d+(\\.\\d+)*")
		currentVersionName := nameRe.FindString(currentVersionNameConfig)
		newVersionName := currentVersionName
		if versionUpdateEnabled {
			//2
			lastNameRe, _ := regexp.Compile("\\d*")
			lastNameArray := lastNameRe.FindAllString(currentVersionName, -1)
			lastName := lastNameArray[len(lastNameArray)-1]
			//2+1
			newLastName, _ := strconv.Atoi(lastName)
			newLastName++
			//35.1.3
			newVersionNamePart2 := strconv.Itoa(newLastName)
			newVersionNamePart1 := currentVersionName[0 : len(currentVersionName)-len(lastName)]
			newVersionName = newVersionNamePart1 + newVersionNamePart2
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
		}

		gradleConfig = strings.Replace(gradleConfig, oldPKGradleConfig, "", -1)
		gradleConfig = strings.Replace(gradleConfig, newPKGradleConfig, "", -1)
		if oldModeEnabled {
			gradleConfig = gradleConfig + oldPKGradleConfig
		} else {
			gradleConfig = gradleConfig + newPKGradleConfig
		}

		//func WriteFile(filename string, data []byte, perm os.FileMode) error
		//WriteFile writes data to a file named by filename. If the file does not exist,
		//WriteFile creates it with permissions perm; otherwise WriteFile truncates it before writing.
		//https://golang.org/pkg/io/ioutil/
		//所以我还是不懂第三个参数具体干嘛的
		ioutil.WriteFile(absolutePath, []byte(gradleConfig), 0644)
		fmt.Println("version code:" + strconv.Itoa(newVersionCode))
		fmt.Println("version name:" + newVersionName)
		if oldModeEnabled {
			ExeCommand("gradle", "assembleRelease")
		} else {
			ExeCommand("gradle", "apkRelease")
		}
		if versionUpdateEnabled {
			ExeCommand("git", "add", gradleFilePath)
			ExeCommand("git", "commit", "-m", "来自自动打包程序，已自动更新到版本v"+newVersionName)
			ExeCommand("git", "tag", newVersionName)
			ExeCommand("git", "push")
		}
		if email != "" {
			applicationIDRe, _ := regexp.Compile("applicationId" + ".\"(.*)\"")
			applicationID := applicationIDRe.FindString(gradleConfig)
			pkNameRe, _ := regexp.Compile("\"(.*)\"")
			pkName := pkNameRe.FindString(applicationID)
			pkName = strings.Replace(pkName, "\"", "", -1)
			outputDirName := pkName + "_code" + strconv.Itoa(newVersionCode) + "_name" + newVersionName
			ExeCommand("7z", "a", "build/"+outputDirName+".7z", "build/"+outputDirName+"/*.apk")
			ExeCommand("qshell", "rput", "adesk", outputDirName+".7z", "build/"+outputDirName+".7z")
			fmt.Println("visit link: " + siteURL)
			body := "<html><body>" +
				"<h3>下载地址：" + downloadURL + "/" + outputDirName + ".7z" + "</h3>" +
				"<h3>更多内容请查看：" + siteURL + "</h3>" +
				"</body></html>"
			err := SendMail("acodeplayer@163.com", "playhard7", "smtp.163.com:25", email, pkName+"更新包_v"+newVersionName, body, "html")
			if err != nil {
				fmt.Println("发送邮件失败!")
				fmt.Println(err)
			} else {
				fmt.Println("发送邮件成功!")
			}
		}
	} else {
		fmt.Println("gradle配置文件不存在")
	}
}

func ExeCommand(commandName string, arg ...string) bool {
	//函数返回一个*Cmd，用于使用给出的参数执行name指定的程序
	cmd := exec.Command(commandName, arg...)

	//显示运行的命令
	fmt.Println(cmd.Args)
	//StdoutPipe方法返回一个在命令Start后与命令标准输出关联的管道。Wait方法获知命令结束后会关闭这个管道，一般不需要显式的关闭该管道。
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		fmt.Println(err)
		return false
	}

	cmd.Start()
	//创建一个流来读取管道内内容，这里逻辑是通过一行一行的读取的
	reader := bufio.NewReader(stdout)

	//实时循环读取输出流中的一行内容
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		fmt.Println(line)
	}

	//阻塞直到该命令执行完成，该命令必须是被Start方法开始执行的
	cmd.Wait()
	return true
}

func SendMail(user, password, host, to, subject, body, mailtype string) error {
	hp := strings.Split(host, ":")
	auth := smtp.PlainAuth("", user, password, hp[0])
	var contentType string
	if mailtype == "html" {
		contentType = "Content-Type: text/" + mailtype + "; charset=UTF-8"
	} else {
		contentType = "Content-Type: text/plain" + "; charset=UTF-8"
	}

	msg := []byte("To: " + to + "\r\nFrom: " + user + "<" + user + ">\r\nSubject: " + subject + "\r\n" + contentType + "\r\n\r\n" + body)
	sendTo := strings.Split(to, ";")
	err := smtp.SendMail(host, auth, user, sendTo, msg)
	return err
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
