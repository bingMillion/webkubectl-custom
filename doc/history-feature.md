# feature：自动记录每个用户的操作记录，按token名保存到mnt文件中

Q: 每个webshell的历史记录怎么区分？
A：
在start-session.sh中，进行启动一个全新的webshell。 在${all} 传入的参数中：
* 如果是kubeconfig的session，就是1个kubeconfig参数。
* 如果是kubetokenapi，则是2个参数分别为apiserver地址和token。
没有任何区分不同用户的信息传入，所以我们需要将${all}的参数，增加一个token参数。 

这样在start-session.sh之后，即下一步【init-kubectl.sh】时，
就可以根据新的token参数，创建以此命名的操作记录文件，将操作记录自动记录到文件中。

而在start-session.sh之前，就要将参数进行编写导入， 即：
当用户在前端页面选中某个session进行连接的时候，假设选中的是通过kubeconfig配置的session。 则发送请求到后端，后端"handleKubeConfigApi"方法接受请求，构造一个新的session webshell。
在Arg拼接kubeconfig，或者apisever+kubetoken时，将token参数也增加进去即可。
