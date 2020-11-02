# enforce-ingress-class

自动强制为 Ingress 指定一个 IngressClass

## 使用方式

* 初始化 `admission-bootstrapper` 
  参照此文档 https://github.com/k8s-autoops/admission-bootstrapper ，完成 `admission-bootstrapper` 的初始化步骤
* 部署以下 YAML

```yaml
# create serviceaccount
apiVersion: v1
kind: ServiceAccount
metadata:
  name: enforce-ingress-class
  namespace: autoops
---
# create clusterrole
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: enforce-ingress-class
rules:
  - apiGroups: [""]
    resources: ["namespaces"]
    verbs: ["get"]
---
# create clusterrolebinding
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: enforce-ingress-class
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: enforce-ingress-class
subjects:
  - kind: ServiceAccount
    name: enforce-ingress-class
    namespace: autoops
---
# create job
apiVersion: batch/v1
kind: Job
metadata:
  name: install-enforce-ingress-class
  namespace: autoops
spec:
  template:
    spec:
      serviceAccount: admission-bootstrapper
      containers:
        - name: admission-bootstrapper
          image: autoops/admission-bootstrapper
          env:
            - name: ADMISSION_NAME
              value: enforce-ingress-class
            - name: ADMISSION_IMAGE
              value: autoops/enforce-ingress-class
            - name: ADMISSION_ENVS
              value: ""
            - name: ADMISSION_SERVICE_ACCOUNT
              value: "enforce-ingress-class"
            - name: ADMISSION_MUTATING
              value: "true"
            - name: ADMISSION_IGNORE_FAILURE
              value: "false"
            - name: ADMISSION_SIDE_EFFECT
              value: "None"
            - name: ADMISSION_RULES
              value: '[{"operations":["CREATE"],"apiGroups":["extensions", "networking.k8s.io"], "apiVersions":["*"], "resources":["ingresses"]}]'
      restartPolicy: OnFailure
```

* 为需要启用的命名空间，添加注解，指明要使用的内网

  `autoops.enforce-ingress-class=nginx`
  
  **可以配合 `enforce-ns-annotations` 自动为新命名空间启用此注解**

## Credits

Guo Y.K., MIT License
