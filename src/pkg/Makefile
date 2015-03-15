# Copyright 2009 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# After editing the DIRS= list or adding imports to any Go files
# in any of those directories, run:
#
#	./deps.bash
#
# to rebuild the dependency information in Make.deps.

nullstring :=
space := $(nullstring) # a space at the end
ifndef GOBIN
QUOTED_HOME=$(subst $(space),\ ,$(HOME))
GOBIN=$(QUOTED_HOME)/bin
endif
QUOTED_GOBIN=$(subst $(space),\ ,$(GOBIN))

all: install

DIRS=\
	util/term\
	math\
	dbg\
	prof\
	http\
	util/misc\
	util/alarm\
	util/backoff\
	util/bytes\
	util/eye64\
	util/ioch\
	util/iofork\
	util/iopipe\
	util/netch\
	util/fixfifo\
	util/rsa64\
	util/rand\
	util/replay\
	util/signal\
	util/tube\
	util/uptime\
	util/varint\
	util/filewriter\
	needle/proto\
	needle\
	crypto\
	sys\
	monitor\
	dialer\
	routing\
	compass\
	vault\
	fe\
	core\

TEST=\
	$(filter-out $(NOTEST),$(DIRS))

BENCH=\
	$(filter-out $(NOBENCH),$(TEST))

clean.dirs: $(addsuffix .clean, $(DIRS))
install.dirs: $(addsuffix .install, $(DIRS))
nuke.dirs: $(addsuffix .nuke, $(DIRS))
test.dirs: $(addsuffix .test, $(TEST))
bench.dirs: $(addsuffix .bench, $(BENCH))

%.clean:
	+cd $* && $(QUOTED_GOBIN)/gomake clean

%.install:
	+cd $* && $(QUOTED_GOBIN)/gomake install

%.nuke:
	+cd $* && $(QUOTED_GOBIN)/gomake nuke

%.test:
	+cd $* && $(QUOTED_GOBIN)/gomake test

%.bench:
	+cd $* && $(QUOTED_GOBIN)/gomake bench

clean: clean.dirs

install: install.dirs

test:	test.dirs

bench:	bench.dirs

nuke: nuke.dirs
	rm -rf "$(GOROOT)"/pkg/*

deps:
	./deps.bash

-include Make.deps
