ayd-web-scenario: internal/lua/clua/liblua.so **/*.go
	go build

internal/lua/clua/liblua.so: internal/lua/clua/*.{c,h} internal/lua/clua/Makefile
	cd internal/lua/clua && make a
