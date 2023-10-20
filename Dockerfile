FROM alpine:3.18.4
WORKDIR /app
COPY azure-admission-controller /app
CMD ["/app/azure-admission-controller"]
