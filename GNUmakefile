default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

tf-reset: ## Reset TF state.
	terraform state rm $(terraform state list)

tf-show: ## Show TF state
	terraform show

tf-list: ## List TF states
	terraform state list