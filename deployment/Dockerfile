from golang:1.11.5

copy webhook /go/bin/webhook
copy go-pki /data/go-pki

entrypoint ["/go/bin/webhook"]

cmd ["--tls-cert-file", "/data/go-pki/amourlinux-admission.pem", "--tls-private-key-file", "/data/go-pki/amourlinux-admission-key.pem"]
