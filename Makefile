
default: moontpl

moontpl: **.go
	go build cmd/moontpl/moontpl.go