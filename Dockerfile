FROM alpine:3.14.1
WORKDIR /app
COPY azure-admission-controller /app
CMD ["/app/azure-admission-controller"]
