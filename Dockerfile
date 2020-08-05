FROM alpine:3.12
WORKDIR /app
COPY azure-admission-controller /app
CMD ["/app/azure-admission-controller"]
