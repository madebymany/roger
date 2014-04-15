install:
	cd "$(SOURCE_ROOT)/../gopath/bin" && \
	  install -C -S -m-0755 -o root -g root roger /usr/bin

.PHONY: install
