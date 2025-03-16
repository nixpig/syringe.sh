FROM golang:1.24.0

ARG VERSION

RUN wget -qO- https://github.com/nixpig/syringe.sh/releases/download/${VERSION}/syringe.sh_syringeserver_${VERSION}_linux_amd64.tar.gz | tar -xzvf - -C /go/bin

EXPOSE 22

CMD [ "syringeserver" ]
