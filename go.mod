module github.com/supabase/jit-db-gatekeeper

go 1.24.0

toolchain go1.24.5

require github.com/lib/pq v1.10.9

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/testify v1.10.0
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/netdata/netdata/go/plugins => github.com/netdata/netdata/src/go v0.0.0-20250731052924-5b9cd0ba9812
