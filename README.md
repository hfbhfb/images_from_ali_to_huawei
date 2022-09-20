

## 原理(在官方镜像工具套了一层): 
- 列举所有阿里的镜像,输出config.json文件
- 需要预先手动通过华为云管理后台[web]创建同名组织(在阿里云是命名空间)
- 使用 "阿里云官方" 镜像工具(https://github.com/AliyunContainerService/image-syncer) 进行镜像同步

## 编译和运行
``` bash
export AK="xxxx" #阿里云ak
export SK="xxxx" #阿里云sk
export RegionAli="cn-hangzhou" # 阿里云区域 https://help.aliyun.com/document_detail/198107.html
export UserAli="xxx" #阿里云镜像用户
export PasswdAli="xxx" #阿里云镜像密码
export RegionHw="cn-south-1" # 华为云区域  https://developer.huaweicloud.com/endpoint?SWR
export UserHw="xxx" #华为云镜像用户
export PasswdHw="xxx" #华为云镜像密码
export RunFlag="1"  # 开关,当此值为1时才真正的执行镜像同步
export OnlyRun="false" #开关,如果打开则跳过config.json的生成过程
make
```


