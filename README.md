# kiwiproxy
动态代理，类似花生壳功能，实现内网应用发布到外网。配合Nginx，实现多域名转发到内网

# 使用方法

1.server 端，编译部署在外网服务器
<code>修改配置文件 server/config.json</code>
 <pre>
{
	"mainConnServer" : ":8888", ＃ 主线程监听地址及端口
	"proxyConnServer" : ":7777", # http服务器监听地址及端口
	"transConnServer" : ":9999" # 数据传输线程监听地址及端口
}
 </pre>
<code>./server </code>

2.client 端,编译部署在内网
<code>修改 client/config.json</code>
 <pre>
{
    "username":"alex", # 用户名，暂时没用
    "password":"xxxxx", # 密码，暂时没用
	"mainConnServer" : "121.40.64.190:8888", # 服务器主线程地址及端口
	"transConnServer" : "121.40.64.190:9999", # 服务器数据传输线程地址及端口
    "domains":[
        {
            "remote":"www.yourdomain.com",      # 要映射的域名或者IP地址。如果是域名，要先把域名DNS指向服务器
            "local":"127.0.0.1:81"         # 在客户端跑的http服务端的监听地址及端口。比如你在客户端跑了一个apache, 监听端口为81
        }
    ]
}

 </pre>
<code>./client </code>

这时候在浏览器输入 www.yourdomain.com:7777 应该就能访问到你本地的http服务了

3. 其它部署方式

	要使用转发功能的同学，肯定是资源上有所限制的。我也是因为要做本地调试，又苦于本地路由做不了端口映射，花生壳对linux客户端支持不太好等问题，才想到自己写一个转发代理。不必担心，结合nginx强大的功能，一切都不是问题。

	3.1 我有域名，我要转发80端口，但是服务器上有其它服务在使用80端口。
		解决方案：使用nginx，做一个路径转发。如 http://mydomain.com 转发到 http://127.0.0.1:7777

	3.2 我没有域名，我服务器有一个IP，我要转发80端口，但是服务器上有其它服务在使用80端口。
		解决方案：使用nginx，做一个路径转发。如 http://myip/wechatdebug 开头的都转发到 http://127.0.0.1:7777
