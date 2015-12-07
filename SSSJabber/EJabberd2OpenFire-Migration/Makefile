TASK = EJabberd2OpenFire

all: $(TASK)

%: %.go
#	gccgo -g $< -o $@
	go build -o $@.gl

clean:
	rm -f $(TASK) *.gl
