# Description of public interface
# all & install: fetch dependencies and compile everything
# test: run "go test" for all packages defined in $(GO_INSTALL)
# test-recursive: run "go test" for all packages defined in $(GO_INSTALL) and all dependent libraries
# clean: remove all packages and binaries

# ===== Configuration =====

#Define where repository lives relative to GOPATH. Make sure the following
#command works: "#go test <<PROJECT_PATH>>"
PROJECT_PATH = github.com/brsc/rlog

#Define repositories to be fetched with go get (space separated list), example:
#GO_GET_DEPEND = "launchpad.net/gocheck" "github.com/bla"
GO_GET_DEPEND = "launchpad.net/gocheck"

#Define packages to run "go install" and "go test" (space separated list), example:
#GO_INSTALL = "log/rlog" "log/httplog"
GO_INSTALL = "."
 
#Define repositories which should be cloned into GOPATH/src/github.com/brsc.
#The repositories defined here have a top-level makefile to be executed. Just
#secify the repository names here, example:
#GIT_RS_DEPEND = "tss" "fas"
GIT_RS_DEPEND =



# ===== leave this alone (usually :-)) =====

# build (no dependencies to fetch for build purposes, only testing).
.PHONY: all
all: go-compiler gopath
	@for pkg in $(GO_INSTALL) ; do \
		go install $(PROJECT_PATH)/$$pkg ; \
	done
	@echo "make all completed for $(CURDIR)"

#make install is alias for make all
.PHONY: install
install: all

#Run specs
.PHONY: test
test: go-compiler gopath go-get
	@for pkg in $(GO_INSTALL) ; do \
		go test $(PROJECT_PATH)/$$pkg ; \
	done


#Fetch using go get
.PHONY: go-get
go-get: go-compiler gopath bazaar
	@for repo in $(GO_GET_DEPEND) ; do \
		go get $$repo ; \
	done


#Execute make test recursively in all directories specified
.PHONY: test-recursive
test-recursive: go-compiler test
	@for recurse in $(GIT_RS_DEPEND) ; do \
		echo "Executing specs for $(GOPATH)/src/github.com/brsc/$$recurse"; \
		make -C $(GOPATH)/src/github.com/brsc/$$recurse test ; \
	done


#Remove all packages and binaries
.PHONY: clean
clean: gopath
	@rm -rf $(GOPATH)/pkg
	@rm -rf $(GOPATH)/bin
	@echo "clean completed"

#===== Check environment helper targets =====

#Check if the go-compiler is installed
.PHONY: go-compiler
GO_COMPILER := $(shell go version >> /dev/null; echo $$?)
go-compiler:
ifneq ($(GO_COMPILER), 0)
	$(error "Please install go compiler first")
endif

#Check if bazaar version control system is installed
.PHONY: bazaar
BAZAAR := $(shell bzr version >> /dev/null; echo $$?)
bazaar:
ifneq ($(BAZAAR), 0)
	$(error "Please install bazaar version control first")
endif

#Check if gopath is set
.PHONY: gopath
gopath:
	@if test "$(GOPATH)" = "" ; then \
		echo "Plese set GOPATH first"; \
		exit 1; \
	fi
