# Description of public interface
#   fmt: format source code in all packages
#   install: fetch dependencies and compile everything
#   dependencies: install all dependnecies using a combination of go-get and git-clone
#   test: run "go test" for all packages
#   test-recursive: run "go test" for all packages and all dependent libraries
#   clean: remove all objects and binaries

# ===== Configuration =====

# Import path prefix for all packages (i.e. where the code lives relative to $GO_PATH/src)
PROJECT_PATH = "github.com/rightscale/rlog"

# List of all packages within PROJECT_PATH
PROJECT_PACKAGES = "." "common" "file" "stdout" "syslog"

# Dependencies to be fetched with "go get"
GO_GET_DEPEND = ""

# Dependencies to be fetched with "git clone git@github.com/..."
GIT_CLONE_DEPEND = ""

# Dependencies to be fetched with "go get" that are only used to run tests
TEST_GO_GET_DEPEND = "launchpad.net/gocheck"

# Packages that contain a binary to be installed
GO_INSTALL = "test/loggerObject" "test/modules" "test/tags"

# ===== leave this alone (usually :-)) =====

#Install binaries to $GO_PATH/bin
.PHONY: install
install: go-compiler gopath dependencies
	@for pkg in $(GO_INSTALL) ; do \
	  cd $(GOPATH) ; \
		go install $(PROJECT_PATH)/$$pkg ; \
	done
	@echo "Binaries have been installed"

#Run tests
.PHONY: test
test: go-compiler gopath dependencies test-dependencies
	@for pkg in $(GO_INSTALL) ; do \
		cd $(GOPATH) ; \
		go test $(PROJECT_PATH)/$$pkg ; \
	done
	@for pkg in $(PROJECT_PACKAGES) ; do \
		cd $(GOPATH) ; \
		go test $(PROJECT_PATH)/$$pkg ; \
	done

# Format all source code
.PHONY: fmt
fmt: go-compiler gopath
	gofmt -w=true .

# Fetch all dependencies
.PHONY: dependencies
dependencies: go-compiler gopath
	@for repo in $(GO_GET_DEPEND) ; do \
	  if [ -n "$$repo" ] ; then \
  		go get $$repo ; \
  	fi ; \
	done
	@for repo in $(GIT_CLONE_DEPEND) ; do \
	  if [ -n "$$repo" ] ; then \
      dir=`dirname $$repo` ; \
      base=`basename $$repo` ; \
      mkdir -p $(GOPATH)/src/github.com/$$dir ; \
      cd $(GOPATH)/src/github.com/$$dir ; \
      if [ ! -d $$base ] ; then \
        git clone git@github.com:$$repo ; \
      fi ; \
	  fi ; \
	done

# Fetch all dependencies
.PHONY: test-dependencies
test-dependencies: go-compiler gopath
	@for repo in $(TEST_GO_GET_DEPEND) ; do \
	  if [ -n "$$repo" ] ; then \
  		go get $$repo ; \
  	fi ; \
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

#Check if gopath is set
.PHONY: gopath
gopath:
	@if test "$(GOPATH)" = "" ; then \
		echo "Please set GOPATH first"; \
		exit 1; \
	fi
