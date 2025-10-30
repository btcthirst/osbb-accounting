dir := bin
app := osbb-accounting
mask := ./${dir}/${app}

build:
	go build -o ${mask}

run:
	${mask}