// Tonika: A distributed social networking platform
// Copyright (C) 2010 Petar Maymounkov <petar@5ttt.org>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.


package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"tonika/core"
	"tonika/sys"
	//"tonika/util/signal"
)

var (
	flagAddr     = flag.String("addr", "", 
		"Address and port of " + sys.Name + " client")
	flagDbFile   = flag.String("id", "MyTonikaIdentity", 
		"Name of your identity file (contains your contacts and your private keys)")
	flagHomeDir  = flag.String("hdir", "", 
		"Your home directory, where published files reside")
	flagCacheDir = flag.String("cdir", "", 
		"Cache directory")
	flagPDir    = flag.String("pdir", "./", 
		"Directory where "+sys.Name+" is installed")
	flagFEAddr   = flag.String("fe", ":4949", 
		"Address and port of "+sys.Name+"'s Front End web proxy")
	flagFEAllow  = flag.String("fe-allow", "127.0.0.1,::1", 
		"Comma-separated list of hosts allowed to access the Front End. " +
		"Both IPv4 and IPv6 must be added if specifying IP's.")
)

func main() {
	//signal.InstallCtrlCPanic()
	fmt.Fprintf(os.Stderr, 
		"%s, 2009-10, Build %s, Released %s, by Petar Maymounkov, Homepage: %s\n", 
		sys.Name, sys.Build, sys.Released, sys.WWWURL)
	flag.Parse()

	if *flagHomeDir == "" {
		fmt.Fprintf(os.Stderr, 
			"Tonika: You forgot to specify your home directory. Use -hdir='dirhere'")
		os.Exit(1)
	}
	if *flagCacheDir == "" {
		fmt.Fprintf(os.Stderr, 
			"Tonika: You forgot to specify a cache directory. Use -cdir='dirhere'")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, 
		"What's next:\n" +
		"  * Make sure to configure your browser's proxy to %s\n" +
		"  * To use Tonika, open your browser and go to http://a."+ sys.Host +"\n", *flagFEAddr)

	go func() {
		if !askForGreenLight() {
			fmt.Fprintf(os.Stderr, "Tonika: Your version of Tonika is too old. Please update!\n")
			os.Exit(1)
		}
	}()

	cargs := &core.Args {
		Addr:     *flagAddr,
		DbFile:   *flagDbFile,
		HomeDir:  *flagHomeDir,
		CacheDir: *flagCacheDir,
		FEDir:    path.Join(*flagPDir, "fe"),
		FEAddr:   *flagFEAddr,
		FEAllow:  *flagFEAllow,
	}
	_,err := core.MakeCore(cargs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Tonika: Error starting: %s\n", err)
		os.Exit(1)
	}
	<-make(chan int)
}
