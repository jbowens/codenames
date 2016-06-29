package codenames

import "net/http"

const tpl = `
<!DOCTYPE html>
<html>
    <head>
        <title>Codenames</title>
        <script src="/js/lib/browser.min.js"></script>
        <script src="/js/lib/react.min.js"></script>
        <script src="/js/lib/react-dom.min.js"></script>
        <script src="/js/lib/jquery-3.0.0.min.js"></script>
        {{range .JSScripts}}
            <script type="text/babel" src="/js/{{ . }}"></script>
        {{end}}
        {{range .Stylesheets}}
            <link rel="stylesheet" type="text/css" href="/css/{{ . }}" />
        {{end}}

        <link href="https://fonts.googleapis.com/css?family=Roboto" rel="stylesheet">
    </head>
    <body>
        <div id="app"></div>
        <script type="text/babel">
            ReactDOM.render(<window.App />, document.getElementById('app'));
        </script>
    </body>
</html>
`

type templateParameters struct {
	JSLibs      []string
	JSScripts   []string
	Stylesheets []string
}

func (s *Server) handleIndex(rw http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(rw, req)
		return
	}

	err := s.tpl.Execute(rw, templateParameters{
		JSLibs:      s.jslib.RelativePaths(),
		JSScripts:   s.js.RelativePaths(),
		Stylesheets: s.css.RelativePaths(),
	})
	if err != nil {
		http.Error(rw, "error rendering", http.StatusInternalServerError)
	}
}
