default: moontpl

moontpl: **.go lua/*.lua
	go build cmd/moontpl/moontpl.go
	go install cmd/moontpl/moontpl.go &
	
readme: moontpl
	./moontpl run README.md.lua > README.md 

dev-readme: moontpl
	./moontpl run README.md.lua -w