# go-sentry-kubernetes
> Sentry.io reporting client for kubernetes


The official [sentry-kubernetes](https://github.com/getsentry/sentry-kubernetes) is written in python and has a major memory leak. This reporter is written in golang and uses less than 10MB ram.


### Install
> Create a new project on Sentry and use your project DSN

Running from cli:
```
kubectl run go-sentry-kubernetes \
  --image stevelacy/go-sentry-kubernetes \
  --env="DSN=$YOUR_DSN"
```

Installing as a deployment:


Save as `deployment.yaml`

```yaml
# Deployment for go-sentry-kubernetes
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: go-sentry-kubernetes
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: go-sentry-kubernetes
    spec:
      containers:
      - name: go-sentry-kubernetes
        env:
        - name: DSN
          value: $SENTRY_DSN
        - name: ENV
          value: production
        image: stevelacy/go-sentry-kubernetes
        resources:
          limits:
            memory: "50M"
            cpu: "0.15"
          requests:
            cpu: "0.1"
            memory: "20M"
```

`$ kubectl apply -f ./deployment.yaml`


Set the `--debug` flag to enable debug logs:

```
      containers:
      - name: go-sentry-kubernetes
        args:
        - --debug
        command:
        - /app/main
        image: stevelacy/go-sentry-kubernetes

```

![screenshot](./screenshot.png)


MIT
