# Nginx Operator

```
operator-sdk new nginx-operator --repo github.com/chenzhiwei/nginx-operator
operator-sdk add api --api-version=app.siji/v1alpha1 --kind=Nginx
operator-sdk generate k8s
operator-sdk generate openapi
operator-sdk add controller --api-version=app.siji/v1alpha1 --kind=Nginx
operator-sdk build quay.io/siji/nginx-operator:v0.0.1
```
