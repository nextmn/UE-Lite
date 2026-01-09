prefix = /usr/local
exec_prefix = $(prefix)
bindir = $(exec_prefix)/bin
BASHCOMPLETIONSDIR = $(exec_prefix)/share/bash-completion/completions


RM = rm -f
INSTALL = install -D
MKDIRP = mkdir -p

.PHONY: install uninstall build clean default
default: build
build:
	go build
clean:
	go clean
reinstall: uninstall install
install:
	$(INSTALL) ue-lite $(DESTDIR)$(bindir)/ue-lite
	$(MKDIRP) $(DESTDIR)$(BASHCOMPLETIONSDIR)
	$(DESTDIR)$(bindir)/ue-lite completion bash > $(DESTDIR)$(BASHCOMPLETIONSDIR)/ue-lite
	@echo "================================="
	@echo ">> Now run the following command:"
	@echo "\tsource $(DESTDIR)$(BASHCOMPLETIONSDIR)/ue-lite"
	@echo "================================="
uninstall:
	$(RM) $(DESTDIR)$(bindir)/ue-lite
	$(RM) $(DESTDIR)$(BASHCOMPLETIONSDIR)/ue-lite
