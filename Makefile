.PHONY: demo verify up down seed logs test smoke doctor terraform-check

PROJECT_DIR := infra-inventory-stream

demo verify up down seed logs test smoke doctor terraform-check:
	$(MAKE) -C $(PROJECT_DIR) $@
