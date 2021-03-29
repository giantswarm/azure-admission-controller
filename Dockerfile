FROM alpine:3.13.3
WORKDIR /app
COPY azure-admission-controller /app
CMD ["/app/azure-admission-controller"]
