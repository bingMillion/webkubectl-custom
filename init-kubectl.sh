#!/bin/bash
set -e

if [ "${WELCOME_BANNER}" ]; then
    echo ${WELCOME_BANNER}
fi

arg1=$1
arg2=$2

# 新增第三个参数。 当kubeconfig时参数2为token。当kubeapi时参数3为token
arg3=$3

mkdir -p /nonexistent
mount -t tmpfs -o size=${SESSION_STORAGE_SIZE} tmpfs /nonexistent
cd /nonexistent
cp /root/.bashrc ./
cp /etc/vim/vimrc.local .vimrc
echo 'source /opt/kubectl-aliases/.kubectl_aliases' >> .bashrc
echo -e 'PS1="> "\nalias ll="ls -la"' >> .bashrc
# 禁用可以查看密钥的一些简单指令
echo 'alias vi="echo 'command forbiddon'"' >> .bashrc
echo 'alias vim="echo 'command forbiddon'"' >> .bashrc
echo 'alias cat="echo 'command forbiddon'"' >> .bashrc
echo 'alias less="echo 'command forbiddon'"' >> .bashrc
echo 'alias more="echo 'command forbiddon'"' >> .bashrc

mkdir -p .kube

export HOME=/nonexistent
if [ -z "${arg3}" ]; then    # 如果变量3为空，则arg的只有两个参数，1为kubeconfig，2为token
    echo $arg1| base64 -d > .kube/config
    # 新增写入历史记录
    touch /mnt/${arg2}
    chown -R nobody:nogroup /mnt/${arg2}
    echo "HISTFILE=/mnt/${arg2}" >> ~/.bashrc
    echo "HISTFILESIZE=5000" >> ~/.bashrc    # 历史记录文件最大行数为5000
    echo "HISTSIZE=5000" >> ~/.bashrc        # 单个会话中history最大记录数为5000
    echo "PROMPT_COMMAND='history -a'" >> ~/.bashrc   #PROMPT_COMMAND表示每次显示提示符前，执行下这个变量的value内容，这里是执行下历史记录追加命令。
else
    # 新增写入历史记录
    touch /mnt/${arg3}
    chown -R nobody:nogroup /mnt/${arg3}
    echo "HISTFILE=/mnt/${arg3}" >> ~/.bashrc
    echo "HISTFILESIZE=5000" >> ~/.bashrc
    echo "HISTSIZE=5000" >> ~/.bashrc
    echo "PROMPT_COMMAND='history -a'" >> ~/.bashrc

    echo `kubectl config set-credentials webkubectl-user --token=${arg2}` > /dev/null 2>&1
    echo `kubectl config set-cluster kubernetes --server=${arg1}` > /dev/null 2>&1
    echo `kubectl config set-context kubernetes --cluster=kubernetes --user=webkubectl-user` > /dev/null 2>&1
    echo `kubectl config use-context kubernetes` > /dev/null 2>&1
fi

if [ ${KUBECTL_INSECURE_SKIP_TLS_VERIFY} == "true" ];then
    {
        clusters=`kubectl config get-clusters | tail -n +2`
        for s in ${clusters[@]}; do
            {
                echo `kubectl config set-cluster ${s} --insecure-skip-tls-verify=true` > /dev/null 2>&1
                echo `kubectl config unset clusters.${s}.certificate-authority-data` > /dev/null 2>&1
            } || {
                echo err > /dev/null 2>&1
            }
        done
    } || {
        echo err > /dev/null 2>&1
    }
fi

chown -R nobody:nogroup .kube

export TMPDIR=/nonexistent

envs=`env`
for env in ${envs[@]}; do
    if [[ $env == GOTTY* ]];
    then
        unset ${env%%=*}
    fi
done

unset WELCOME_BANNER PPROF_ENABLED KUBECTL_INSECURE_SKIP_TLS_VERIFY SESSION_STORAGE_SIZE KUBECTL_VERSION

exec su -s /bin/bash nobody