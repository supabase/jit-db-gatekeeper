module github.com/supabase/jit-db-gatekeeper

go 1.24.0

toolchain go1.24.5

require github.com/lib/pq v1.10.9

require github.com/netdata/netdata/go/plugins v0.0.0-00010101000000-000000000000

replace github.com/netdata/netdata/go/plugins => github.com/netdata/netdata/src/go v0.0.0-20250731052924-5b9cd0ba9812
