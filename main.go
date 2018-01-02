package main

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
//
// Copyright © Alexander Ben Nasrallah 2018 <abn@posteo.de>

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

func handle(writer http.ResponseWriter, request *http.Request) {
	if strings.HasSuffix(request.URL.Path, ".md") {
		cmd := exec.Command("markdown", request.URL.Path[1:])
		reader, err := cmd.StdoutPipe()
		if err != nil {
			fmt.Fprintf(writer, err.Error())
			log.Print("Error connecting to command pipe: ", err)
			return
		}
		defer reader.Close()
		if err = cmd.Start(); err != nil {
			log.Print("Error running command: ", err)
			return
		}
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err = io.Copy(writer, reader); err != nil {
			log.Print("Error copying commands output to writer", err)
			return
		}
		if err = cmd.Wait(); err != nil {
			log.Print("Command not successful: ", err)
			return
		}
	} else {
		http.FileServer(http.Dir(".")).ServeHTTP(writer, request)
	}
}

// TODO usage mdserve [-p port] [markdown-cmd agrs...]
func main() {
	http.HandleFunc("/", handle)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
