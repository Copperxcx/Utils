## 依赖
1. `MySQL 8.0.27`，通过 `mysql --version` 查看；
2. `Go 1.17.6`，通过 `go version` 查看；

## 功能
1. 监听`localhost:9000`，提供`http`服务以对项目日志表、配置表存储/查看；
2. 仅支持`POST`服务，提交`JSON`格式数据；
3. 指定项目名称如`test`后，将对数据库`db_project_test`下的`tbl_config`和`tbl_log`进行操作；

### 数据库操作模块

通过监听某本地端口，仅提供`HTTP POST`服务，实现对现有数据库表格记录的**增、查**。增加对应日志，查询对应配置文件。

#### 设置数据库参数

可以设置全部，或是只设置数据库名称，只设置名称时提供`op、dbname`即可，`op`为`config_database`：

```json
{
    "op": "config",
    "ip": "127.0.0.1",
    "port": "3306",
    "username": "root",
    "pwd": "root",
    "dbname": "db_proj_test"
}
```

#### 配置文件增删改查

1. 在当前数据库中创建`config`表格，字段有：`id name value memo`

   ```json
   {
       "op": "create_config_table"
   }
   ```

2. 添加配置条目，提供`name value`字段，可选`memo`字段：

   ```json
   {
       "op": "add_config",
       "records": [
       	["key1", "value1", "memo1"],
   		["key2", "value2"]
       ]
   }
   ```

3. 删除配置条目，提供`name`字段，不提供时不删除并报错：

   ```json
   {
   	"op": "del_config",
       "records": ["key1", "key2"]
   }
   ```

4. 修改配置条目，提供`old_name new_value new_memo new_name`，后`2`个可以不提供：

   ```json
   {
       "op": "update_config",
       "records": [
           ["key1", "new_value1", "new_memo1", "new_key1"],
           ["key2", "new_value2", "new_memo1"],
           ["key3", "new_value3"]
       ]
   }
   ```

5. 查询配置条目，提供`name`字段时，返回对应配置条目，否则返回所有配置：

   ```json
   {
       "op": "get_config",
       "records": ["key1", "key2", "key3"]
   }
   ```

6. 删除配置表格：

   ```json
   {
       "op": "del_config_table"
   }
   ```

7. 清空配置条目：

   ```json
   {
   	"op": "clear_config"
   }
   ```

   

#### 日志表格添加

1. 创建日志表格，字段包含：`id type detail occur_time`

   ```json
   {
       "op": "create_log_table",
   }
   ```

2. 删除日志表格：

   ```json
   {
       "op": "del_log_table",
   }
   ```

3. 添加日志条目

   ```json
   {
       "op": "add_log",
       "records": [
           ["INFO", "This is an INFO.", "2022-04-01 19:00:00"],
           ["ERROR", "This is an ERROR.", "2022-04-01 19:00:00"],
       ]
   }
   ```

   