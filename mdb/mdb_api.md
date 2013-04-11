# 概述 #
## 目标用户 ##
MDB API是MDB平台向其它业务模块公开的一组网络对象模型的访问接口。它基于HTTP协议，并以REST风格提供被管理资源服务。相应HTTP方法的含义为：

- GET：获取一个资源的信息。
- POST：创建一个资源。
- PUT：更新一个资源的内容。
- DELETE：删除一个资源。

## 知识准备 ##
阅读者需要具备以下基础知识：
- 理解HTTP协议：请求和响应的基本过程和内容格式，基本的HTTP 方法。
- 了解HTTP各种返回码的含义
- REST：了解REST风格的WEB服务。
- JSON：了解JSON数据格式
## 阅读指南 ##

每个接口涉及以下几部分：

- 请求方式

	- HTTP方法：描述访问URL使用的HTTP方法。
	- URL：描述公开资源的URL。

- 请求参数：包含在URL Query String中的参数说明，如果某接口说明没有该项，说明该接口不支持任何参数（如果你提供了在接口定义之外的参数，系统将会忽略这些参数）。
- 请求体：包含在Body中的输入说明；如果某接口说明没有该项，说明该接口不支持任何body输入（额外输入的body将会被忽略）
- 正常结果

	HTTP Response Code

	HTTP Response Body：描述BODY的内容。
- 错误结果

	HTTP Response Code

	HTTP Response Body：描述错误原因。

## 使用说明 ##
　　客户使用任何支持GET、POST、PUT和DELTE方法的HTTP协议客户端，通过MDB API的URL提交请求，并接收响应。
### Method规范 ###
　　本接口文档部分接口要求的请求方法为HTTP GET方法，但又需要请求者提供HTTP BODY，对于这种情况，客户端需要进行HTTP方法转义，具体转义方法为：

- 用HTTP POST方式提交请求
- URL后缀中增加一个参数对: _method=GET

### URL规范 ###
	URL中的内容必须按照 RFC1738 规范经过编码
### 参数规范 ###
1. URL请求参数名称大小写敏感

	keyword写成Keyword也会导致该参数无效
2. 未定义的参数将会被忽略

	名称错误的参数将会被忽略

	这些参数错误的请求不会被识别为错误的请求，仅仅是忽略。

### 请求体规范 ###

如果提交需要提交 HTTP BODY内容时，内容必须是JSON数据且必须采用严格的JSON格式，参考RFC 4627

### 响应规范 ###

接收数据时，应该判断 HTTP 的状态码。
20x：表示正常响应，客户可从HTTP BODY中获得约定的数据，均为JSON数据格式。
4xx：由于客户端的输入错误请求导致的异常情况。
5xx：服务端执行时发生异常情况。


# 接口 #
## 元模型的定义 ##

元模型有一份完整的 xml schema 文档， 它对元模型的完整的描述， 请见 <<typeDefinitions.xsd>>, 这里就不说了。

## 创建 ##
### 描述 ###

用请求参数创建指定的模块。


### 请求方式 ###

POST http://mdb_server/mdb/<模型名>

<模型名> 为模型的名字，它定义在元模型中。

### 请求参数 ###

 请求体为模块的 json 字符表达。请求参数中所有的属性必须是元模型中定义的，并符合元模型中该属性的约束。需要注意的是：

- created_at、updated_at 和 _id 不能更改(updated_at 每次更新时会自动填充值，忽略用户的值)
 

### 正常结果 ###

  HTTP状态码 | Http Body 
 -----------|-----------
  201       | json 对象  

对象结构如下：

  字段       | 类型      | Http Body       
 -----------|-----------|-----------------
  created   | datetime  | 创建时的时间      
  value     | objectId  | 数据库中对象的id  


### 错误结果 ###

  HTTP状态码 | Http Body 
 -----------|-----------
  500       | 错误原因   

### 示例 ###
POST http://mdb_server/mdb/document 
{"author":"mfk","isbn":"sfasfd-sf-ssf","name":"aa","page_count":300,"publish_at":1,"type":"book"}

响应：
HTTP/1.1 201 Created

{"created_at":"2013-04-07T22:02:11.3595766+08:00","value":"51617c6362ddea1b500002eb"}

## 更新 ##
### 描述 ###

用请求参数更新指定的对象。

### 请求方式 ###

PUT http://mdb_server/mdb/<模型名>/<对象id>

<模型名> 为模型的名字，它定义在元模型中。

