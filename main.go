package main

// Copyright © Alexander Ben Nasrallah 2018 <abn@posteo.de>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const placeholder = "%"

var parser = "markdown"
var parserArgs = []string{}

func replaceOrAppend(buf []string, e string) []string {
	replaced := false
	for i, v := range buf {
		if v == placeholder {
			buf[i] = e
			replaced = true
		}
	}
	if !replaced {
		buf = append(buf, e)
	}
	return buf
}

func handle(writer http.ResponseWriter, request *http.Request) {
	if strings.HasSuffix(request.URL.Path, ".md") {
		args := replaceOrAppend(parserArgs, request.URL.Path[1:])
		cmd := exec.Command(parser, args...)
		reader, err := cmd.StdoutPipe()
		if err != nil {
			fmt.Fprintf(writer, err.Error())
			log.Print("Error connecting to command pipe: ", err)
			return
		}
		defer reader.Close()
		if err = cmd.Start(); err != nil {
			fmt.Fprintf(writer, err.Error())
			log.Print("Error running command: ", err)
			return
		}
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err = io.Copy(writer, reader); err != nil {
			fmt.Fprintf(writer, err.Error())
			log.Print("Error copying commands output to writer", err)
			return
		}
		if err = cmd.Wait(); err != nil {
			fmt.Fprintf(writer, err.Error())
			log.Print("Command not successful: ", err)
			return
		}
	} else {
		http.FileServer(http.Dir(".")).ServeHTTP(writer, request)
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [flags] [parser [parser args]]\n", os.Args[0])
		fmt.Println("'%' as parser arg is replaced be the requested markdown file.")
		flag.PrintDefaults()
	}
	port := flag.Uint("port", 8080, "the port to listen to")
	flag.Parse()

	if flag.NArg() > 0 {
		parser = flag.Arg(0)
		parserArgs = flag.Args()[1:]
	}

	http.HandleFunc("/", handle)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
