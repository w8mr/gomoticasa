package controller

import (
	"github.com/husobee/vestigo"
	"net/http"
)

func SetupStatic(router *vestigo.Router) {
	router.Get("/*", handlerWrapper(http.FileServer(http.Dir("./server/web/dist/"))))
}
