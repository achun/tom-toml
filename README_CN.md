# tom-toml

[TOML](https://github.com/mojombo/toml) 格式 Go 语言支持包.

本包支持 TOML 版本
[v0.2.0](https://github.com/mojombo/toml/blob/master/versions/toml-v0.2.0.md)

[![Build Status](https://api.travis-ci.org/achun/tom-toml.png?branch=master)](https://travis-ci.org/achun/tom-toml)

这里有一篇关于 [TOML, 新的简洁配置语言](http://hit9.org/post/toml.html) 中文简介.

## Import

    import "github.com/achun/tom-toml"


## 使用

假设你有一个TOML文件 `example.toml` 看起来像这样:

```toml
# 注释以"#"开头, 这是多行注释, 可以分多行
# tom-toml 把这两行注释绑定到紧随其后的 key, 也就是 title

title = "TOML Example" # 这是行尾注释, tom-toml 把此注释绑定给 title

# 虽然只有一行, 这也属于多行注释, tom-toml 把此注释绑定给 owner
[owner] # 这是行尾注释, tom-toml 把这一行注释绑定到 owner

name = "om Preston-Werner" # 这是行尾注释, tom-toml 把这一行注释绑定到 owner.name

# 下面列举 TOML 所支持的类型与格式要求
organization = "GitHub" # 字符串
bio = "GitHub Cofounder & CEO\nLikes tater tots and beer." # 字符串可以包含转义字符
dob = 1979-05-27T07:32:00Z # 日期, 必须使用 RFC3339 格式. 对 Go 来说这很简单.

[database]
server = "192.168.1.1"
ports = [ 8001, 8001, 8002 ] # 数组, 其元素类型也必须是TOML所支持的. Go 语言下类型是 slice
connection_max = 5000 # 整型, tom-toml 使用 int64 类型
enabled = true # 布尔型

[servers]

  # 可以使用缩进, tabs 或者 spaces 都可以, 毫无问题.
  [servers.alpha]
  ip = "10.0.0.1" # IP 格式只能用字符串了
  dc = "eqdc10"

  [servers.beta]
  ip = "10.0.0.2"
  dc = "eqdc10"

[clients]
data = [ ["gamma", "delta"], [1, 2] ] # 又一个数组
donate = 49.90 # 浮点, tom-toml 使用 float64 类型
```

读取 `servers.alpha` 部分好像这样:

```go
import (
    "fmt"
    "github.com/achun/tom-toml"
)
func main() {
    conf, err := toml.LoadFile("good.toml")
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println(conf["servers.alpha.ip"].String())
    fmt.Println(conf["servers.alpha.dc"].String())
}
```

输出是这样的:

```
10.0.0.1
eqdc10
```

您应该注意到了注释的表现形式, tom-toml 提供了注释支持.

## 文档

Go DOC 文档请访问
[gowalker.org](http://gowalker.org/github.com/achun/tom-toml).


## 贡献

请使用 GitHub 系统提出 issues 或者 pull 补丁到
[achun/tom-toml](https://github.com/achun/tom-toml). 欢迎任何反馈！


## License
Copyright (c) 2014, achun
All rights reserved.

Redistribution and use in source and binary forms, with or without modification,
are permitted provided that the following conditions are met:

* Redistributions of source code must retain the above copyright notice, this
  list of conditions and the following disclaimer.

* Redistributions in binary form must reproduce the above copyright notice, this
  list of conditions and the following disclaimer in the documentation and/or
  other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
