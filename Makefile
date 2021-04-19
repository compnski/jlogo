build:
	go build -o bin/logoj .

buildpi:
	GOARM=6 GOARCH=arm GOOS=linux go build -o bin/logoj-pi . 

installpi: buildpi
	scp bin/logoj-pi test.logo turtle.local:

runpi:	installpi
	ssh turtle.local 'sudo ./logoj-pi -pi'
