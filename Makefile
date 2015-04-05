publish:
	go generate || exit 1
	./armbuild.sh || exit 1

	ssh pi "sudo service lightserverd stop" || exit 1
	scp lightserver pi: || exit 1
	ssh pi "sudo service lightserverd start" || exit 1


