package web

import (
	"io"
	"net/http"
	"os"
	"strings"
)

type MenuData struct {
	CurrentPage string
	HasPDF      bool
}

func (srv *Server) handleMenu(w http.ResponseWriter, r *http.Request) {
	_, err := os.Stat(srv.menuFile)
	srv.render(w, "menu.html", MenuData{CurrentPage: "menu", HasPDF: err == nil})
}

func (srv *Server) handleMenuUpload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		jsonErr(w, "Request too large", 400)
		return
	}
	file, header, err := r.FormFile("pdf")
	if err != nil {
		jsonErr(w, "No file uploaded", 400)
		return
	}
	defer file.Close()
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".pdf") {
		jsonErr(w, "Only PDF files allowed", 400)
		return
	}
	dst, err := os.Create(srv.menuFile)
	if err != nil {
		jsonErr(w, "Failed to save file", 500)
		return
	}
	defer dst.Close()
	io.Copy(dst, file)
	jsonOK(w, map[string]bool{"success": true})
}