### 请求参数 ###

请求参数中所有的属性必须是元模型中定义的，并符合元模型中该属性的约束。但有两点需要注意：

- 只读属性、created_at、updated_at 和 _id 不能更改(updated_at 每次更新时会自动填充值，忽略用户的值)
- 必选属性在更新时不再是必要的了
 

### 正常结果 ###

  HTTP状态码 | Http Body 
 -----------+-----------
  200       |   <none>  


### 错误结果 ###

  HTTP状态码 | Http Body 
 -----------|-----------
  500       | 错误原因   

### 示例 ###
PUT http://mdb_server/mdb/device/51617a6062ddea1b500000cd

{"name":"dd21", "catalog":21, "services":221}

响应：
HTTP/1.1 200 OK

{"created_at":"2013-04-07T22:02:11.3595766+08:00","value":"51617c6362ddea1b500002eb"}

## 删除 ##
### 描述 ###

删除指定的对象及其子对象。子对象是指以本对象作为父亲的对象。

### 请求方式 ###

DELETE http://mdb_server/mdb/<模型名>/<对象id>

<模型名> 为模型的名字，它定义在元模型中。

### 请求参数 ###

无

### 正常结果 ###

  HTTP状态码 | Http Body 
 -----------|-----------
  200       |   <none>  


对象结构如下：

  字段       | 类型      | Http Body       
 -----------|-----------|-----------------
  created   | datetime  | 创建时的时间      
  effected  | integer   | 影响的行数       
  value     | objectId  | 数据库中对象的id  

### 错误结果 ###

  HTTP状态码 | Http Body 
 -----------|-----------
  500       | 错误原因   

### 示例 ###

DELETE http://mdb_server/mdb/device/51617a6062ddea1b500000cd

响应：

HTTP/1.1 200 OK

{"created_at":"2013-04-07T21:52:52.8656325+08:00","effected":1,"value":"51617a6062ddea1b500000cd"}



## 查询 ##
### 描述 ###

查询指定的对象

### 请求方式 ###

GET http://mdb_server/mdb/<模型名>/<对象id>

<模型名> 为模型的名字，它定义在元模型中。

### 请求参数 ###

无

### 正常结果 ###

  HTTP状态码 | Http Body 
 -----------|-----------
  200       |   <none>  

对象结构如下：

  字段       | 类型      | Http Body           
 -----------|-----------|---------------------
  created   | datetime  | 创建时的时间          
  value     | json      | 模块的 json 字符表达  

### 错误结果 ###

  HTTP状态码 | Http Body 
 -----------|-----------
  500       | 错误原因   

### 示例 ###

GET http://mdb_server/mdb/device/51617a6062ddea1b500000cd

响应：

HTTP/1.1 200 OK

{"created_at":"2013-04-07T21:52:53.7086808+08:00","value":{"_id":"51617a3562ddea1b50000001","address":"192.168.1.9","catalog":2,"created_at":"2013-04-07T21:52:53.013+08:00","description":"Hardware: Intel64 Family 6 Model 58 Stepping 9 AT/AT COMPATIBLE - Software:Windows Version 6.1 (Build 7601 Multiprocessor Free)","location":"","name":"meifakun-PC","oid":"1.3.6.1.4.1.311.1.1.3.1.1","services":76,"updated_at":"2013-04-07T21:52:53.013+08:00"}}


## 批量查询 ##
### 描述 ###

按条件查询指定的对象

### 请求方式 ###

