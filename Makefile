target = sloppysrv
GOOS=linux
GOARCH=amd64
CGO_ENABLE=0
disttar=$(target)_$(GOOS)_$(GOARCH).tar.gz

$(target):
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLE=0 go build -o $@ ./server/cmd/sloppysrv

.PHONY: clean
clean:
	rm -r $(target) $(disttar)

.POHNY: dist
dist: $(disttar)

$(disttar): $(target) install.sh
	tar zcf $@ $(target) install.sh
