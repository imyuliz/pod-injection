# pod-injection

####  项目介绍

1. 如果发布的pod的Annotations中包含了 `yulibaozi/injected:true`和`yulibaozi/noinjected:true`中任意一个注解, 此Pod不会被注入sidecar
    `yulibaozi/injected:true`  : 表明这个Pod已经被注入过sidecar
    `yulibaozi/noinjected:true`:  表明这个pod不需要注入sidecar

2. 此项目在 kubernetes `v1.11.3` 测试通过, 运行正常


#### 项目结构
```
./
├── Dockerfile                     
├── README.md
├── .gitlab-ci.yml                          -- ci构建工具
├── deployment                              -- 部署应用的yaml配置
│   ├── 1-webhook-create-signed-cert.sh
│   ├── 2-mutatingwebhook-ca-bundle.yaml
│   ├── 2-mutatingwebhook.yaml
│   ├── 2-webhook-patch-ca-bundle.sh
│   ├── 3-rbac.yaml
│   ├── 4-crd-myconf.yaml
│   ├── 5-deployment.yaml
│   └── 6-test-pod.yaml
├── env.sh                                  -- 设置环境变量后,本地启动
└── src                                     -- 项目源码包
    └── injection
        ├── main.go                         -- 项目入口
        └── webhook          
            ├── admit_inject.go             -- 具体的sidecar注入逻辑
            ├── crd.go                      -- 具体从k8s中获取crd的配置信息逻辑
            ├── frame.go                    -- admission controller 大概框架
            └── types.go                    -- 常量和变量定义声明 

4 directories, 17 files
```


#### 部署步骤

##### 生成配置和准备阶段

1. 生成必要的证书到secret, 默认在`default` namesavce下
```
sh ./deployment/1-webhook-create-signed-cert.sh

# 查看是否创建成功
kubectl get secret yulibaozi-webhook-certs
```
2. 配置`MutatingWebhookConfiguration` 并设置 `caBundle` 并创建

```
cat ./deployment/2-mutatingwebhook.yaml | ./deployment/2-webhook-patch-ca-bundle.sh > ./deployment/2-mutatingwebhook-ca-bundle.yaml


# 创建 2-mutatingwebhook-ca-bundle.yaml 文件成功后,发布到kubernetes

kubectl apply -f ./deployment/2-mutatingwebhook-ca-bundle.yaml

# 查看是否创建成功

kubectl get MutatingWebhookConfiguration

```
3. 由于 webhook 需要访问k8s(原因是获取crd配置信息), 所以需要设置rbac

```
kubectl apply -f ./deployment/3-rbac.yaml
```

4. 由于此 webhook 需要去 crd 中获取sidecar 配置,所以需要配置crd (MyConf)

```
kubectl apply -f ./deployment/4-crd-myconf.yaml
```

##### webhook部署阶段

5. 部署webhook 应用 和暴露服务

```
kubectl create -f ./deployment/5-deployment.yaml
```

##### 测试`webhook`是否生效

6. 发布测试Pod, 查看此Pod是否会注入sidecar
```
kubectl create -f ./deployment/6-test-pod.yaml

# 查看是否注入pod
kubectl get pod  podtest -o yaml
```



