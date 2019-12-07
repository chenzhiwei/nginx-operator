# Integrate with OLM

## Generate CSV

https://github.com/operator-framework/operator-sdk/blob/master/doc/user/olm-catalog/generating-a-csv.md

```
operator-sdk olm-catalog gen-csv --csv-version 0.0.1 --update-crds
```

### Update the generated CSV required fields:

deploy/olm-catalog/nginx-operator/0.0.1/nginx-operator.v0.0.1.clusterserviceversion.yaml

* keywords
* maintainers
* provider

```
spec:
  keywords:
  - nginx
  maintainers:
  - email: zhiweik@gmail.com
    name: zhiwei
  provider:
    name: zhiwei
```

### Update the generated CSV owned CRD fields:

deploy/olm-catalog/nginx-operator/0.0.1/nginx-operator.v0.0.1.clusterserviceversion.yaml

* description
* displayName

```
spec:
  customresourcedefinitions:
    owned:
    - kind: Nginx
      name: nginxes.app.siji
      version: v1alpha1
      description: The Nginx CRD
      displayName: Nginx
```

### Update package name to nginx-app

deploy/olm-catalog/nginx-operator/nginx-operator.package.yaml

## Push to quay.io


### Get quay.io token

```
pip3 install operator-courier
git clone https://github.com/operator-framework/operator-courier.git
./operator-courier/scripts/get-quay-token
# {"token": "basic abcdefghijkl=="}

export QUAY_TOKEN="basic abcdefghijkl=="
```

### Lint

```
operator-courier verify --ui_validate_io deploy/olm-catalog/nginx-operator
```

### Push

```
export OPERATOR_DIR=deploy/olm-catalog/nginx-operator
export QUAY_NAMESPACE=siji
export PACKAGE_NAME=nginx-app
export PACKAGE_VERSION=0.0.1

operator-courier push "$OPERATOR_DIR" "$QUAY_NAMESPACE" "$PACKAGE_NAME" "$PACKAGE_VERSION" "$QUAY_TOKEN"
```

### Test with OCP

```
kubectl apply -f deploy/olm-test/operator-source.yaml

kubectl create namespace siji-nginx

kubectl apply -f deploy/olm-test/nginx-operator-group.yaml

kubectl apply -f deploy/olm-test/nginx-subscription.yaml
```


## In the end

Finally we will get following:

* quay.io/siji/nginx:fake The app image
* quay.io/siji/nginx-operator The oprator image
* quay.io/siji/nginx-app The OLM application
