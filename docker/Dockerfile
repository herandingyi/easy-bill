FROM golang AS builder
WORKDIR /app

#RUN apk add --no-cache git
RUN git clone --branch main https://github.com/herandingyi/easy-bill.git  \
&& cd easy-bill \
&& go mod tidy \
&& go build -o /app/easy-bill .

FROM golang
WORKDIR /app
COPY --from=builder /app/easy-bill/easy-bill /app/easy-bill
COPY --from=builder /app/easy-bill/simsun.ttc /app/simsun.ttc
RUN chmod +x /app/easy-bill

CMD ["/app/easy-bill"]

#docker build -t herandingyi/easy-bill:latest . --no-cache
#docker push herandingyi/easy-bill:latest