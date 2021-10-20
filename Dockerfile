FROM alpine:3.14.2
WORKDIR /app
COPY azure-admission-controller /app
CMD ["/app/azure-admission-controller"]
