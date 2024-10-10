default: moontpl

moontpl: **.go
	go build cmd/moontpl/moontpl.go
	
readme: moontpl
	./moontpl run README.md.lua > README.md 

dev-readme: moontpl
	./moontpl run README.md.lua -w