GET http://mdb_server/mdb/<模型名>/query?<查询表达式>
(下个版本改为 http://mdb_server/mdb/<模型名>?<查询表达式> )

<模型名> 为模型的名字，它定义在元模型中。
 
<查询表达式> 请见下面的《查询条件式》节

### 请求参数 ###

无

### 正常结果 ###

  HTTP状态码 | Http Body 
 -----------|-----------
  200       |   <none>  

对象结构如下：

  字段       | 类型      | Http Body           
 -----------|-----------|---------------------
  created   | datetime  | 创建时的时间          
  value     | json      | 模块的 json 字符表达的数组  

### 错误结果 ###

  HTTP状态码 | Http Body 
 -----------|-----------
  500       | 错误原因   

### 示例 ###

GET http://mdb_server/mdb/device?@catalog=3

响应：

HTTP/1.1 200 OK


{"created_at":"2013-04-07T21:52:53.7086808+08:00","value":[{"_id":"51617a3562ddea1b50000001","address":"192.168.1.9","catalog":3,"created_at":"2013-04-07T21:52:53.013+08:00","description":"Hardware: Intel64 Family 6 Model 58 Stepping 9 AT/AT COMPATIBLE - Software:Windows Version 6.1 (Build 7601 Multiprocessor Free)","location":"","name":"meifakun-PC","oid":"1.3.6.1.4.1.311.1.1.3.1.1","services":76,"updated_at":"2013-04-07T21:52:53.013+08:00"}]}

## 统计总数 ##
### 描述 ###

按条件查询指定的对象

### 请求方式 ###

GET http://mdb_server/mdb/<模型名>/query?<查询表达式>
(下个版本改为 http://mdb_server/mdb/<模型名>?<查询表达式> )

<模型名> 为模型的名字，它定义在元模型中。
 
<查询表达式> 请见下面的《查询条件式》节

### 请求参数 ###

无

### 正常结果 ###

  HTTP状态码 | Http Body 
 -----------|-----------
  200       |   <none>  

对象结构如下：

  字段       | 类型      | Http Body           
 -----------|-----------|---------------------
  created   | datetime  | 创建时的时间          
  value     | integer   | 模块的 json 字符表达的数组  

### 错误结果 ###

  HTTP状态码 | Http Body 
 -----------|-----------
  500       | 错误原因   

### 示例 ###

GET http://mdb_server/mdb/device?@catalog=3

响应：

HTTP/1.1 200 OK

{"created_at":"2013-04-07T21:52:53.748683+08:00","value":12}

# 特殊字段 #

数据库中有三个字段名作为保留字给系统使用，用户不需要定义。

- _id

	对象的唯一标识符，类型为一个特殊的字符串

- created_at

	创建对象的时间戳，为一个datetime，它是只读的（仅在创建时填值）

- updated_at

	创建对象的时间戳，为一个datetime，它是只读的（在创建或更新时会忽略用户的值）

- type

	用户保存对象的类型，为一个字符串，它的值为对象的名称，它是只读的

# 继承 #

当一个模型继承与另一个模型时在对它的增删改查有什么影响呢？比如下面有三个对象：

		  Document
            /   \
           /     \
         Book    Magazine

	Book和Magazine都是继承于Document

当然下面分别说明对各种操作的影响.

## 创建 ##

当创建 Document 时没有什么特殊的说法，它之前的完全一样。而当创建 Book 时有两种方式

- 标准做法, 和上面一样 url 为 http://mdb_server/mdb/book
- 非标准做法,  url 为 http://mdb_server/mdb/document, 但在对象的属性中必须含有 "type=book" 

## 更新 ##

当更新 Document 时没有什么特殊的说法，它之前的完全一样。而当更新 Book 时有两种方式

- 标准做法, 和上面一样 url 为 http://mdb_server/mdb/book/xxxx_id
- 非标准做法,  url 为 http://mdb_server/mdb/document/xxxx_id, (注意与创建不一样，对象的属性中必须不能含有 "type=book"）

## 删除 ##

当删除 Document 时没有什么特殊的说法，它之前的完全一样。而当删除 Book 时有两种方式

- 标准做法, 和上面一样 url 为 http://mdb_server/mdb/book/xxxx_id
- 非标准做法,  url 为 http://mdb_server/mdb/document/xxxx_id

## 查询 ##

当查询时就能体现继承对建模带来的好处了，

	http://mdb_server/mdb/book/ 仅能查询类型为 book 的对象
	http://mdb_server/mdb/magazine/ 仅能查询类型为 magazine 的对象
	http://mdb_server/mdb/document/ 仅能查询类型为 document、book 和 magazine 的对象


# 命名 #



# 数据格式 #
## 查询表达式 ##
查询表达式位于url中queryString部分，因此必须符合 queryString 格式。 queryString中可能还会放置其它信息，因必须能与其它信息区分开，因此我设计如下：

- 查询表达式由多个过滤表达式组成，多个过滤表达式之间为"and"关系
- queryString中有 “@” 开头的项为过滤表达式
- 过滤表达式分为三个部分， 两个操作数和一个操作符，格式为 "操作数1=[操作符]操作数2", 当操作符为“=”时可以省略为 "操作数1=操作数2"

操作符暂时有:

  操作符  |  含义
  -------|----------
  eq     |  等于
  gt     |  大于
  gte    |  大于等于
  lt     |  小于
  lte    |  小于等于
  

