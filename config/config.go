package config

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	defaultAddress = ":8080"
)

type Config struct {
	BaseDomain        string
	CertFile          string
	KeyFile           string
	Address           string
	AvailabilityZones string
	Location          string
}

func Parse() (Config, error) {
	var result Config

	kingpin.Flag("tls-cert-file", "File containing the certificate for HTTPS").Required().StringVar(&result.CertFile)
	kingpin.Flag("tls-key-file", "File containing the private key for HTTPS").Required().StringVar(&result.KeyFile)
	kingpin.Flag("address", "The address to listen on").Default(defaultAddress).StringVar(&result.Address)
	kingpin.Flag("base-domain", "The base domain of the installation").Required().StringVar(&result.BaseDomain)

	kingpin.Parse()
	return result, nil
}
