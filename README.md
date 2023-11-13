# 项目说明
* 新增自动记录历史命令功能，实现了按token导出，方便审计。
* 新增禁用一些简单指令功能，防止用户查看token信息等。
* 新增kubectl扩展指令-argorollout

# 构建与部署
1、在github页actions输入版本号进行自动构建,例如v202311131501100zz
2、本地部署运行
```shell
docker rm -f webkubectl && \
docker run --name="webkubectl" -p 8080:8080 -d --privileged zbxx/webkubectl:v202311131501100zz && \
docker exec -it webkubectl /bin/bash
```
3、访问localhost:8080即可
