SERVICE_NAME := roger-service-name
SV_DIR := /etc/sv/$(SERVICE_NAME)

all:
	@echo "nothing to build"

install:
	install -C -d -m=0755 -o root -g root "$(CURDIR)/run" "$(SV_DIR)"
	ln -s "$(SV_DIR)" "/etc/service/$(SERVICE_NAME)"

.PHONY: all install
