#!/bin/bash
set -e

all=$*

if [ -z "${all}" ]; then
    echo No Args provided
    echo Terminal will exit.
    sleep 1
    exit 1
fi

if [[ $all == ERROR:* ]]; then
    echo ${all}
    sleep 1
    exit 1
fi

# unshare 用于在新的命名空间中运行程序
# fork 创建一个新的进程；
# pid：用新的 PID 命名空间运行程序，这意味着这个新的进程将有它自己的 PID，与父进程分离。
# mount-proc： 挂载一个新的 /proc 文件系统到新的 PID 命名空间
# mount： 新的进程将拥有自己的挂载点命名空间
# /opt/webkubectl/init-kubectl.sh： 这是要在新的命名空间中运行的脚本文件。
# ${all}： 这是一个变量，表示要传递给 init-kubectl.sh 脚本的所有参数。   all的实际值，如果1个参数，就是kubeconfig，如果2个参数就是apiserver和token。 这是哪个接口调用过来的？？？
unshare --fork --pid --mount-proc --mount /opt/webkubectl/init-kubectl.sh ${all}