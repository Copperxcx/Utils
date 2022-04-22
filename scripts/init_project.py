import mysqlx
import sys

def exit():
    print("\n-------------------项目数据库初始化脚本-End-------------------\n\n")
    sys.exit()


'''
初始化一个项目，创建项目数据库
设项目名称为test，则：
1. 创建数据库 db_project_test
2. 数据库中创建日志表 tbl_log
3. 数据库中创建配置文件表 tbl_config
'''

if __name__ == '__main__':
        

    print("""\n-------------------项目数据库初始化脚本-Start-------------------\n""")

    # 输入参数，至少需要2个，依次为：脚本名称（始终存在）、项目名称、用户名、密码、
    args = sys.argv

    name = ""
    dbname = ""
    ip = "localhost"
    port = 33060
    user = "root"
    pwd = "root"

    if len(args) < 2: 
        name = input("请输入项目名称: ")
    else:
        name = args[1]
    dbname = "db_project_" + name

    print("项目名称: ", name)
    # 第3,4个参数为用户名和密码
    if len(args) >= 4:
        user = args[2]
        pwd = args[3]


    # 第5,6个参数为mysql服务器IP地址和端口
    if len(args) >= 6:
        ip = args[4]
        port = int(args[5])



    # Connect to server on localhost
    my_session = mysqlx.get_session({
        'host': ip, 'port': port,
        'user': user, 'password': pwd
    })

    # 数据库存在判断
    s = my_session.get_schema(dbname)
    if s.exists_in_database():
        str = "项目 " + name + " 已存在，是否覆盖? [y/n] :"
        x = input(str)
        print()
        if x.lower() != "y":
            print("保留现有项目数据，脚本结束！")
            exit()
        else:
            print("删除现有项目数据...")
            my_session.sql("DROP DATABASE IF EXISTS " + dbname + ";").execute()


    print("创建新数据库，项目名称：", name)
    # 创建数据库
    my_session.sql("CREATE DATABASE "+dbname+";").execute()
    my_session.sql("USE " + dbname + ";").execute()


    print("创建配置文件表：tbl_config")
    # 创建日志表 tbl_log 和配置文件表 tbl_config
    my_session.sql("""CREATE TABLE tbl_config( 
        name VARCHAR(20) KEY NOT NULL, 
        value VARCHAR(20) NOT NULL, 
        memo VARCHAR(100) NOT NULL DEFAULT \"\");""").execute()


    print("创建日志表：tbl_log")
    my_session.sql("""CREATE TABLE tbl_log( 
        id INT KEY NOT NULL AUTO_INCREMENT, 
        type VARCHAR(10) NOT NULL, 
        detail VARCHAR(100) NOT NULL,
        ts TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3));""").execute()

    # 添加日志：项目创建
    # tbl = my_session.get_schema(dbname).get_table("tbl_log")
    
    # tbl.insert("['type', 'detail', 'occur_time']").values("INFO", "项目数据库创建", "2022-04-20 19:00:00").execute()

    print("添加日志：创建数据库...")
    my_session.sql("INSERT INTO tbl_log (type, detail) VALUES( 'INFO', 'database created' );").execute()


    print("\n项目", name, "数据库初始化成功！")

    exit()