package corehttp

import (
	"net"
	"net/http"

	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core"
)

func RedirectOption(path string, redirect string) ServeOption {
	return func(n *core.SubnetNode, _ net.Listener, mux *http.ServeMux) (*http.ServeMux, error) {
		cfg := n.Repo.Config()
		headers := parseHeadersFromConfig(cfg)
		handler := &redirectHandler{redirect, headers}

		if len(path) > 0 {
			mux.Handle("/"+path+"/", handler)
		} else {
			mux.Handle("/", handler)
		}
		return mux, nil
	}
}

type redirectHandler struct {
	path    string
	headers map[string][]string
}

func (i *redirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for k, v := range i.headers {
		w.Header()[http.CanonicalHeaderKey(k)] = v
	}

	http.Redirect(w, r, i.path, http.StatusFound)
}

func parseHeadersFromConfig(cfg *config.C) map[string][]string {
	headers := make(map[string][]string)

	// Lấy các headers từ cấu hình
	headerI := cfg.Get("api.http.http_headers").(map[interface{}][]interface{})

	// Duyệt qua từng cặp key-value và chuyển đổi kiểu dữ liệu
	for key, values := range headerI {
		// Chuyển đổi key từ interface{} sang string
		keyStr, ok := key.(string)
		if !ok {
			continue // Bỏ qua nếu không phải là string
		}

		// Chuyển đổi mỗi giá trị trong values từ interface{} sang string
		var strValues []string
		for _, value := range values {
			valueStr, ok := value.(string)
			if ok {
				strValues = append(strValues, valueStr)
			}
		}

		// Lưu header vào map
		headers[keyStr] = strValues
	}

	return headers
}
