# PK

做了个 Android 自动化打包的命令，用 Golang 写。自增版本号，打包，压缩并上传到七牛，提交仓库，发送Email 所有的东西一气呵成

## 下载与配置

[Windows](https://github.com/ChenViVi/pk/releases/download/yeah/pk-win.exe) | [Mac](https://github.com/ChenViVi/pk/releases/download/yeah/pk-mac) | [Linux](https://github.com/ChenViVi/pk/releases/download/yeah/pk-linux)

首先得确定 [7z](https://www.7-zip.org/) 和 [qshell](https://github.com/qiniu/qshell) 已经配置好了

然后下载 [pk.zip](https://github.com/ChenViVi/pk/releases/download/yeah/pk.zip)，解压到项目的根目录下，请不要更改文件夹名称。可以把测试 key 换成你用来打包的 key，并在 `pk/base.gradle`中配置密码。如果你使用新打包方式，那么请在 `pk/channel.txt` 中定义好渠道名，这种打包方式只能用代码设置友盟渠道，像是这样`UMConfigure.init(this, "dfrr3e43ed", PackerNg.getChannel(this), UMConfigure.DEVICE_TYPE_PHONE, null);`。如果使用旧打包方式，请在 `pk/pk_old.gradle` 中定义渠道名

接着要确认的是 `app/build.gradle`中定义好了版本号和版本名，不要使用 ` rootProject.ext.versionCode`之类的，像是这样

```
apply plugin: 'com.android.application'

android {
    compileSdkVersion rootProject.ext.compileSdkVersion
    defaultConfig {
        applicationId "com.iwritebug.pk"
        minSdkVersion rootProject.ext.minSdkVersion
        targetSdkVersion rootProject.ext.targetSdkVersion
        versionCode 32
        versionName "1.17"
        flavorDimensions "pk"
    }
}

dependencies {
    implementation fileTree(dir: 'libs', include: ['*.jar'])
    implementation 'com.android.support:appcompat-v7:28.0.0'
}
```

然后在项目根目录下的 `build.gradle`的 `dependencies`配置中加入 `classpath 'com.mcxiaoke.packer-ng:plugin:2.0.1'`，像是这样

```
buildscript {
    
    repositories {
        google()
        jcenter()
    }
    dependencies {
        classpath 'com.android.tools.build:gradle:3.1.4'
        classpath 'com.mcxiaoke.packer-ng:plugin:2.0.1'
    }
}

allprojects {
    repositories {
        google()
        jcenter()
    }
}

task clean(type: Delete) {
    delete rootProject.buildDir
}
```

完成上面的步骤后，在项目根目录下打开终端并输入 `pk` 命令，执行完成之后应该能在项目的`build`目录下看到打包好的文件

## 参数解释

### `pk -o`

是否使用旧版本打包方式，默认不需要提供此参数，直接使用新打包方式。什么是新打包方式，参考 [这里](https://github.com/mcxiaoke/packer-ng-plugin)，目前发现此方式的唯一致命缺点就是无法根据不同渠道号使用不同资源文件，这种情况可能就需要使用旧版打包方式，也就是 ` gradle assembleRelease`

旧版本打包方式需要在 `pk/pk_old.gradle`中定义渠道名，新打包方式需要在 `pk/channel.txt` 中定义渠道名

### `pk -v`

是否自增版本号并提交到 Git 仓库，默认关闭此功能

### `pk -e`

打包完成后是否压缩上传并以邮件告知，需提供邮件地址

### eg

使用旧打包方式，自增版本号并提交到 Git 仓库，并把打包好的压缩包发送提醒到 `test@test.com`

```
pk -o -v -e="test@test.com"
```

## 交叉编译

```
### 编译Windows可执行程序
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o pk-win.exe pk.go

### 编译Linux可执行程序
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o pk-linux pk.go

### 编译Mac可执行程序
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o pk-mac pk.go
```