golang api框架
==============

示例
----

服务端:

```go
package main

import "github.com/solomoner/gozilla"

type HelloRequest struct {
	Name string
}

type HelloReply struct {
	Reply string
}

// 定义一个Service
type HelloService struct {
}

// 定义Hello方法，最终方法名字为/hello/Hello
func (s *HelloService) Hello(ctx *gozilla.Context, r *HelloRequest) (*HelloReply, error) {
	rep := &HelloReply{Reply: "hello " + r.Name}
	return rep, nil
}

func main() {
	gozilla.RegisterService(new(HelloService), "hello")
	gozilla.ListenAndServe(":8000")
}
```

客户端:

-	通过GET请求 `curl 127.0.0.1:8000/hello/Hello?name=icexin`
-	通过POST form请求 `curl 127.0.0.1:8000/hello/Hello -d name=icexin`
-	通过POST json请求 `curl 127.0.0.1:8000/hello/Hello -H "Content-Type: application/json" -d {"name":"icexin"}`

返回结果:

```json
{
  "code": 200,
  "msg": "",
  "data": {
    "Reply": "hello icexin"
  }
}
```

service规范
-----------

-	receiver必须是exported(第一个字母是大写)或者是调用`RegisterService`所在的package
-	方法名字必须是exported
-	方法有两个参数:`*Context`, `*args`
-	两个参数必须是指针
-	第二个参数必须是exported或是本地可见.
-	方法有两个返回值: reply和error

codec说明
---------

参数的获取以及返回值的格式化是通过codec定义的，codec的选取是通过HTTP头里面的`Content-Type`确定的。 默认集成了3个codec，其对应的Content-Type以及类型为:

| Content-Type                      | Codec     | 说明                               |
|-----------------------------------|-----------|------------------------------------|
| 空                                | FormCodec | 参数会从querystring里面获取        |
| application/x-www-form-urlencoded | FormCodec | 参数或从post form里面获取          |
| application/json                  | JSONCodec | 参数从body里面获取，参数为json格式 |

这三个codec的返回值统一为如下格式

```json
{
  "code": 200,
  "msg": "",
  "data": {}
}
```

其中data字段是根据每个方法json序列化后得到的。

参数验证
--------

gozilla使用`https://github.com/go-playground/validator`作为参数验证器，参数验证在方法的入参上进行。
可以通过设置`Options`里面的`EnableValidator`来开启或关闭参数验证，默认开启。
参数验证的具体语法参照`https://godoc.org/gopkg.in/go-playground/validator.v9`。

日志打印
-------

gozilla默认会把各个方法的access log打开，具体配置可以通过`LogOptions`来控制。
如果想了解日志组件的使用，可以参考`ListenAndServe`的实现。
