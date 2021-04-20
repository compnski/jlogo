build:
	go build -o bin/jlogo .

buildpi:
	GOARM=6 GOARCH=arm GOOS=linux go build -o bin/jlogo-pi . 

installpi: buildpi
	scp bin/jlogo-pi test.logo turtle.local:

runpi:	installpi
	ssh turtle.local 'sudo ./jlogo-pi -pi'
