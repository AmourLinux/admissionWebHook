from golang:1.11.12 as builder

copy cmd /go/src/admissionWebhook/cmd
copy pki /go/src/admissionWebhook/pki

env GOPROXY https://goproxy.io
env GO111MODULE on

workdir /go/src/admissionWebhook/cmd

entrypoint ["go build ."]



from golang:1.11.12 as runner

copy --from=builder /go/src/admissionWebhook/cmd/main /go/bin/webhook

#copy cmd/webhook /go/bin/webhook
copy pki /data/pki

entrypoint ["/go/bin/webhook"]

cmd ["--tls-cert-file", "/data/pki/amourlinux-admission.pem", "--tls-private-key-file", "/data/pki/amourlinux-admission-key.pem"]
