FROM alpine:3.14.0
WORKDIR /app
COPY azure-admission-controller /app
CMD ["/app/azure-admission-controller"]
