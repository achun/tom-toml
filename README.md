tom-toml
========

[TOML](https://github.com/mojombo/toml) format parser for Golang

This library supports TOML version
[v0.2.0](https://github.com/mojombo/toml/blob/master/versions/toml-v0.2.0.md)

[![Build Status](https://drone.io/github.com/achun/tom-toml/status.png)](https://drone.io/github.com/achun/tom-toml/latest)

## Import

    import "github.com/achun/tom-toml"

## Usage

Say you have a TOML file that looks like this:

```toml
[servers.alpha]
ip = "10.0.0.1" # IP
dc = "eqdc10"
```

Read the ip and dc like this:

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

outputs:

```
10.0.0.1
eqdc10
```

## Documentation

The documentation is available at
[gowalker.org](http://gowalker.org/github.com/achun/tom-toml).

## Contribute

Feel free to report bugs and patches using GitHub's pull requests system on
[achun/tom-toml](https://github.com/achun/tom-toml). Any feedback would be
much appreciated!


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
