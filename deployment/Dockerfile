from golang:1.11.12 as builder

copy cmd /go/src/admissionWebhook/cmd

env GOPROXY https://goproxy.io
env GO111MODULE on

workdir /go/src/admissionWebhook
run go mod init admissionWebhook

workdir /go/src/admissionWebhook/cmd
run go build -o admissionwebhook .



from golang:1.11.12 as runner

copy --from=builder /go/src/admissionWebhook/cmd/admissionwebhook /go/bin/admissionwebhook

#copy cmd/webhook /go/bin/webhook
copy pki /data/pki

entrypoint ["/go/bin/admissionwebhook"]
cmd ["--tls-cert-file", "/data/pki/amourlinux-admission.pem", "--tls-private-key-file", "/data/pki/amourlinux-admission-key.pem"]
