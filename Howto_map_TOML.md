# 问题

这里讨论基于用 map 实现 TOML 的问题.

当然, 我们把嵌套的 key 用 "." 连接. 这样的 map 访问数据非常直接方便. 你可以用

```go
tm["k.e.y"]  // 一下就访问到最终的目标
// 而不用像这样
tm.Get("k").Get("e").Get("y")
```

TOML v0.2.0 的定义中, Table/ArrayOfTables 是可以深层嵌套的.

参见 https://github.com/mojombo/toml/pull/153

用 map 完全实现 TOML, 最简单的方法是:

```go
map[string]interface{}
```

但是这种万能接口, 使用者还需要 type assertion, 并不方便, 所以用 struct 是需要的.

# 分析

先看看 TOML 都定义了什么?

 - Comment       注释. 前置(多行)注释和行尾注释还有 TOML 文本最后的注释
 - Value         String, Integer, Float, Boolean, Datetime 以及数组形式
 - Key           给 Value 命名, 以便访问
 - Table         一组 Key = Value 的集合
 - TableName     给 Table 命名
 - ArrayOfTables 正如其名, 笔者认为, 事实上这是 TOML 的嵌套语法
 - TablesName    给 ArrayOfTables 命名, 它表示了一组 Table 的名字

TOML 文本最后的注释比较特别, 只有一个. 可以特殊处理, 比如用 key="", 下文用 TomlComments 表示.

显而易见的规则:

    非被嵌套的 Value 是可以直接 key 访问.

    TableName 也是可以直接 "k.e.y" 访问. TableName 成了 Value 的一种类型.

    tm["k.e.y"] 里面的 key 其实是个访问路径, 由 "TableName[.Key]" 组成. 这是个完整的全路径.

    Table 是可以嵌套 Table 的, 同样用 "k.e.y" 访问.

    Table, ArrayOfTables 是可以互相嵌套的.

    在 map 中无法直接用 tm["k.e.y"] 访问到 ArrayOfTables 中的元素, 因为 key 中没法让数组下标生效, 如果加入下标,维护将会很麻烦.
    正如笔者前面说的, 如果把 ArraOfTables 当作 `Array Of TOML` 这就很容易理解了. 
    还不如直接命名为 TOMLArray 来的简单明了.

    注释. ArrayOfTables 中的每一个下标 [[TablesName]] 也允许有注释.

    格式化输出 TOML文本要求数据必须能被有序访问.

对 Toml 定义的影响:

```go
type Toml map[string]Item
```

Item 有可能是

 - Value
 - TableName
 - ArrayOfTables 经过前述分析, 就是嵌套的 TOML.
 - 内部实现需要的 "." 开头的数据

Value 的 Kind 包括

    String
    Integer
    Float
    Boolean
    Datetime
    StringArray
    IntegerArray
    FloatArray
    BooleanArray
    DatetimeArray
    Array         元素是 xxxxArray 类型, TOML 规范没有明确是否可以嵌套 Array.

TableName 和 ArrayOfTables 是独立的, 就是他们自己.

实现的时候, 当然所有的数据都用 interface{} 保存在 Value 结构中. 只不过在接口上 Value和Item有所区别. 实际上

    Value 的接口囊括了 TableName 的支持.
    ArrayOfTables 只有在 Item 接口中才能访问到.


# 结果

## 定义

```go
type Value struct{}  // 省略细节, 包含 TableName 类型

type Toml map[string]Item

func (t Toml) Fetch(prefix string) Toml // 返回访问路径以 "prefix." 开头的子集

type Tables []Toml  // 官方没有给出具体名字, 只有造一个

type Item struct{
    *Value
}

func (i Item) Tables() Tables // 如果 kind 是ArrayOfTables 的话

```
Item 导出 Value 可以方便一些操作, 目前 Item 只是多支持了 ArrayOfTables.
尽管有些不同, Kind 的命名尽量采用 TOML 定义的字面值.

## 访问

前面分析过"k.e.y"是个完全路径, 总是写完全路径有时候不是很方便. `Fetch(prefix)` 方法返回的 Toml 子集省略了 "prefix." 部分, 方便 map 访问. 总结如下

 - 数据不在 ArrayOfTables 中, 完全路径直接访问: tm["k.e.y"]
 - 数据在 ArrayOfTables 中, 先 tm["k.e.y"].Table(idx) 得到其中一个 Table, 这就是一个原 Toml 的子集, 然后仍然是完全路径访问数据.
 - 访问 `Fetch(prefix)` 返回的 Toml, key 减省了 `prefix.`.
 - 子集对象的修改会影响原 Toml, 但是增加和删除不会影响原 Toml.

## 限制

 - Toml 对象中 key 以"."开头的元素内部保留. 修改这些元素会产生不可预计的结果.
 - 禁止循环嵌套. 以"."开头的元素用来完成防止循环嵌套.
 - 有序嵌套, 先生成的对象, 不能作为后生成对象的嵌套部分.

限制的原因和内部实现有关, 不细述.

## ArrayOfTables
官方申明下面 Table 的文档是合法的.

```toml
# [x] you
# [x.y] don't
# [x.y.z] need these
[x.y.z.w] # for this to work
```

官方申明下面 ArrayOfTables的文档是非法的.

```toml
# INVALID TOML DOC
[[fruit]]
  name = "apple"

  [[fruit.variety]]
    name = "red delicious"

  # This table conflicts with the previous table
  [fruit.variety]
    name = "granny smith"
```

官方文档未明确下面的文档是否合法.

没有声明 `[foo]` 或者 `[[foo]]`, 直接

```toml
[[foo.bar]]
```

这应该是非法的, 因为如果补全这种写法的话, 可能是

```toml
[foo]
[[foo.bar]]
```

也有可能是

```toml
[[foo]]
[[foo.bar]]
```

会产生歧义.
