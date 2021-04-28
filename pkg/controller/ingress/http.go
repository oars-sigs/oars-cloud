package ingress

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

func HTTPServer(port int) {
	logrus.Infof("Listen ingress http server :%d", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
