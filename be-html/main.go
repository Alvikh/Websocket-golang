package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

//https://gin-gonic.com/docs/examples/html-rendering/
//https://stackoverflow.com/questions/57354389/how-to-render-static-files-within-gin-router

func main() {
	router := gin.Default()

	// Memuat semua file HTML yang ada di folder "templates"
	router.LoadHTMLGlob("./templates/*")

	// Mendefinisikan route GET untuk "/index"
	router.GET("/index", func(c *gin.Context) {
		// Menampilkan file "index.tmpl" dengan HTTP status 200 OK, tanpa data tambahan
		c.HTML(http.StatusOK, "index.tmpl", nil)
	})

	// Menyajikan file statis (seperti CSS, JavaScript, gambar) dari folder "static"
	router.Static("/static", "./static")

	// Menjalankan server pada port 8080
	router.Run(":8080")
}
