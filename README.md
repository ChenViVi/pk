# PK

准备做个 Android 自动化打包的命令，用 Golang 写。自增版本号，打包，压缩并上传到七牛，提交仓库，发送Email 所有的东西一气呵成

## 食用方法

### 直接执行编译好的可执行文件

Windows
```
./pk-win
```

Linux

```
chmod +x pk-linux
./pk-linux
```

Mac

```
chmod +x pk-mac
./pk-mac
```


### 直接执行 go 文件，需要带上 app build.gradle 的绝对路径

```
go run pk.go --path C:\Users\vivi\Desktop\pk\app\build.gradle
```

### 交叉编译

```
### 编译Windows可执行程序
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o pk-win pk.go

### 编译Linux可执行程序
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o pk-linux pk.go

### 编译Mac可执行程序
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o pk-mac pk.go
```