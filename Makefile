.PHONY: all
all:
	@echo "**********************************************************"
	@echo "**                    ws build tool                    **"
	@echo "**********************************************************"

.PHONY: test
test:
	go test -timeout 30s -race -v .
