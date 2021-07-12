FROM golang:1.16
WORKDIR /go/src/github.com/kubernetes
RUN git clone https://github.com/kubernetes/ingress-nginx
WORKDIR /go/src/github.com/kubernetes/ingress-nginx/cmd/waitshutdown
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o new-wait-shurdown .

FROM busybox:latest  
WORKDIR /
COPY --from=0 /go/src/github.com/kubernetes/ingress-nginx/cmd/waitshutdown/new-wait-shurdown .
