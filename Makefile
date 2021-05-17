# DO NOT EDIT. Generated with:
#
#    devctl@4.4.0
#

include Makefile.*.mk

.PHONY: help
## help: prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /' | sort

.PHONY: live-installation-test
## live-installation-test: runs CR validation tests on a live installation
live-installation-test:
	@echo "====> $@"
	go test -ldflags "$(LDFLAGS)" -tags="liveinstallation validate" -race ./...
