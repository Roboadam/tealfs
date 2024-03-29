package ui

import (
	"fmt"
	"net/http"
	"os"
	"tealfs/pkg/mgr"
	"tealfs/pkg/model/events"
)

type Ui struct {
	manager  *mgr.Mgr
	userCmds chan events.Event
}

func NewUi(manager *mgr.Mgr, userCmds chan events.Event) Ui {
	return Ui{manager, userCmds}
}

func (ui Ui) Start() {
	ui.registerHttpHandlers()
	ui.handleRoot()
	err := http.ListenAndServe(":0", nil)
	if err != nil {
		os.Exit(1)
	}
}

func (ui Ui) registerHttpHandlers() {
	http.HandleFunc("/connect-to", func(w http.ResponseWriter, r *http.Request) {
		hostAndPort := r.FormValue("hostandport")
		ui.userCmds <- events.NewString(events.ConnectTo, hostAndPort)
		_, _ = fmt.Fprintf(w, "Connecting to: %s", hostAndPort)
	})
}

func (ui Ui) handleRoot() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `
			<!DOCTYPE html>
			<html>
			<head>
				<title>TealFS</title>
				<link rel="stylesheet" href="https://unpkg.com/mvp.css@1.12/mvp.css" /> 
				<script src="https://unpkg.com/htmx.org@1.9.2"></script>
			</head>
			<body>
			    <main>
					<h1>TealFS: ` + ui.manager.GetId().String() + `</h1>
					` + htmlMyhost("TODO") + `
					<p>Input the host and port of a node to add</p>
					<form hx-put="/connect-to">
						<label for="textbox">Host and port:</label>
						<input type="text" id="hostandport" name="hostandport">
						<input type="submit" value="Connect">
					</form>
				</main>
			</body>
			</html>
		`

		// Write the HTML content to the response writer
		_, _ = fmt.Fprintf(w, "%s", html)
	})
}

func htmlMyhost(address string) string {
	return wrapInDiv(`
			<h2>My host</h2>
			<p>Host: `+address+`</p>`,
		"myhost")
}

func wrapInDiv(html string, divId string) string {
	return `<div id="` + divId + `">` + html + `</div>`
}
