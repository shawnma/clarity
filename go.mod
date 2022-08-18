module shawnma.com/clarity

go 1.18

require github.com/google/martian/v3 v3.3.2

require (
	github.com/go-sql-driver/mysql v1.6.0
	golang.org/x/net v0.0.0-20190628185345-da137c7871d7 // indirect
	golang.org/x/text v0.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/google/martian/v3 => github.com/shawnma/martian/v3 v3.3.3
