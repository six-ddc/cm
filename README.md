# cm

一个简单的配置管理工具，基于redis存储，然后通过读取redis设定的key来替换配置文件


### 安装

1. 有go语言环境

```
    go get github.com/six-ddc/cm
```

2. 没有go语言环境，可以下载[release包](https://github.com/six-ddc/cm/releases)


### 使用


// 查看帮助
	
	cm --help

// 增加username和passwd配置

	cm --redis "127.0.0.1:6379" --redis-prefix "test" set db.username=root db.passwd=123456

// 替换前 test/example.ini

	[db]
	username=${db.username}
	passwd=${db.passwd}

// 替换配置文件中的对应配置

	cm --redis "127.0.0.1:6379" --redis-prefix "test" get --output test/example.ini.out test/example.ini

// 替换后 test/example.ini.out

	[db]
	username=root
	passwd=123456
